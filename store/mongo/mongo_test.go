package mongo

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"testing"
	"time"
)

var store *MgoStore
var defaultCollection *demoCollection

func init() {
	conf := &MgoConf{
		User:        "root",
		Password:    "111111",
		DataSource:  []string{"127.0.0.1:27017"},
		DB:          "test",
		AuthDB:      "admin",
		ReplicaSet:  "",
		MaxPoolSize: 100,
	}
	dbCli, err := NewClient(conf)
	if err != nil {
		panic(err)
	}
	dbStore := NewStore(dbCli, conf.DB)
	store = dbStore
	defaultCollection = &demoCollection{
		Title:     "什么是Lambda架构",
		Author:    "数据社",
		Content:   "Lambda架构背后的需求是由于MR架构的延迟问题。MR虽然实现了分布式、可扩展数据处理系统的目的，但是在处理数据时延迟比较严重。",
		Status:    1,
		CreatedAt: time.Now(),
	}
}

type demoCollection struct {
	Id        primitive.ObjectID `bson:"_id,omitempty"`
	Title     string             `bson:"title"`
	Author    string             `bson:"author"`
	Content   string             `json:"content"`
	Status    int32              `bson:"status"`
	CreatedAt time.Time          `bson:"at"`
}

func (c *demoCollection) Name() string {
	return "demo_collection"
}

func (c *demoCollection) GetId() primitive.ObjectID {
	return c.Id
}

func (c *demoCollection) SetId(id primitive.ObjectID) {
	c.Id = id
}

func TestFindOne(t *testing.T) {
	col := &demoCollection{}
	filter := bson.D{
		{"title", "什么是Lambda架构"},
	}
	opt := options.FindOne().SetSort(bson.D{{"_id", -1}})
	finder := NewOneFinder(col).Where(filter).Options(opt)

	b, err := store.FindOne(context.TODO(), finder)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("b", b)
	t.Logf("%#v", col)
}

func TestFindMany(t *testing.T) {
	col := &demoCollection{}
	filter := bson.D{
		//{"title", "什么是Lambda架构"},
	}
	opt := options.Find().SetSort(bson.D{{"_id", -1}})
	records := make([]*demoCollection, 0)
	finder := NewFinder(col).Where(filter).Options(opt).Records(&records)

	err := store.FindMany(context.TODO(), finder)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range records {

		t.Log(item)
	}
}

func TestInsertOne(t *testing.T) {
	err := store.InsertOne(context.TODO(), defaultCollection)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("id", defaultCollection.Id.Hex())
}

func TestDeleteOne(t *testing.T) {
	id, _ := primitive.ObjectIDFromHex("5f6183a9ed076ced7eacec3a")
	col := &demoCollection{
		Id: id,
	}
	cnt, err := store.DeleteOne(context.TODO(), col)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("cnt", cnt)
}

func TestUpdateOne(t *testing.T) {
	id, _ := primitive.ObjectIDFromHex("5f618414c978e349ced0c81f")
	col := &demoCollection{
		Id: id,
	}
	filter := bson.D{
		{"_id", id},
	}
	update := bson.D{
		{"title", "什么是Lambda架构?"},
	}
	updater := NewUpdater(col).Where(filter).Update(update)
	cnt, err := store.UpdateOne(context.TODO(), updater)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("cnt", cnt)
}

func TestInsertMany(t *testing.T) {
	cols := make([]*demoCollection, 0)
	title := defaultCollection.Title
	for i := 0; i < 10; i++ {
		item := &demoCollection{
			Title:     fmt.Sprintf("%s_%d", title, i),
			Author:    defaultCollection.Author,
			Content:   defaultCollection.Content,
			Status:    1,
			CreatedAt: time.Now(),
		}
		cols = append(cols, item)
	}

	docs := make([]interface{}, 0)
	for i := 0; i < len(cols); i++ {
		docs = append(docs, cols[i])
	}

	err := store.InsertMany(context.TODO(), docs)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < len(cols); i++ {
		t.Log("id", cols[i].GetId())
		t.Log("title", cols[i].Title)
	}
}

func TestDeleteMany(t *testing.T) {
	//filter := bson.D{
	//	{"title", "什么是Lambda架构"},
	//}
	deleter := NewDeleter(defaultCollection).Where(nil)
	cnt, err := store.DeleteMany(context.TODO(), deleter)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("cnt", cnt)
}

func TestUpdateMany(t *testing.T) {
	filter := bson.D{
		{"author", "数据社"},
	}
	update := bson.D{
		{"title", "什么是Lambda架构"},
	}
	updater := NewUpdater(defaultCollection).Where(filter).Update(update)
	cnt, err := store.UpdateMany(context.TODO(), updater)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("cnt", cnt)
}

func TestAggregate(t *testing.T) {

	var records []*struct {
		Total int `bson:"total"`
		Count int `bson:"count"`
	}

	match := bson.D{
		{"$match", bson.D{
			{"author", "数据社1"},
		}},
	}

	group := bson.D{
		{"$group", bson.D{
			{"_id", nil},
			{"total", bson.M{"$sum": "$status"}},
			{"count", bson.M{"$sum": 1}},
		}},
	}

	aggregator := NewAggregator(defaultCollection).Stage(match).Stage(group).Records(&records)

	err := store.Aggregate(context.TODO(), aggregator)
	if err != nil {
		t.Fatal(err)
	}

	if len(records) > 0 {
		t.Log(records[0])
	}
}

func TestCountDocuments(t *testing.T) {
	//filter := bson.D{
	//	{"author", "数据社1"},
	//}
	counter := NewCounter(defaultCollection).Where(nil)
	cnt, err := store.CountDocuments(context.TODO(), counter)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("cnt", cnt)
}

func TestCountEstimateDocuments(t *testing.T) {
	counter := NewEstimateCounter(defaultCollection)
	cnt, err := store.CountEstimateDocuments(context.TODO(), counter)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("cnt", cnt)
}
