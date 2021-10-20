package timeWheel

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"
)

var (
	callbackExecNumber  int
	callbackExecNumber2 = 1
	callbackExecNumber3 = 1
	callbackExecNumber4 = 1
	debugLog            *log.Logger
)

func initLog() {
	logFile, err := os.OpenFile("timeWheelRun.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("open file error !")
	}
	debugLog = log.New(logFile, "[Debug]", log.LstdFlags)
	debugLog.SetOutput(logFile)
}

func TestTimeWheel(t *testing.T) {
	beatSchedule := []Task{{
		Job:     shiyanhanshu3,
		JobData: nil,
		Repeat:  true,
		Crontab: Crontab{Year: "2022", Month: "1,5,8,12", Day: "8,18,28"},
		JobName: "shiyanhanshu3",
	}}
	initLog()
	tw := NewTimeWheel(&WheelConfig{Log: debugLog, BeatSchedule: beatSchedule})

	printTimeWheelTime := func(data interface{}) {
		printLog("执行时间：%s 执行次数: %d \n", tw.PrintTime(), callbackExecNumber)
		callbackExecNumber++
	}
	go appendTask(tw, printTimeWheelTime, nil, "测试任务", Crontab{ExpiredTime: 36000}, 2, "printTimeWheelTime")
	go appendTask(tw, shiyanhanshu1, nil, "", Crontab{Month: "/2", Day: "11,21,31", Hour: "12,23"}, 10, "shiyanhanshu1")
	go appendTask(tw, shiyanhanshu2, nil, "", Crontab{Month: "/3", Day: "9,19,29", Hour: "6", Minute: "30"}, 16, "shiyanhanshu2")

	tw.Start()

}

func appendTask(tw *TimeWheel, job func(interface{}), jobData interface{}, jobName string, expiredTime Crontab, sleepTime time.Duration, funcName string) {
	time.Sleep(time.Second * sleepTime)
	taskId, err := tw.AppendCycleFunc(job, jobData, jobName, expiredTime)
	if err != nil {
		printLog("%s 任务添加失败 %s ", funcName, err.Error())
	} else {
		printLog("%s 任务添加完成，任务id是 %d。 调用时间表是 %s ", funcName, taskId,
			fmt.Sprintf("{year:%s, month:%s, day:%s, hour:%s, minute:%s }", expiredTime.Year, expiredTime.Month, expiredTime.Day, expiredTime.Hour, expiredTime.Minute))
	}
}

func shiyanhanshu1(data interface{}) {
	printLog("1111111111 执行次数：%d \n", callbackExecNumber2)
	callbackExecNumber2++
}
func shiyanhanshu2(data interface{}) {
	printLog("2222222222 执行次数：%d \n", callbackExecNumber3)
	callbackExecNumber3++
}
func shiyanhanshu3(data interface{}) {
	printLog("333333333 执行次数：%d \n", callbackExecNumber4)
	callbackExecNumber4++
}
