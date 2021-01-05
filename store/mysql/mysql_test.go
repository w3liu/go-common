package mysql

import (
	"testing"
	"time"
)

type TestUser struct {
	Id        int64     `xorm:"not null pk autoincr BIGINT(20)"`
	UserName  string    `xorm:"not null UNIQUE VARCHAR(50) comment('用户名')"`
	Password  string    `xorm:"not null VARCHAR(100) comment('密码')"`
	CreatedAt time.Time `xorm:"created"`
	UpdatedAt time.Time `xorm:"updated"`
}

func initStore() *Store {
	cfg := Conf{
		HostPort: "127.0.0.1:3306",
		Username: "root",
		DBName:   "test",
		Password: "111111",
		MaxConns: 100,
		MaxIdle:  10,
	}
	store := NewStore(&cfg)
	return store
}

func TestInsert(t *testing.T) {
	store := initStore()
	_, err := store.Insert(&TestUser{
		UserName: "test1",
		Password: "111111",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestSync2(t *testing.T) {
	store := initStore()
	err := store.Sync2(new(TestUser))
	if err != nil {
		t.Fatal(err)
	}
}
