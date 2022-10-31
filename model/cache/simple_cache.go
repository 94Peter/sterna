package cache

import (
	"time"

	"github.com/94peter/sterna/db"
)

type SimpleCacheData interface {
	GetKey() string
	GetData() ([]byte, error)
	Expired() time.Duration
}

type SimpleCache interface {
	Save() (string, error)
	Get() ([]byte, error)
	SetData(d SimpleCacheData)
}

func NewCache(data SimpleCacheData, db db.RedisClient) SimpleCache {
	return &cacheService{
		cache: db,
		data:  data,
	}
}

type cacheService struct {
	cache db.RedisClient
	data  SimpleCacheData
}

func (c *cacheService) Save() (string, error) {
	d, err := c.data.GetData()
	if err != nil {
		return "", err
	}
	return c.cache.Set(c.data.GetKey(), d, c.data.Expired())
}
func (c *cacheService) Get() ([]byte, error) {
	data, err := c.cache.Get(c.data.GetKey())
	if err == nil {
		return data, nil
	}
	_, err = c.Save()
	if err != nil {
		return nil, err
	}
	return c.Get()
}

func (c *cacheService) SetData(d SimpleCacheData) {
	c.data = d
}
