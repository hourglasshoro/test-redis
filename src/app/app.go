package main

import (
	"github.com/joho/godotenv"
	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	"net/http"
	"strconv"
	"time"
)

const MaxQueue = 20000
const BatchJobWaitTime = 1000 * time.Millisecond
const MigrateWaitTime = 10 * time.Second

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Print(err)
	}

	JobQueue := make(chan Job, MaxQueue)

	worker := NewWorker(JobQueue)
	worker.Start()

	migrate := NewMigrateWorker()
	migrate.Start()

	messageId := 0

	e := echo.New()

	e.GET("/", GetIndex())
	e.GET("/messages", GetMessages())
	e.GET("/messages/:id", GetMessageDetail())
	e.POST("/messages", PostMessages(JobQueue, messageId))
	e.Start(":8000")
}

func GetIndex() echo.HandlerFunc {
	return func(c echo.Context) error {
		totalNum, err := GetTotalNum()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}

		return c.JSON(http.StatusOK, map[string]int64{
			"totalNum": totalNum,
		})
	}
}

func GetMessages() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctrl := MessageController{}
		res, err := ctrl.GetAll()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, res)
	}
}

func PostMessages(queue chan Job, id int) echo.HandlerFunc {
	return func(c echo.Context) error {
		message := new(Payload)
		if err := c.Bind(message); err != nil {
			log.Print(err)
			return err
		}
		message.MessageId = id

		// JWTかcacheを確認する処理の代わり
		message.UserId = 1
		message.ChannelId = 1

		id++
		queue <- Job{Payload: *message}
		return c.JSON(http.StatusCreated, nil)
	}
}

func GetMessageDetail() echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}

		ctrl := MessageController{}
		res, err := ctrl.GetById(id)
		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}

		return c.JSON(http.StatusOK, res)
	}
}
