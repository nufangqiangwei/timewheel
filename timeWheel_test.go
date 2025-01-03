package timeWheel

import (
	"encoding/json"
	"log"
	"os"
	"testing"
	"time"
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
	printTimeWheelTime := func() {
		println("执行时间：", tw.PrintTime(), "执行次数：", callbackExecNumber)
		callbackExecNumber++
	}
	go tw.AppendCycleFunc(printTimeWheelTime, "", 30, "")
	go func() {
		time.Sleep(time.Second)
		tw.AppendCycleFunc(func() {
			allTask := tw.GetAllTask()
			da, err := json.Marshal(allTask)
			if err != nil {
				println(err.Error())

			} else {
				println(string(da))
			}
		}, "打印全部任务", 10, "")
	}()
	tw.Start()
	println("结束")
}

func printTimeWheelTime(data interface{}) {
	println("执行时间：", "执行次数：", 1)
	panic("挂掉")
}
