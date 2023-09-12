package db

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisDI interface {
	NewRedisClientDB(ctx context.Context, db int) (RedisClient, error)
	GetDB(dbname string) int
}

type RedisConf struct {
	Host  string         `yaml:"host"`
	Pwd   string         `yaml:"pass"`
	DbMap map[string]int `yaml:"dbMap"`
}

func (rc *RedisConf) GetDB(dbname string) int {
	return rc.DbMap[dbname]
}

func (rc *RedisConf) NewRedisClientDB(ctx context.Context, db int) (RedisClient, error) {
	//const connTimeout = time.Second * 5
	var r RedisClient
	r = &redisV8CltImpl{
		clt: redis.NewClient(&redis.Options{
			Addr: rc.Host,
			// Password:     r.Pwd, // no password set
			DB: db, // use default DB
			// DialTimeout:  connTimeout,
			// ReadTimeout:  connTimeout,
			// WriteTimeout: connTimeout,
		}),
		ctx: ctx,
		db:  db,
	}

	if r.Ping() != "PONG" {
		return nil, errors.New("redis connect error")
	}
	return r, nil
}

type RedisClient interface {
	Close() error
	Ping() string
	CountKeys() (int, error)
	Get(k string) ([]byte, error)
	Set(k string, v interface{}, exp time.Duration) (string, error)
	Del(k string) (int64, error)
	LPush(k string, v interface{}) (int64, error)
	RPop(k string) ([]byte, error)
	HMGet(key string, field ...string) []interface{}
	HMSet(key string, values ...interface{}) error
	Exists(key string) bool
	Expired(key string, d time.Duration) (bool, error)
	NewPiple() CachePipel
	Keys(pattern string) ([]string, error)
}

type CachePipel interface {
	Get(key string) *redis.StringCmd
	Exec() ([]redis.Cmder, error)
}

type redisV8CltImpl struct {
	clt *redis.Client
	ctx context.Context
	db  int
}

func (rci *redisV8CltImpl) Keys(pattern string) ([]string, error) {
	return rci.clt.Keys(rci.ctx, pattern).Result()
}

func (rci *redisV8CltImpl) Get(key string) ([]byte, error) {
	return rci.clt.Get(rci.ctx, key).Bytes()
}

func (rci *redisV8CltImpl) Expired(key string, d time.Duration) (bool, error) {
	return rci.clt.Expire(rci.ctx, key, d).Result()
}

func (rci *redisV8CltImpl) CountKeys() (int, error) {
	r := rci.clt.Info(rci.ctx, "keyspace").String()
	k := fmt.Sprintf("db%d:keys=", rci.db)
	i := strings.Index(r, k)
	l := len(r)
	var count []byte
	for i = i + len(k); i < l; i++ {
		if r[i] == ',' {
			break
		}
		count = append(count, r[i])
	}
	return strconv.Atoi(string(count))
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

func (rci *redisV8CltImpl) HMSet(key string, values ...interface{}) error {
	return rci.clt.HMSet(rci.ctx, key, values...).Err()
}

func (rci *redisV8CltImpl) NewPiple() CachePipel {
	return &myPipel{
		redisPiple: rci.clt.Pipeline(),
		ctx:        rci.ctx,
	}
}

type myPipel struct {
	ctx        context.Context
	redisPiple redis.Pipeliner
}

func (p *myPipel) Get(key string) *redis.StringCmd {
	return p.redisPiple.Get(p.ctx, key)
}

func (p *myPipel) Exec() ([]redis.Cmder, error) {
	return p.redisPiple.Exec(p.ctx)
}
