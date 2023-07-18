package cron

import (
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	second = 0
	minute = 1
	hour   = 2
	day    = 3
	month  = 4
	week   = 5
	year   = 6
)

var (
	timeFormatErr = errors.New("错误的时间格式") //时间格式错误
)

type Engine struct {
	time     string //cron表达式，暂时不支持cron的全部表达形式，只支持具体数字和*
	h        *handle
	ch       chan int //控制进程，发送停止执行信号
	function func()   //准备执行的函数
	nextTime time.Duration
}

// 内部调用
type handle struct {
	s         []int       //秒
	min       []int       //分
	h         []int       //时,使用24h制
	d         []int       //日
	m         []int       //月
	week      []int       //星期。若已有日期字段则不触发。
	y         []int       //年
	timeLine  []time.Time //时刻队列
	timeChan1 chan int    //goroutine间通信管道
	timeChan2 chan int
	mu        sync.Mutex //防止并发问题
}

// CreateEngine 创建一个新的cron引擎
func CreateEngine() Engine {
	return Engine{}
}

// Set 设定表达式和要执行的函数
func (engine *Engine) Set(expression string, function func()) error {
	engine.time = expression
	engine.function = function
	return engine.handelExpression()
}

func (engine *Engine) Run() {
	engine.ch = make(chan int, 1)
	engine.h.timeChan1 = make(chan int, 1)
	engine.h.timeChan2 = make(chan int, 1)
	engine.run()
}

// 处理cron字符串
func (engine *Engine) handelExpression() error {
	//将表达式按照空格进行切分
	expressions := strings.Split(engine.time, " ")
	h := handle{}
	var weekFlag = true
	for i, exp := range expressions {
		switch i {
		case second:
			err := h.handelSecond(exp)
			if err != nil {
				return err
			}
		case minute:
			err := h.handelMin(exp)
			if err != nil {
				return err
			}
		case hour:
			err := h.handelHour(exp)
			if err != nil {
				return err
			}
		case day:
			weekFlag = false
			err := h.handelDay(exp)
			if err != nil {
				return err
			}
		case month:
			err := h.handelMonth(exp)
			if err != nil {
				return err
			}
		case week:
			if !weekFlag {
				err := h.handelWeek(exp)
				if err != nil {
					return err
				}
			}
		case year:
			err := h.handelYear(exp)
			if err != nil {
				return err
			}
		}
	}
	engine.h = &h
	return nil
}

// 解析秒
func (h *handle) handelSecond(expression string) error {
	exp := strings.Split(expression, ",")
	for _, e := range exp {
		//判断该字段是否为空
		if e == "*" {
			return nil
		}
		//将字符串转化为秒
		s, err := strconv.Atoi(e)
		if err != nil {
			return timeFormatErr
		}
		//检查正确性
		if s < 0 || s >= 60 {
			return timeFormatErr
		}
		//将结果追加
		h.s = append(h.s, s)
	}
	return nil
}

// 解析分
func (h *handle) handelMin(expression string) error {
	exp := strings.Split(expression, ",")
	for _, e := range exp {
		if e == "*" {
			return nil
		}
		m, err := strconv.Atoi(e)
		if err != nil {
			return timeFormatErr
		}
		if m < 0 || m > 60 {
			return timeFormatErr
		}
		h.min = append(h.min, m)
	}
	return nil
}

// 解析时
func (h *handle) handelHour(expression string) error {
	exp := strings.Split(expression, ",")
	for _, e := range exp {
		if e == "*" {
			return nil
		}
		hour, err := strconv.Atoi(e)
		if err != nil {
			return timeFormatErr
		}
		if hour < 0 || hour >= 24 {
			return timeFormatErr
		}
		h.h = append(h.h, hour)
	}
	return nil
}

// 解析日
func (h *handle) handelDay(expression string) error {
	exp := strings.Split(expression, ",")
	for _, e := range exp {
		if e == "*" {
			return nil
		}
		d, err := strconv.Atoi(e)
		if err != nil {
			return timeFormatErr
		}
		//TODO:这里多少有点太粗暴了
		if d < 0 || d > 31 {
			return timeFormatErr
		}
		h.d = append(h.d, d)
	}
	return nil
}

// 解析月
func (h *handle) handelMonth(expression string) error {
	exp := strings.Split(expression, ",")
	for _, e := range exp {
		if e == "*" {
			return nil
		}
		m, err := strconv.Atoi(e)
		if err != nil {
			return timeFormatErr
		}
		if m < 0 || m > 12 {
			return timeFormatErr
		}
		h.m = append(h.m, m)
	}
	return nil
}

// 解析星期
func (h *handle) handelWeek(expression string) error {
	exp := strings.Split(expression, ",")
	for _, e := range exp {
		if e == "*" {
			return nil
		}
		w, err := strconv.Atoi(e)
		if err != nil {
			return timeFormatErr
		}
		if w < 0 || w > 7 {
			return timeFormatErr
		}
		h.week = append(h.week, w)
	}
	return nil
}

// 解析年
func (h *handle) handelYear(expression string) error {
	exp := strings.Split(expression, ",")
	for _, e := range exp {
		if e == "*" {
			return nil
		}
		y, err := strconv.Atoi(e)
		if err != nil {
			return timeFormatErr
		}
		h.y = append(h.y, y)
	}
	return nil
}
