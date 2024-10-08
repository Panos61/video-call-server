package redisclient

import (
	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client

func InitRedisClient() *redis.Client {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	return rdb
}

func GetRedisClient() *redis.Client {
	if rdb == nil {
		rdb = InitRedisClient()
	}

	return rdb
}
