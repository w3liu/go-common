package mysql

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/w3liu/go-common/log"
	"go.uber.org/zap"
	"time"
	"xorm.io/xorm"
)

type Conf struct {
	HostPort string
	Username string
	DBName   string
	Password string
	MaxConns int
	MaxIdle  int
	ShowSQL  bool
}

type Store struct {
	*xorm.Engine
}

type Query struct {
	*xorm.Engine
}

type Transaction interface {
	GetSession() *xorm.Session
}

func NewStore(cfg *Conf) *Store {
	dial := fmt.Sprintf("%v:%v@tcp(%v)/%v?charset=utf8&loc=%v", cfg.Username,
		cfg.Password, cfg.HostPort, cfg.DBName, "Asia%2fShanghai")
	engine, err := xorm.NewEngine("mysql", dial)
	if err != nil {
		panic(err)
	}
	engine.SetMaxOpenConns(cfg.MaxConns)
	engine.SetMaxIdleConns(cfg.MaxIdle)
	engine.ShowSQL(cfg.ShowSQL)
	engine.SetConnMaxLifetime(time.Hour * 2)

	return &Store{engine}
}

func NewQuery(cfg *Conf) *Query {
	dial := fmt.Sprintf("%v:%v@tcp(%v)/%v?charset=utf8&loc=%v", cfg.Username,
		cfg.Password, cfg.HostPort, cfg.DBName, "Asia%2fShanghai")
	engine, err := xorm.NewEngine("mysql", dial)
	if err != nil {
		panic(err)
	}
	engine.SetMaxOpenConns(cfg.MaxConns)
	engine.SetMaxIdleConns(cfg.MaxIdle)
	engine.ShowSQL(cfg.ShowSQL)
	engine.SetConnMaxLifetime(time.Hour * 2)

	return &Query{engine}
}

func Execute(trans Transaction, fn func() error) error {
	if err := trans.GetSession().Begin(); err != nil {
		return err
	}
	defer trans.GetSession().Close()
	e := fn()
	if e != nil {
		log.Error("trans error", zap.Error(e))
		_ = trans.GetSession().Rollback()
		return e
	}
	err := trans.GetSession().Commit()
	return err
}
