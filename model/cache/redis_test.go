package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/94peter/sterna/dao"
	"github.com/94peter/sterna/db"
	"github.com/94peter/sterna/model/cache"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

type test struct {
	A string
	B int

	common dao.ComCacheObj
}

func (t *test) Encode() ([]byte, error) {
	return t.common.Encode(t)
}

func (t *test) Decode(data []byte) error {
	return t.common.Decode(data, t)
}

func (t *test) DecodePipe() error {
	return t.common.DecodePipe(t)
}

func (s *test) SetStringCmd(sc *redis.StringCmd) {
	s.common.SetStringCmd(sc)
}

func (s *test) GetStringCmd() (*redis.StringCmd, error) {
	return s.common.GetStringCmd()
}

func (s *test) GetError() error {
	return s.common.GetError()
}

func (s *test) HasError() bool {
	return s.common.HasError()
}

func GetTestRedisClt() (db.RedisClient, error) {
	conf := db.RedisConf{
		Host: "127.0.0.1:6379",
		DB:   0,
	}
	return conf.NewRedisClient(context.TODO())
}

func Test_RedisSetGetObj(t *testing.T) {

	clt, err := GetTestRedisClt()
	assert.Nil(t, err)
	rc := cache.NewRedisCache(clt)
	obj := &test{
		A: "aaa",
		B: 2,
	}
	err = rc.SaveObj("test", obj, time.Hour)
	assert.Nil(t, err)
	newObj := &test{}
	err = rc.GetObj("test", newObj)
	assert.Nil(t, err)
	assert.Equal(t, obj.A, newObj.A)
}

func Test_RedisPipeGetObjs(t *testing.T) {
	clt, err := GetTestRedisClt()
	assert.Nil(t, err)
	rc := cache.NewRedisCache(clt)

	err = rc.SaveObj("key1", &test{
		A: "aaa",
		B: 2,
	}, time.Hour)
	assert.Nil(t, err)
	err = rc.SaveObj("key2", &test{
		A: "bbb",
		B: 2,
	}, time.Hour)
	assert.Nil(t, err)
	err = rc.SaveObj("key3", &test{
		A: "ccc",
		B: 2,
	}, time.Hour)
	assert.Nil(t, err)

	result, err := rc.GetObjs([]string{"key1", "key2", "key4"}, &test{})
	assert.Nil(t, err)
	assert.Equal(t, 3, len(result))

	tt1 := result[0].(*test)
	assert.Equal(t, "aaa", tt1.A)
	tt2 := result[1].(*test)
	assert.Equal(t, "bbb", tt2.A)
	tt3 := result[2].(*test)
	assert.True(t, tt3.HasError())

}
