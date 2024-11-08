package main

import (
	"fmt"
	"github.com/nufangqiangwei/timewheel"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
)

func main() {
	fmt.Printf("%+v\n", Goroutine())
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
	debugLog.Printf("%+v\n", Goroutine())

	tasks := []timeWheel.Task{
		{
			Job:     func(data interface{}) { println(data) },
			JobData: "测试1",
			Repeat:  true,
			Crontab: timeWheel.Crontab{ExpiredTime: 3600 * 24},
			JobName: "测试1",
		}, {
			Job:     func(data interface{}) { println(data) },
			JobData: "测试2",
			Repeat:  true,
			Crontab: timeWheel.Crontab{Hour: ""},
			JobName: "测试2",
		},
	}

	tw := timeWheel.NewTimeWheel(&timeWheel.WheelConfig{Log: debugLog, BeatSchedule: tasks})
	tw.AppendOnceFunc(func(data interface{}) { println(data) }, 10, "debug任务", timeWheel.Crontab{ExpiredTime: 3600 * 2})
	//tw.AppendCycleFunc(func(i interface{}) {
	//	debugLog.Print(i)
	//	debugLog.Printf("goroutines数量：%+v", Goroutine())
	//}, "打印goroutines", "打印线程数", timeWheel.Crontab{Minute: "/1"})

	debugLog.Printf("%+v\n", Goroutine())

	tw.Start()
}

func Goroutine() map[string]interface{} {
	res := map[string]interface{}{}
	res["goroutines"] = runtime.NumGoroutine()
	res["OS threads"] = pprof.Lookup("threadcreate").Count()
	res["GOMAXPROCS"] = runtime.GOMAXPROCS(0)
	res["num CPU"] = runtime.NumCPU()
	return res
}
