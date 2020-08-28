package main

import (
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	"net/http"
	"strconv"
	"time"
)

const MaxQueue = 20000
const BatchJobWaitTime = 1000 * time.Millisecond
const MigrateWaitTime = 5 * time.Second

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Print(err)
	}

	redisInst, err := NewRedis()
	if err != nil {
		log.Print(err)
	}

	messageController := *NewMessageController(redisInst)

	JobQueue := make(chan Job, MaxQueue)

	redisUpdate := NewRedisUpdateWorker(redisInst, JobQueue)
	redisUpdate.Start()

	migrate := NewMigrateWorker()
	migrate.Start()

	e := echo.New()

	e.GET("/", GetIndex(redisInst))
	e.GET("/channel/:id/messages", GetMessages(messageController))
	e.GET("/messages/:id", GetMessageDetail(messageController))
	e.POST("/messages", PostMessages(JobQueue))
	e.Start(":8000")
}

func GetIndex(redis *redis.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		totalNum, err := GetTotalNum(redis)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}

		return c.JSON(http.StatusOK, map[string]int64{
			"totalNum": totalNum,
		})
	}
}

func GetMessages(ctrl MessageController) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
		res, err := ctrl.GetAll(id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, res)
	}
}

func PostMessages(queue chan Job) echo.HandlerFunc {
	return func(c echo.Context) error {
		message := new(Payload)
		if err := c.Bind(message); err != nil {
			log.Print(err)
			return err
		}
		message.MessageId = uuid.New()

		// JWT解凍する処理の代わり
		message.UserId = 1

		queue <- Job{Payload: *message}
		return c.JSON(http.StatusCreated, nil)
	}
}

func GetMessageDetail(ctrl MessageController) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}

		res, err := ctrl.GetById(id)
		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}

		return c.JSON(http.StatusOK, res)
	}
}
