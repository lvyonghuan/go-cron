package cron

import (
	"time"
)

//计时

// 计时器
func (engine *Engine) clock() {
	go engine.storeClock()
	time.Sleep(1 * time.Second)
	for {
		if len(engine.h.timeLine) == 0 {
			engine.ch <- 1
			return
		} else {
			engine.ch <- 0
		}
		<-engine.h.timeChan1
		nextTimeLine := engine.h.timeLine[0]
		engine.nextTime = nextTimeLine.Sub(time.Now())
		engine.h.mu.Lock()
		engine.h.timeLine = engine.h.timeLine[1:]
		engine.h.mu.Unlock()
		engine.h.timeChan2 <- 1
	}
}

// 将下一个时刻存储进一个队列里，用的时候再取出来。队列连续追加到70，到达之后睡30s再判断。最坏的情况就是每秒都得跑。拿10秒做冗余。
func (engine *Engine) storeClock() {
	//指示切片下标的位置
	var (
		secondNum     = -1
		minuteNum     = -1
		hourNum       = -1
		dayNum        = -1
		monthNum      = -1
		weekNum       = -1
		yearNum       = 0
		yearSubscript = 0
	)
	var (
		tempSecond = 0
		tempMinute = 0
		tempHour   = 0
		tempDay    = 0
		tempMonth  = 0
		tempWeek   = 0
	)
	if len(engine.h.y) == 0 {
		yearNum = time.Now().Year()
	} else {
		yearNum = engine.h.y[0]
	}
	if len(engine.h.m) == 0 {
		monthNum = int(time.Now().Month())
	}
	if len(engine.h.d) == 0 {
		dayNum = int(time.Now().Day())
	}
	for {
		engine.judge(&secondNum, &minuteNum, &hourNum, &dayNum, &monthNum, &weekNum, &yearNum, &yearSubscript)
		//log.Println(secondNum, minuteNum, hourNum, dayNum, monthNum, weekNum, yearNum)
		//获取时间值
		if len(engine.h.s) != 0 {
			tempSecond = engine.h.s[secondNum]
		} else {
			tempSecond = secondNum
		}
		if len(engine.h.min) != 0 {
			tempMinute = engine.h.min[minuteNum]
		} else {
			tempMinute = minuteNum
		}
		if len(engine.h.h) != 0 {
			tempHour = engine.h.h[hourNum]
		} else {
			tempHour = hourNum
		}
		if len(engine.h.d) != 0 {
			tempDay = engine.h.d[dayNum]
		} else {
			tempDay = dayNum
		}
		if len(engine.h.m) != 0 {
			tempMonth = engine.h.m[monthNum]
		} else {
			tempMonth = monthNum
		}
		if len(engine.h.week) != 0 && weekNum != -1 {
			tempWeek = engine.h.week[weekNum]
			tempDay = getNthWeekdayOfMonth(yearNum, time.Month(tempMonth), time.Weekday(tempWeek))
		} else if weekNum != -1 {
			tempWeek = weekNum
			tempDay = getNthWeekdayOfMonth(yearNum, time.Month(tempMonth), time.Weekday(tempWeek))
		}
		tempTime := time.Date(yearNum, time.Month(tempMonth), tempDay, tempHour, tempMinute, tempSecond, 0, time.Local)

		//log.Println(tempTime)
		engine.h.mu.Lock()
		engine.h.timeLine = append(engine.h.timeLine, tempTime)
		engine.h.mu.Unlock()
		if len(engine.h.timeLine) >= 70 {
			time.Sleep(30 * time.Second)
		}
	}
}

// 进位指示器，低级每次遍历完一次之后，高级可以进位一次。第一次大家都需要判断。故第一次全部为true。判断完成之后，指示器状态改动为false，直到下一次进位。
var (
	minuteCan = true
	hourCan   = true
	dayCan    = true
	monthCan  = 0 //月进位的机制不一样。月进位依靠日进行判断。
	weekCan   = true
	yearCan   = 0 //年进位的机制也是不一样的。年进位依靠日进行判断。
)

