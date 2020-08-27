package main

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"os"
)

var ctx = context.Background()

func NewRedis() (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		Password: "",
		DB:       0,
	})
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}
	return rdb, nil
}
