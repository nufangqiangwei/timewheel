package timeWheel

import (
	"log"
	"os"
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
