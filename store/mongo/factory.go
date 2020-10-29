package mongo

import (
	"errors"
	"go.mongodb.org/mongo-driver/mongo"
	"sync"
)

type MgoConfStore struct {
	conf      *MgoConf
	secondary *MgoStore
	primary   *MgoStore
}

type MgoFactory struct {
	sync.Mutex
	stores map[string]*MgoConfStore
}

func NewMgoFactory() *MgoFactory {
	return &MgoFactory{
		stores: make(map[string]*MgoConfStore),
	}
}

// 添加配置
// key：数据库关键字
// conf：配置
func (f *MgoFactory) AddConf(key string, conf *MgoConf) error {
	f.Lock()
	defer f.Unlock()
	if _, ok := f.stores[key]; ok {
		return errors.New("the config is existed")
	}
	f.stores[key] = &MgoConfStore{conf: conf}
	return nil
}

// 获取store
// key：数据库关键字
// secondary：读模式是否为secondary模式
func (f *MgoFactory) GetStore(key string, secondary bool) *MgoStore {
	f.Lock()
	defer f.Unlock()
	if s, ok := f.stores[key]; ok {
		if secondary {
			if store, err := getSecondary(s); err != nil {
				panic(err)
			} else {
				return store
			}
		} else {
			if store, err := getPrimary(s); err != nil {
				panic(err)
			} else {
				return store
			}
		}
	}
	panic("MgoStore not found")
}

// 获取secondary store
func getSecondary(s *MgoConfStore) (*MgoStore, error) {
	if s.secondary != nil {
		return s.secondary, nil
	} else {
		if store, err := createSecondary(s.conf); err != nil {
			return nil, err
		} else {
			s.secondary = store
			return s.secondary, nil
		}
	}
}

// 获取primary store
func getPrimary(s *MgoConfStore) (*MgoStore, error) {
	if s.primary != nil {
		return s.primary, nil
	} else {
		if store, err := createPrimary(s.conf); err != nil {
			return nil, err
		} else {
			s.primary = store
			return s.primary, nil
		}
	}
}

// 添加secondary模式的store
// conf：数据库配置文件
func createSecondary(conf *MgoConf) (*MgoStore, error) {
	var dbCli *mongo.Client
	if cli, err := NewQueryClient(conf); err != nil {
		return nil, err
	} else {
		dbCli = cli
	}
	dbStore := NewStore(dbCli, conf.DB)
	return dbStore, nil
}

// 添加primary模式的store
// conf：数据库配置文件
func createPrimary(conf *MgoConf) (*MgoStore, error) {
	var dbCli *mongo.Client
	if cli, err := NewClient(conf); err != nil {
		return nil, err
	} else {
		dbCli = cli
	}
	dbStore := NewStore(dbCli, conf.DB)
	return dbStore, nil
}
