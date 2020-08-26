package main

import (
	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	"math/rand"
	"net/http"
)

func MainPage() echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello World")
	}
}

func GetPage() echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.String(http.StatusOK, "Message page")
	}
}

func PostHandle(queue chan Job) echo.HandlerFunc {
	return func(c echo.Context) error {
		message := new(Payload)
		if err := c.Bind(message); err != nil {
			log.Print(err)
			return err
		}
		message.MessageId = message.MessageId + rand.Int31()
		queue <- Job{Payload: *message}
		return c.JSON(http.StatusCreated, nil)
	}

}

func main() {

	JobQueue := make(chan Job)

	worker := NewWorker(JobQueue)
	worker.Start()

	e := echo.New()

	e.GET("/", MainPage())
	e.GET("/message", GetPage())
	e.POST("/message", PostHandle(JobQueue))

	e.Start(":8000")
}
