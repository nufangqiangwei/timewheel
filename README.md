## 试用方法

```go
package main

import (
	"fmt"
	"github.com/nufangqiangwei/timewheel"
	"log"
)

var debugLog *log.Logger

func main() {
	tw := timeWheel.NewTimeWheel(&timeWheel.WheelConfig{Log: debugLog})
	printTimeWheelTime := func(data interface{}) {
		printLog("执行时间：%s 执行次数: %d \n", tw.PrintTime(), callbackExecNumber)
		callbackExecNumber++
	}
	go appendTask(tw, printTimeWheelTime, nil, "测试任务", timeWheel.Crontab{ExpiredTime: 100}, 2, "printTimeWheelTime")
	go appendTask(tw, shiyanhanshu1, nil, "", timeWheel.Crontab{Month: "/2", Day: "11,21,31", Hour: "12,23"}, 10, "shiyanhanshu1")
	tw.Start()

}

func appendTask(tw *timeWheel.TimeWheel, job func(interface{}), jobData interface{}, jobName string, expiredTime timeWheel.Crontab, sleepTime time.Duration, funcName string) {
	taskId, err := tw.AppendCycleFunc(job, jobData, jobName, expiredTime)
	if err != nil {
		fmt.Printf("%s 任务添加失败 %s ", funcName, err.Error())
	} else {
		fmt.Printf("%s 任务添加完成，任务id是 %d。 调用时间表是 %s ", funcName, taskId,
			fmt.Sprintf("{year:%s, month:%s, day:%s, hour:%s, minute:%s }", expiredTime.Year, expiredTime.Month, expiredTime.Day, expiredTime.Hour, expiredTime.Minute))
	}
}

```
### WheelConfig 介绍
    Log   输出的日志
    IsRun 是否直接启动，默认需要手动调用 TimeWheel.Start
    BeatSchedule 初始化的时候就添加的任务


### Crontab 介绍
    Crontab 时间执行表
    字符串 按照给定的数字，当时间到给定的刻度就会执行
    比如 Crontab{Minute:10,Second:30} 每个小时的十分三十秒的时候就会执行
    支持一次传入多个时间点 Crontab{Minute:"10,11,12",Second:30}每个小时的10：30，11：30，12：30三个时间点执行
    连续时间点可以用-表示 Minute:"10-12" 代表 Minute:"10,11,12"
    也可以 /5表示当时间点可以被5整除的时候就执行任务 前面可以写自己指定的时间段 默认的是当前时间段的起止 比如 秒 就是0-59
    10-20/5 表示当时间点在 10 15 20 这三个时间点执行任务
    参照python的celery的crontab实现的
    也可以直接传一个间隔时间，会以当前的时间点为起点，向后推迟到目标时间点执行任务

## 注意事项
    