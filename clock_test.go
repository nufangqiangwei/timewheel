package timeWheel

import "testing"

func TestAddTask(t *testing.T) {
	clockObj := newClock()
	task := &selfTask{
		Job: func() {
			println("执行时间：", clockObj.getNowTime().Format("2006-01-02 15:04:05"), "执行次数：", callbackExecNumber)
			callbackExecNumber++
		},
		jobName:     "printTimeWheelTime",
		key:         843548134565,
		expiredTime: expiredTime(90),
	}
	clockObj.addTask(task)
}
