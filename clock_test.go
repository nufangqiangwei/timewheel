package timeWheel

import (
	"encoding/json"
	"testing"
)

func TestAddTask(t *testing.T) {
	clockObj := newClock()
	task := &selfTask{
		Job: func() {
			println("执行时间：", clockObj.getNowTime().Format("2006-01-02 15:04:05"), "执行次数：", callbackExecNumber)
			callbackExecNumber++
		},
		jobName: "printTimeWheelTime",
		key:     843548134565,
		crontab: &Crontab{Spec: "7 */3 */4 * *"},
	}
	clockObj.addTask(task)
	var result []ManageTask
	result = getRouletteTaskInfo(clockObj.year)
	result = append(result, getRouletteTaskInfo(clockObj.month)...)
	result = append(result, getRouletteTaskInfo(clockObj.day)...)
	result = append(result, getRouletteTaskInfo(clockObj.hour)...)
	result = append(result, getRouletteTaskInfo(clockObj.minute)...)
	result = append(result, getRouletteTaskInfo(clockObj.second)...)
	t.Log(result)
	data, err := json.Marshal(result)
	if err != nil {
		println(err.Error())
	} else {
		println(string(data))
	}
}
