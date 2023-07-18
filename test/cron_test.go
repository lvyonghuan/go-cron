package test

import (
	"go-cron"
	"log"
	"testing"
)

func TestCron(t *testing.T) {
	engine := cron.CreateEngine()
	err := engine.Set("5 * * * * * *", func() {
		log.Println("hello")
	})
	if err != nil {
		log.Fatalf(err.Error())
	}
	engine.Run()
}
