package main

import (
	"github.com/labstack/gommon/log"
	"strconv"
	"time"
)

type Payload struct {
	UserId      int32  `json:"user_id"`
	ChannelId   int32  `json:"channel_id"`
	Body        string `json:"body"`
	MessageType string `json:"type"`
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

			//var duration = 1 * time.Second
			var batchJob []Job

			DuringFunc(1000*time.Millisecond,
				func() {
					select {
					case job := <-w.JobQueue:
						batchJob = append(batchJob, job)
					case <-w.quit:
						return
					}
				},
			)

			for _, job := range batchJob {
				pipe.Set(ctx, strconv.Itoa(int(job.Payload.UserId)), job.Payload.Body, 0)
			}
			_, err = pipe.Exec(ctx)
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
