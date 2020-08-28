package main

import "github.com/go-redis/redis/v8"

func GetTotalNum(redis *redis.Client) (res int64, err error) {
	res, err = redis.DBSize(ctx).Result()
	return
}
