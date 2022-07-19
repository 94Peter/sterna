package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/94peter/sterna/db"
	"github.com/94peter/sterna/model/cache"
	"github.com/stretchr/testify/assert"
)

func GetTestRedisClt() (db.RedisClient, error) {
	conf := db.RedisConf{
		Host: "127.0.0.1:6379",
		DB:   0,
	}
	return conf.NewRedisClient(context.TODO())
}

func Test_RedisSetGetObj(t *testing.T) {
	type test struct {
		A string
		B int
	}
	clt, err := GetTestRedisClt()
	assert.Nil(t, err)
	rc := cache.NewRedisCache(clt)
	obj := test{
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
