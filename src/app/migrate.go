package main

import (
	"errors"
	"fmt"
	"github.com/bamzi/jobrunner"
	"github.com/go-redis/redis/v8"
	"github.com/labstack/gommon/log"
	"strconv"
	"strings"
	"time"
)

type MigrationWorker struct {
}

func NewMigrateWorker() MigrationWorker {
	return MigrationWorker{}
}

func (w *MigrationWorker) Start() {
	jobrunner.Start()
	jobrunner.Every(MigrateWaitTime, NewMigrate())
}

func NewMigrate() *Migrate {
	return &Migrate{
		LastRun: "0",
	}
}

type Migrate struct {
	LastRun string
}

func (m *Migrate) Run() {
	redisInst, err := NewRedis()
	if err != nil {
		log.Print(err)
	}

	mysqlInst, err := NewMysql()
	if err != nil {
		log.Print(err)
	}

	now := time.Now().UnixNano()

	// LastRunから現在までのmessageのIdを取得
	messageIds, err := redisInst.ZRevRangeByScore(
		ctx,
		"messages/createdAt",
		&redis.ZRangeBy{
			Min: m.LastRun,
			Max: strconv.Itoa(int(now)),
		},
	).Result()

	// あとからをvaluesを追加するバルクインサート用のクエリ
	query := `
INSERT INTO messages (body, user_id, channel_id, type, created_at)
VALUES
`
	var values []string

	//HACK: redisに毎回アクセスしているのでsqlのinみたいにkeyを絞り込みたいが今のところやり方不明
	for _, messageId := range messageIds {
		val, err := redisInst.HGetAll(ctx, messageId).Result()
		if err != nil {
			break
		}
		body, ok := val["body"]
		if !ok {
			err = errors.New("the body key does not exist")
			log.Print(err)
			break
		}
		userId, ok := val["userId"]
		if !ok {
			err = errors.New("the user id key does not exist")
			log.Print(err)
			break
		}
		channelId, ok := val["channelId"]
		if !ok {
			err = errors.New("the channel id key does not exist")
			log.Print(err)
			break
		}
		messageType, ok := val["type"]
		if !ok {
			err = errors.New("the type key does not exist")
			log.Print(err)
			break
		}
		createdAt, ok := val["createdAt"]
		if !ok {
			err = errors.New("the created at key does not exist")
			log.Print(err)
			break
		}
		createdAtInt, err := strconv.Atoi(createdAt)
		createdAtTime := time.Unix(int64(createdAtInt/1e9), 0)

		value := fmt.Sprintf(
			"('%s', %s, %s, '%s', '%s')",
			body,
			userId,
			channelId,
			messageType,
			createdAtTime.Format("2006-01-02 15:04:05"),
		)
		values = append(values, value)
	}

	// クエリに追加
	query += strings.Join(values, ", ")

	// クエリの実行
	if len(values) > 0 {
		_, err = mysqlInst.Exec(query)
		if err != nil {
			log.Print(err)
		} else {
			log.Print("Exec migration")
		}
		m.LastRun = strconv.Itoa(int(now))
	}
}
