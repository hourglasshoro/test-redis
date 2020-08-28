package main

import (
	"fmt"
	"github.com/bamzi/jobrunner"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/labstack/gommon/log"
	"time"
)

type Payload struct {
	MessageId   uuid.UUID
	Body        string `json:"body"`
	DisplayName string `json:"displayName"`
	MessageType string `json:"type"`
	ChannelId   int
	UserId      int
}

type Job struct {
	Payload Payload
}

type RedisUpdateWorker struct {
	RedisUpdate RedisUpdate
	Redis       *redis.Client
}

func NewRedisUpdateWorker(redis *redis.Client, jobQueue chan Job) *RedisUpdateWorker {
	return &RedisUpdateWorker{
		RedisUpdate: *NewRedisUpdate(jobQueue),
		Redis:       redis,
	}
}

func (w *RedisUpdateWorker) Start() {
	pipe := w.Redis.Pipeline()
	jobrunner.Start()
	jobrunner.Every(BatchJobWaitTime, &w.RedisUpdate)

	go func() {
		for {
			// ループ毎にbatchJobChannelに保存されたバッチ化された複数個のJobを取り出す
			go func(batchJob chan Job) {
				// 複数あるJobを並行で変換してredisのpipelineに保存
				for {
					job := <-batchJob
					key := fmt.Sprintf("message:%s", job.Payload.MessageId)
					now := time.Now().UnixNano()
					err := pipe.HMSet(
						ctx,
						key,
						"body", job.Payload.Body,
						"displayName", job.Payload.DisplayName,
						"type", job.Payload.MessageType,
						"createdAt", now,
						"userId", job.Payload.UserId,
						"channelId", job.Payload.ChannelId,
					).Err()
					if err != nil {
						log.Print(err)
					}

					err = pipe.SAdd(
						ctx,
						fmt.Sprintf("messages/channelId:%d", job.Payload.ChannelId),
						key,
					).Err()
					if err != nil {
						log.Print(err)
					}

					err = pipe.ZAdd(
						ctx,
						"messages/createdAt",
						&redis.Z{
							Score:  float64(now),
							Member: key,
						},
					).Err()

					if err != nil {
						log.Print(err)
					}
				}
			}(<-w.RedisUpdate.batchJobChannel)

			// 保存されたpipelineのコマンドを実行
			_, err := pipe.Exec(ctx)
			if err != nil {
				log.Print(err)
				break
			}
		}
	}()
}

type RedisUpdate struct {
	JobQueue        chan Job
	batchJobChannel chan chan Job
}

func NewRedisUpdate(jobQueue chan Job) *RedisUpdate {
	return &RedisUpdate{
		JobQueue:        jobQueue,
		batchJobChannel: make(chan chan Job),
	}
}

// job runnerによって発火する関数
func (ru *RedisUpdate) Run() {
	// 例えば1秒ごとであれば1秒ごとにJobQueueがバッチ化されてbatchJobChannelに追加される
	ru.batchJobChannel <- ru.JobQueue
}