// 判断器，用于判断并移动各个切片下标（早说把各个具体判断抽象成函数得了，，，
func (engine *Engine) judge(secondNum, minuteNum, hourNum, dayNum, monthNum, weekNum, yearNum, yearSubscript *int) {
	h := engine.h
	//对秒进行处理
	//首先判断切片长度是否是0。如果切片长为0，则表示
	if len(h.s) != 0 {
		//再根据具体情况进行判断并移动下标
		if len(h.s)-1 > *secondNum {
			*secondNum++
		} else if len(h.s)-1 == *secondNum { //当下标等于切片长-1，即下标移动到切片末尾的时候时归零
			*secondNum = 0
			minuteCan = true //进位指示器改动为true
		}
	} else { //假如切片长为0，则代表每秒/每分...都要加1，此时下标指示直接用于显示秒、分（而不通过切片设定的数值）
		*secondNum++
		//如果该进位了，则归零。
		if *secondNum == 60 {
			*secondNum = 0
			minuteCan = true //设置进位指示器
		}
	}

	//对分钟进行处理
	//检查进位指示器
	if minuteCan {
		if len(h.min) != 0 {
			//重复对秒的判断操作
			if len(h.min)-1 > *minuteNum {
				*minuteNum++
			} else if len(h.min)-1 == *minuteNum {
				*minuteNum = 0
				hourCan = true
			}
			minuteCan = false //进位完成，本级进位指示器改动为false
		} else {
			*minuteNum++
			if *minuteNum == 60 {
				*minuteNum = 0
				hourCan = true
			}
			minuteCan = false //进位完成，本级进位指示器改动为false
		}
	}

	//对小时进行处理
	if hourCan {
		if len(h.h) != 0 {
			if len(h.h)-1 > *hourNum {
				*hourNum++
			} else if len(h.h)-1 == *hourNum {
				*hourNum = 0
				//小时需要维护两个进位指示器
				dayCan = true
				weekCan = true
			}
		} else {
			*hourNum++
			if *hourNum == 24 {
				*hourNum = 0
				dayCan = true
				weekCan = true
			}
		}
		hourCan = false
	}

	//对天进行处理。不管周是否启用，天的计数器都会运行，以进行月进位。
	if dayCan {
		var tempDay int
		if len(h.d) != 0 {
			if len(h.d)-1 > *dayNum {
				*dayNum++
			} else if len(h.d)-1 == *dayNum {
				*dayNum = 0
				monthCan = 1
			}
			tempDay = h.d[*dayNum]
		} else {
			*dayNum++
			tempDay = *dayNum
		}
		//天需要特殊处理。对每月的天数进行判断。
		//先对当前月份进行临时判断
		var tempMonth int
		if len(h.m) != 0 {
			if *monthNum == -1 {
				tempMonth = h.m[0]
			} else {
				if monthCan == 0 {
					tempMonth = h.m[*monthNum]
				} else {
					tempMonth = h.m[(*monthNum)+1]
				}
			}
		} else {
			if *monthNum == -1 {
				*monthNum = 1
			} else {
				tempMonth = *monthNum
			}
		}

		//判断月进位和年进位
		for i, monthDay := tempMonth, 0; monthDay <= tempDay; {
			monthDay = daysInMonth(*yearNum, time.Month(i))
			if monthDay <= tempDay {
				break
			} else {
				monthCan++
				if len(h.m) != 0 {
					//在这里直接进行年进位判断
					if *monthNum+monthCan > len(h.m) {
						yearCan++
						*monthNum = 0
						monthCan = 0
					}
					tempMonth = h.m[*monthNum+monthCan]
				} else {
					tempMonth++
					if tempMonth > 12 {
						yearCan++
						*monthNum = 1
						monthCan = 0
					}
				}
			}
			//进行年进位
			if yearCan > 0 {
				if len(h.y) != 0 {
					if *yearSubscript+yearCan > len(h.y) {
						return //到头了，停止
					} else {
						*yearSubscript += yearCan
						*yearNum = h.y[*yearSubscript]
					}
				} else {
					*yearNum++
					yearCan = 0
				}
			}
		}
		//进行月进位
		*monthNum = tempMonth
		monthCan = 0
		yearCan = 0
		dayCan = false
	}

	//周进位判断
	if len(h.week) != 0 {
		if weekCan {
			if len(h.week) != 0 {
				if len(h.week)-1 > *weekNum {
					*weekNum++
				} else {
					*weekNum = 0
				}
			} else {
				if *weekNum == -1 {
					*weekNum = 1
				}
				*weekNum++
				if *weekNum > 7 {
					*weekNum = 1
				}
			}
		}
	}
}

// 获取一个月该有多少天
func daysInMonth(year int, month time.Month) int {
	// 当前月的最后一天
	firstDay := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	lastDay := firstDay.AddDate(0, 1, -1)
	// 获取当前月的天数
	days := lastDay.Day()
	return days
}

// 获取星期-天数的关系
func getNthWeekdayOfMonth(year int, month time.Month, targetWeekday time.Weekday) int {
	// 获取当前月份的第一天
	firstOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	// 计算目标星期与当前月份第一天的星期差值
	daysUntilTargetWeekday := int(targetWeekday-firstOfMonth.Weekday()+7) % 7
	// 计算目标星期在每月中是第几天
	nthDay := 1 + daysUntilTargetWeekday
	return nthDay
}
