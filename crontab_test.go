package timeWheel

import (
	"fmt"
	"testing"
	"time"
)

func TestCrontab(t *testing.T) {
	now := timestamp{
		year:   2021,
		month:  9,
		day:    9,
		hour:   12,
		minute: 0,
		second: 0,
	}
	a := Crontab{Day: "10,20,30"}
	println(a.getNextExecTime(now))
	println(a.beforeRunTime.PrintTime())
	println(a.getNextExecTime(a.beforeRunTime))
	println(a.beforeRunTime.PrintTime())
	println(a.getNextExecTime(a.beforeRunTime))
	println(a.beforeRunTime.PrintTime())
	println(a.getNextExecTime(a.beforeRunTime))
	println(a.beforeRunTime.PrintTime())
	println(a.getNextExecTime(a.beforeRunTime))
	println(a.beforeRunTime.PrintTime())
}
func TestSplitArgs(t *testing.T) {
	initYear = 2021
	fmt.Printf("%v\n", splitArgs("/2", "month"))
	fmt.Printf("%v\n", splitArgs("", "month"))
	fmt.Printf("%v\n", splitArgs("/2", "day"))
	fmt.Printf("%v\n", splitArgs("", "second"))

}

func TestCrontabNextExecTime(t *testing.T) {
	defer func() {
		panicErr := recover()
		if panicErr != nil {
			_, ok := panicErr.(TimeOut)
			if ok {
				printLog("任务所有指定时间全部执行完毕，任务不在执行")
				return
			}
			printLog("执行函数出错。错误信息为%s", panicErr)
		}
	}()
	now := timestamp{
		year:   2021,
		month:  9,
		day:    9,
		hour:   12,
		minute: 0,
		second: 0,
	}
	initYear = 2021
	c := Crontab{
		Year:  "2021",
		Month: "/3",
		Day:   "10,20,30"}
	c.getNextExecTime(now)
	println(c.beforeRunTime.PrintTime())
	c.getNextExecTime(c.beforeRunTime)
	println(c.beforeRunTime.PrintTime())
	c.getNextExecTime(c.beforeRunTime)
	println(c.beforeRunTime.PrintTime())
	c.getNextExecTime(c.beforeRunTime)
	println(c.beforeRunTime.PrintTime())
	c.getNextExecTime(c.beforeRunTime)
	println(c.beforeRunTime.PrintTime())
	c.getNextExecTime(c.beforeRunTime)
	println(c.beforeRunTime.PrintTime())
	c.getNextExecTime(c.beforeRunTime)
	println(c.beforeRunTime.PrintTime())
	c.getNextExecTime(c.beforeRunTime)
	println(c.beforeRunTime.PrintTime())
	c.getNextExecTime(c.beforeRunTime)
	println(c.beforeRunTime.PrintTime())
}

func TestCrontabAddTask(t *testing.T) {
	var (
		snapRoulette *roulette // 当前
		lastRoulette *roulette // 上一个轮盘
	)
	ti := time.Now()
	timeMap := map[string]int{
		"year":   ti.Year(),
		"month":  int(ti.Month()),
		"day":    ti.Day(),
		"hour":   ti.Hour(),
		"minute": ti.Minute(),
		"second": ti.Second(),
	}
	initYear = int64(timeMap["year"])
	for _, defaultModel := range timeList {
		snapRoulette = newRoulette(defaultModel, timeMap[defaultModel])
		if defaultModel == "year" {
			lastRoulette = snapRoulette
			continue
		}
		lastRoulette.afterRoulette = snapRoulette
		snapRoulette.beforeRoulette = lastRoulette
		lastRoulette = snapRoulette
	}
	now := timestamp{
		year:   2021,
		month:  9,
		day:    9,
		hour:   12,
		minute: 0,
		second: 0,
	}
	c := Crontab{
		Year:  "2022",
		Month: "/3",
		Day:   "10,20,30"}
	c.getNextExecTime(now)

	ttask := task{
		delay: 0,
		rouletteSite: map[string]int64{
			"year":   int64(c.beforeRunTime.year),
			"month":  int64(c.beforeRunTime.month),
			"day":    int64(c.beforeRunTime.day),
			"hour":   int64(c.beforeRunTime.hour),
			"minute": int64(c.beforeRunTime.minute),
			"second": int64(c.beforeRunTime.second),
		},
		key:     3245,
		Job:     func() {},
		crontab: &c,
		jobName: "jobName",
	}

	snapRoulette.addTask(&ttask)
	fmt.Printf("%v\n", snapRoulette.getYearRoulette().taskKeyMap)
}

func TestJudgment(t *testing.T) {
	now := timestamp{
		year:   2023,
		month:  10,
		day:    10,
		hour:   11,
		minute: 31,
		second: 51,
	}
	println(now.isFutureTime(timestamp{
		year:   2023,
		month:  10,
		day:    10,
		hour:   11,
		minute: 31,
		second: 51,
	}))
}
