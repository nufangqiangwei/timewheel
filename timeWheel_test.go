package timeWheel

import (
	"log"
	"os"
	"testing"
)

func TestTimeWheel_AppendCycleFunc(t *testing.T) {
	logFile, err := os.OpenFile("logs.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	defer logFile.Close()
	if err != nil {
		log.Fatalln("open file error !")
	}
	debugLog := log.New(logFile, "[Debug]", log.LstdFlags)
	debugLog.Println("A debug message here")
	debugLog.SetPrefix("[Info]")
	debugLog.Println("A Info Message here ")
	//配置log的Flag参数
	debugLog.SetFlags(debugLog.Flags() | log.LstdFlags)
	debugLog.Println("A different prefix")

	tw := NewTimeWheel(&WheelConfig{Log: debugLog})
	printTimeWheelTime := func(data interface{}) {
		println("执行时间：", tw.PrintTime(), "执行次数：", callbackExecNumber)
		callbackExecNumber++
	}
	go tw.AppendCycleFunc(printTimeWheelTime, 2, "测试任务", Crontab{Day: "10,20,30"})
	tw.Start()
}

func printTimeWheelTime(data interface{}) {
	println("执行时间：", "执行次数：", 1)
	panic("挂掉")
}

func TestAppendTask(t *testing.T) {
	//tw := NewTimeWheel(nil)
	initYear = 2021
	c := Crontab{
		Day: "10,20,30",
	}
	c.getNextExecTime(timestamp{
		year:   2021,
		month:  10,
		day:    10,
		hour:   11,
		minute: 31,
		second: 51,
	})
	println(c.beforeRunTime.PrintTime())

}
