package main

import (
	"context"
	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()
var rdb = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379",
	Password: "test", // no password set
	DB:       0,      // use default DB
})

func addRedis(key, value string) error {

	err := rdb.RPush(ctx, key, value).Err()
	if err != nil {
		panic(err)
	}

	return err
}

func readRedis(key string) []string {

	val, err := rdb.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		panic(err)
	}
	return val

}
