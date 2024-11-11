package timeWheel

import "time"

type Task interface {
	RunJob()                         // 执行任务,最好在这里记录下上次的执行时间
	GetSchedule() Schedule           // 获取任务执行计划，主要用于再api接口去展示
	NextRunTime(time.Time) time.Time // 获取下次运行时间,建议实现的时候。先使用上次运行时间去计算，没有的话就使用当前时间
	GetJobName() string              // 任务标记。打印给用户看的
	SetTaskKey(key int64)            // 用于设置任务唯一标识，这个需要保留
	GetTaskKey() int64               // 用于获取任务唯一标识
	GetLastRunTime() *time.Time      // 上次的执行时间，可以返回nil
	SetLastRunTime(t time.Time)      // 设置上次的执行时间
	GetExecNumber() int64            // 执行次数
}

// task 延时任务
type selfTask struct {
	expiredTime           // 延迟时间
	crontab      *Crontab // 重复调用时间表
	delay        int64
	rouletteSite map[string]int64 // 在每个时间轮的位置
	key          int64            // 定时器唯一标识, 用于删除定时器
	Job          func()           // Job 延时任务回调函数
	jobName      string           // 任务标记。打印给用户看的
	funcName     string           // 任务标记。打印给用户看的
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
		if st.crontab != nil {
			err := st.crontab.init()
			if err != nil {
				panic(CrontabError{err.Error()})
			}
			st.schedule = st.crontab
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
func (st *selfTask) NextRunTime(nowTime time.Time) time.Time {
	if st.lastRunTime != nil {
		nowTime = *st.lastRunTime
	}
	return st.GetSchedule().NextRunTime(nowTime)
}
