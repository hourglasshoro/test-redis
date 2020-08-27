package main

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/labstack/gommon/log"
	"time"
)

type Payload struct {
	MessageId   int    `json:"messageId"`
	Body        string `json:"body"`
	DisplayName string `json:"displayName"`
	MessageType string `json:"type"`
	UserId      int
	ChannelId   int
}

type Job struct {
	Payload Payload
}

type Worker struct {
	JobQueue chan Job
	quit     chan bool
}

func NewWorker(queue chan Job) Worker {
	return Worker{
		JobQueue: queue,
		quit:     make(chan bool),
	}
}

func DuringFunc(d time.Duration, f func()) {
	flg := true
	go func() {
		for flg {
			f()
		}
	}()
	time.Sleep(d)
	flg = false
}

func (w Worker) Start() {
	redisInst, err := NewRedis()
	if err != nil {
		log.Print(err)
	}

	pipe := redisInst.Pipeline()

	go func() {
		for {

			batchJobChannel := make(chan Job, MaxQueue)
			counter := 0

			DuringFunc(BatchJobWaitTime,
				func() {
					select {
					case job := <-w.JobQueue:
						batchJobChannel <- job
						counter++
					case <-w.quit:
						return
					}
				},
			)

			for i := 0; i < counter; i++ {
				job := <-batchJobChannel
				key := fmt.Sprintf("message:%d", int(job.Payload.MessageId))
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

			_, err := pipe.Exec(ctx)
			if err != nil {
				log.Print(err)
				break
			}
		}
	}()
}

// Stop signals the worker to stop listening for work requests.
func (w Worker) Stop() {
	go func() {
		w.quit <- true
	}()
}
