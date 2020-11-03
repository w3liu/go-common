package mongo

import (
	"context"
	"errors"
	"fmt"
	"github.com/w3liu/go-common/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"strings"
	"time"
)

var logger = log.L()

type MgoConf struct {
	User        string
	Password    string
	DataSource  []string
	AuthDB      string
	DB          string
	ReplicaSet  string
	MaxPoolSize uint64
}

type MgoStore struct {
	cli     *mongo.Client
	db      *mongo.Database
	timeOut time.Duration
}

func NewClient(cfg *MgoConf) (*mongo.Client, error) {
	return newClient(cfg, "")
}

func NewQueryClient(cfg *MgoConf) (*mongo.Client, error) {
	return newClient(cfg, "secondaryPreferred")
}

func newClient(cfg *MgoConf, readPreference string) (*mongo.Client, error) {
	uri := fmt.Sprintf("mongodb://%s:%s@%s/%s?", cfg.User, cfg.Password, strings.Join(cfg.DataSource, ","), cfg.AuthDB)
	if cfg.ReplicaSet != "" {
		uri = fmt.Sprintf("%sreplicaSet=%s&", uri, cfg.ReplicaSet)
	}
	if readPreference != "" {
		uri = fmt.Sprintf("%sreadPreference=%s&", uri, readPreference)
	}
	uri = strings.TrimRight(strings.TrimRight(uri, "?"), "&")
	client, err := mongo.NewClient(options.Client().ApplyURI(uri).SetMaxPoolSize(cfg.MaxPoolSize))
	if err != nil {
		return nil, err
	}
	err = client.Connect(context.Background())
	if err != nil {
		return nil, err
	}
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = client.Ping(ctxTimeout, nil)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func NewStore(cli *mongo.Client, db string) *MgoStore {
	return &MgoStore{
		cli:     cli,
		db:      cli.Database(db),
		timeOut: time.Second * 5,
	}
}

func (s *MgoStore) DB() *mongo.Database {
	return s.db
}

func (s *MgoStore) Cli() *mongo.Client {
	return s.cli
}

func (s *MgoStore) CloseCursor(ctx context.Context, cursor *mongo.Cursor) {
	err := cursor.Close(ctx)
	if err != nil {
		logger.Error("CloseCursor", zap.Error(err))
	}
}

func (s *MgoStore) CreateIndexMany(indexes []Index) error {
	indexModels := make(map[string][]mongo.IndexModel)
	for _, index := range indexes {
		if err := index.Validate(); err != nil {
			return err
		}
		model := mongo.IndexModel{
			Keys: index.Keys,
		}
		opt := options.Index()
		if index.Name != "" {
			opt.SetName(index.Name)
		}
		opt.SetUnique(index.Unique)
		opt.SetBackground(index.Background)

		if index.ExpireAfterSeconds > 0 {
			opt.SetExpireAfterSeconds(index.ExpireAfterSeconds)
		}

		model.Options = opt

		v, ok := indexModels[index.Collection]
		if ok {
			indexModels[index.Collection] = append(v, model)
		} else {
			indexModels[index.Collection] = []mongo.IndexModel{model}
		}
	}

	for collection, index := range indexModels {
		_, err := s.db.Collection(collection).Indexes().CreateMany(context.Background(), index)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *MgoStore) FindOne(ctx context.Context, o *OneFinder) (bool, error) {
	if o == nil || o.col == nil {
		return false, errors.New("oneFinder is invalid")
	}
	result := s.db.Collection(o.col.Name()).FindOne(ctx, o.filter, o.options...)
	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, result.Err()
	}
	err := result.Decode(o.col)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *MgoStore) FindMany(ctx context.Context, o *Finder) error {
	if o == nil || o.col == nil || o.records == nil {
		return errors.New("finder is invalid")
	}
	cursor, err := s.db.Collection(o.col.Name()).Find(ctx, o.filter, o.options...)
	defer s.CloseCursor(ctx, cursor)
	if err != nil {
		return err
	}
	defer s.CloseCursor(ctx, cursor)
	err = cursor.All(ctx, o.records)
	if err != nil {
		return err
	}
	return nil
}

func (s *MgoStore) InsertOne(ctx context.Context, col Collection) error {
	if col == nil {
		return errors.New("collection is invalid")
	}
	result, err := s.db.Collection(col.Name()).InsertOne(ctx, col)
	if err != nil {
		return err
	}
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		col.SetId(oid)
	}
	return err
}

