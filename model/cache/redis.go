package cache

import (
	"errors"
	"reflect"
	"time"

	"github.com/94peter/sterna/dao"
	"github.com/94peter/sterna/db"
)

type Cache interface {
	SaveObj(i dao.CacheObj, exp time.Duration) error
	GetObj(key string, i dao.CacheObj) error
	GetObjs(keys []string, d dao.CacheObj) (objs []dao.CacheObj, err error)
}

func NewRedisCache(clt db.RedisClient) Cache {
	return &redisCache{
		RedisClient: clt,
	}
}

type redisCache struct {
	db.RedisClient
}

func (c *redisCache) SaveObj(i dao.CacheObj, exp time.Duration) error {

	b, err := i.Encode()
	if err != nil {
		return err
	}
	_, err = c.Set(i.GetKey(), b, exp)
	return err
}

func (c *redisCache) GetObj(key string, i dao.CacheObj) error {
	if reflect.ValueOf(i).Type().Kind() != reflect.Ptr {
		return errors.New("must be pointer")
	}
	data, err := c.Get(key)
	if err != nil {
		return err
	}
	err = i.Decode(data)
	return err
}

func (c *redisCache) GetObjs(keys []string, d dao.CacheObj) (objs []dao.CacheObj, err error) {
	var sliceList []dao.CacheObj

	objType := reflect.TypeOf(d)
	if objType.Kind() == reflect.Ptr {
		objType = objType.Elem()
	}

	var newValue reflect.Value
	var newDoc dao.CacheObj
	pipe := c.NewPiple()
	for _, k := range keys {
		newValue = reflect.New(objType)
		newDoc = newValue.Interface().(dao.CacheObj)
		newDoc.SetStringCmd(pipe.Get(k))
		sliceList = append(sliceList, newDoc)
	}
	pipe.Exec()
	for _, s := range sliceList {
		if !s.HasError() {
			s.DecodePipe()
		}
	}
	return sliceList, nil
}
