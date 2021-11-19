package db

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/94peter/sterna/util"
	"github.com/go-redis/redis/v8"
)

const (
	CtxRedisKey = util.CtxKey("ctxRedisKey")
)

type RedisDI interface {
	NewRedisClient(ctx context.Context) (RedisClient, error)
}

func GetRedisByReq(req *http.Request) RedisClient {
	ctx := req.Context()
	cltInter := ctx.Value(CtxRedisKey)

	if dbclt, ok := cltInter.(RedisClient); ok {
		return dbclt
	}
	return nil
}

type RedisConf struct {
	Host string `yaml:"host"`
	Pwd  string `yaml:"pass"`
	DB   int    `yaml:"db"`
}

func (rc *RedisConf) NewRedisClient(ctx context.Context) (RedisClient, error) {
	//const connTimeout = time.Second * 5
	var r RedisClient
	r = &redisV8CltImpl{
		clt: redis.NewClient(&redis.Options{
			Addr: rc.Host,
			// Password:     r.Pwd, // no password set
			DB: rc.DB, // use default DB
			// DialTimeout:  connTimeout,
			// ReadTimeout:  connTimeout,
			// WriteTimeout: connTimeout,
		}),
		ctx: ctx,
	}

	if r.Ping() != "PONG" {
		return nil, errors.New("redis connect error")
	}
	return r, nil
}

type RedisClient interface {
	Close() error
	Ping() string
	Set(k string, v interface{}, exp time.Duration) (string, error)
	Del(k string) (int64, error)
	LPush(k string, v interface{}) (int64, error)
	RPop(k string) ([]byte, error)
	HMGet(key string, field ...string) []interface{}
	Exists(key string) bool
}

type redisV8CltImpl struct {
	clt *redis.Client
	ctx context.Context
}

func (rci *redisV8CltImpl) Close() error {
	return rci.clt.Close()
}

func (rci *redisV8CltImpl) Ping() string {
	return rci.clt.Ping(rci.ctx).Val()
}

func (rci *redisV8CltImpl) Set(k string, v interface{}, exp time.Duration) (string, error) {
	return rci.clt.Set(rci.ctx, k, v, exp).Result()
}

func (rci *redisV8CltImpl) Del(k string) (int64, error) {
	return rci.clt.Del(rci.ctx, k).Result()
}

func (rci *redisV8CltImpl) LPush(k string, v interface{}) (int64, error) {
	return rci.clt.LPush(rci.ctx, k, v).Result()
}

func (rci *redisV8CltImpl) RPop(k string) ([]byte, error) {
	return rci.clt.RPop(rci.ctx, k).Bytes()
}

func (rci *redisV8CltImpl) Exists(key string) bool {
	return rci.clt.Exists(rci.ctx, key).Val() == 1
}

func (rci *redisV8CltImpl) HMGet(key string, field ...string) []interface{} {
	return rci.clt.HMGet(rci.ctx, key, field...).Val()
}