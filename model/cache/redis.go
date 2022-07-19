package cache

import (
	"bytes"
	"encoding/gob"
	"errors"
	"reflect"
	"time"

	"github.com/94peter/sterna/db"
)

type Cache interface {
	SaveObj(key string, i interface{}, exp time.Duration) error
	GetObj(key string, i interface{}) error
}

func NewRedisCache(clt db.RedisClient) Cache {
	return &redisCache{
		RedisClient: clt,
	}
}

type redisCache struct {
	db.RedisClient
}

func (c *redisCache) SaveObj(key string, i interface{}, exp time.Duration) error {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(i)
	if err != nil {
		return err
	}
	_, err = c.Set(key, buf.Bytes(), exp)
	return err
}

func (c *redisCache) GetObj(key string, i interface{}) error {
	if reflect.ValueOf(i).Type().Kind() != reflect.Ptr {
		return errors.New("must be pointer")
	}
	data, err := c.Get(key)
	if err != nil {
		return err
	}
	dec := gob.NewDecoder(bytes.NewReader(data))
	err = dec.Decode(i)
	return err
}
