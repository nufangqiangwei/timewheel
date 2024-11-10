package timeWheel

import "time"

type Task interface {
	RunJob()
	GetSchedule() Schedule
	NextRunTime(time.Time) time.Time
	GetJobName() string
	SetTaskKey(key int64)
	GetTaskKey() int64
	GetLastRunTime() *time.Time
	SetLastRunTime(t time.Time)
	GetExecNumber() int64
}

// task 延时任务
type selfTask struct {
	expiredTime         // 延迟时间
	crontab      string // 重复调用时间表
	delay        int64
	rouletteSite map[string]int64 // 在每个时间轮的位置
	key          int64            // 定时器唯一标识, 用于删除定时器
	Job          func()           // Job 延时任务回调函数
	jobName      string           // 任务标记。打印给用户看的
	schedule     Schedule
	lastRunTime  *time.Time
	execNumber   int64 // 执行次数
}

func (st *selfTask) RunJob() {
	st.Job()
	a := time.Now()
	st.lastRunTime = &a
	st.execNumber++
}
func (st *selfTask) GetSchedule() Schedule {
	if st.schedule == nil {
		if st.crontab != "" {
			var err error
			st.schedule, err = standardParser.parse(st.crontab)
			if err != nil {
				panic(CrontabError{err.Error()})
			}
		} else if st.expiredTime > 0 {
			st.schedule = &st.expiredTime
		}
	}
	return st.schedule
}
func (st *selfTask) GetJobName() string {
	return st.jobName
}
func (st *selfTask) SetTaskKey(key int64) {
	st.key = key
}
func (st *selfTask) GetTaskKey() int64 {
	return st.key
}
func (st *selfTask) GetLastRunTime() *time.Time {
	return st.lastRunTime
}
func (st *selfTask) SetLastRunTime(t time.Time) {
	st.lastRunTime = &t
}
func (st *selfTask) GetExecNumber() int64 {
	return st.execNumber
}