func (s *MgoStore) DeleteOne(ctx context.Context, col Collection) (int64, error) {
	if col == nil {
		return 0, errors.New("collection is invalid")
	}
	filter := bson.D{
		{"_id", col.GetId()},
	}
	result, err := s.db.Collection(col.Name()).DeleteOne(ctx, filter)
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, err
}

func (s *MgoStore) UpdateOne(ctx context.Context, o *Updater) (int64, error) {
	if o == nil || o.col == nil || len(o.filter) == 0 || len(o.update) == 0 {
		return 0, errors.New("updater is invalid")
	}
	update := bson.D{
		{"$set", o.update},
	}
	result, err := s.db.Collection(o.col.Name()).UpdateOne(ctx, o.filter, update, o.options...)
	if err != nil {
		return 0, err
	}
	return result.ModifiedCount, nil
}

func (s *MgoStore) InsertMany(ctx context.Context, cols []interface{}) error {
	if len(cols) == 0 {
		return errors.New("cols is invalid")
	}

	var name string
	if v, ok := cols[0].(Collection); ok {
		name = v.Name()
	} else {
		return errors.New("cols not implement collection interface")
	}

	result, err := s.db.Collection(name).InsertMany(ctx, cols)
	if err != nil {
		return err
	}

	for i, _ := range result.InsertedIDs {
		if id, ok := result.InsertedIDs[i].(primitive.ObjectID); ok {
			if v, ok := cols[i].(Collection); ok {
				v.SetId(id)
			}
		}
	}

	return err
}

func (s *MgoStore) DeleteMany(ctx context.Context, o *Deleter) (int64, error) {
	if o == nil || o.col == nil {
		return 0, errors.New("deleter is invalid")
	}
	result, err := s.db.Collection(o.col.Name()).DeleteMany(ctx, o.filter, o.options...)
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, err
}

func (s *MgoStore) UpdateMany(ctx context.Context, o *Updater) (int64, error) {
	if o == nil || o.col == nil || len(o.update) == 0 {
		return 0, errors.New("updater is invalid")
	}
	update := bson.D{
		{"$set", o.update},
	}
	result, err := s.db.Collection(o.col.Name()).UpdateMany(ctx, o.filter, update, o.options...)
	if err != nil {
		return 0, err
	}
	return result.ModifiedCount, nil
}

func (s *MgoStore) Aggregate(ctx context.Context, o *Aggregator) error {
	if o == nil || o.col == nil || len(o.pipeline) == 0 || o.records == nil {
		return errors.New("aggregator is invalid")
	}
	cursor, err := s.db.Collection(o.col.Name()).Aggregate(ctx, o.pipeline, o.options...)
	if err != nil {
		return err
	}
	defer s.CloseCursor(ctx, cursor)
	err = cursor.All(ctx, o.records)
	if err != nil {
		return err
	}
	return nil
}

func (s *MgoStore) CountDocuments(ctx context.Context, o *Counter) (int64, error) {
	if o == nil || o.col == nil {
		return 0, errors.New("counter is invalid")
	}
	cnt, err := s.db.Collection(o.col.Name()).CountDocuments(ctx, o.filter)
	if err != nil {
		return 0, err
	}
	return cnt, nil
}

func (s *MgoStore) CountEstimateDocuments(ctx context.Context, o *EstimateCounter) (int64, error) {
	if o == nil || o.col == nil {
		return 0, errors.New("estimateCounter is invalid")
	}
	cnt, err := s.db.Collection(o.col.Name()).EstimatedDocumentCount(ctx, o.options...)
	if err != nil {
		return 0, err
	}
	return cnt, nil
}
