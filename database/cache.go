package database

import (
	"time"

	"github.com/go-redis/redis"
)

/*
Caching implementation:

we will be using redis for all caching purposes in this project.
for our purposes we will be caching the most useful names in our database to reduce the number of
calls we make to our database.
*/

// the time in ms cache will be kept
const CACHE_BUFFER = 10000

type redisCache struct {
	host    string
	db      int
	expires time.Duration
}

func NewRedisCache(host string, db int, expires time.Duration) *redisCache {
	return &redisCache{
		host:    host,
		db:      db,
		expires: expires,
	}
}

func (cache *redisCache) getClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     cache.host,
		Password: "",
		DB:       cache.db,
	})
}

func (cache *redisCache) Set(key string, value []byte) {
	client := cache.getClient()
	client.Set(key, value, cache.expires)
}

func (cache *redisCache) Get(key string) []byte {
	client := cache.getClient()
	val, err := client.Get(key).Result()
	CheckError(err)
	return []byte(val)
}
