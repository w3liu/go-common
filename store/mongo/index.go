package mongo

import (
	"errors"
	"go.mongodb.org/mongo-driver/bson"
)

type Index struct {
	Collection         string
	Name               string // 指定索引名称
	Keys               bson.D
	Unique             bool  // 唯一索引
	Background         bool  // 非阻塞创建索引
	ExpireAfterSeconds int32 // 多少秒后过期
}

func (i Index) Validate() error {
	if i.Collection == "" {
		return errors.New("collection required")
	}
	if len(i.Keys) == 0 {
		return errors.New("keys required")
	}
	return nil
}
