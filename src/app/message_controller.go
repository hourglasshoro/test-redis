package main

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"reflect"
	"strconv"
	"sync"
	"time"
)

type MessageController struct {
	Redis *redis.Client
}

func NewMessageController(redis *redis.Client) *MessageController {
	return &MessageController{
		Redis: redis,
	}
}

// redisから最新100件のメッセージを取得
func (ctrl *MessageController) GetAll(channelId int) (res []map[string]string, err error) {
	now := time.Now().UnixNano()

	query := fmt.Sprintf("messages/channelId:%d", channelId)
	_, err = ctrl.Redis.ZInterStore(
		ctx,
		"outputKey",
		&redis.ZStore{
			Keys:      []string{query, "messages/createdAt"},
			Aggregate: "MAX",
		}).Result()

	messageKeys, err := ctrl.Redis.ZRevRangeByScore(
		ctx,
		"outputKey",
		&redis.ZRangeBy{
			Min:   "0",
			Max:   strconv.Itoa(int(now)),
			Count: 100},
	).Result()

	pipe := ctrl.Redis.Pipeline()

	wg := sync.WaitGroup{}
	m := sync.Map{}
	for _, messageKey := range messageKeys {
		wg.Add(1)
		go func(val string) {
			m.Store(val, pipe.HGetAll(ctx, val))
			if err != nil {
				wg.Done()
				return
			}
			wg.Done()

		}(messageKey)
	}

	wg.Wait()

	_, err = pipe.Exec(ctx)
	if err != nil {
		return
	}

	res = []map[string]string{}
	for _, messageKey := range messageKeys {
		v, ok := m.Load(messageKey)
		if !ok {
			break
		}
		var cmd = v.(*redis.StringStringMapCmd)
		result, err := cmd.Result()
		if err != nil {
			break
		}
		res = append(res, result)
	}
	return
}

func (ctrl *MessageController) GetById(id int) (res map[string]string, err error) {
	redisInst, err := NewRedis()
	if err != nil {
		return
	}
	key := fmt.Sprintf("message:%d", id)
	res, err = redisInst.HGetAll(ctx, key).Result()
	if reflect.DeepEqual(res, map[string]string{}) {
		err = errors.New("no such a message")
	}
	return
}
