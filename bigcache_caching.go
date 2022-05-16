package cached_caller

import (
	"time"

	"github.com/allegro/bigcache/v3"
)

type bigCacheDecorator struct {
	cache *bigcache.BigCache
}

var (
	defaultConfig = bigcache.DefaultConfig(20 * time.Minute)
)

func NewBigCaching(config ...bigcache.Config) Caching {
	c := defaultConfig
	if len(config) > 0 {
		c = config[0]
	}
	cache, err := bigcache.NewBigCache(c)
	if err != nil {
		panic(err)
	}
	res := &bigCacheDecorator{
		cache: cache,
	}
	return res
}

func (b *bigCacheDecorator) Get(key string) (value []byte, err error) {
	return b.cache.Get(key)
}

func (b *bigCacheDecorator) Put(key string, value []byte) error {
	return b.cache.Set(key, value)
}

func (b *bigCacheDecorator) Del(key string) error {
	return b.cache.Delete(key)
}

func (b *bigCacheDecorator) GetIter() CacheIter {
	return &bigcacheIter{iter: b.cache.Iterator()}
}
