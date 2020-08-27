package main

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"reflect"
	"strconv"
	"time"
)

type MessageController struct {
}

// redisから最新100件のメッセージを取得
func (ctrl *MessageController) GetAll() (res []map[string]string, err error) {
	redisInst, err := NewRedis()
	if err != nil {
		return
	}
	now := time.Now().UnixNano()
	messageIds, err := redisInst.ZRevRangeByScore(
		ctx,
		"messages/createdAt",
		&redis.ZRangeBy{
			Min:   "0",
			Max:   strconv.Itoa(int(now)),
			Count: 100},
	).Result()

	//HACK: redisに毎回アクセスしているのでsqlのinみたいにkeyを絞り込みたいが今のところやり方不明
	for _, messageId := range messageIds {
		val, err := redisInst.HGetAll(ctx, messageId).Result()
		if err != nil {
			break
		}
		delete(val, "userId")
		delete(val, "channelId")

		res = append(res, val)
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
