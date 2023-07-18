package cron

import (
	"log"
	"time"
)

const (
	continueStep = 0
	stopStep     = 1
)

func (engine *Engine) run() {
	go engine.clock()
	engine.h.timeChan1 <- 1
	<-engine.h.timeChan2
	for {
		//log.Println(engine.nextTime)
		<-time.After(engine.nextTime)
		select {
		case choice := <-engine.ch:
			switch choice {
			case continueStep:
				go engine.execute()
			case stopStep:
				log.Println("结束任务")
				return
			}
		}
		//log.Println(engine.h.timeLine)
		engine.h.timeChan1 <- 1
		<-engine.h.timeChan2
	}
}

func (engine *Engine) execute() {
	engine.function()
}
