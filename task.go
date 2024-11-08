package timeWheel

// task 延时任务
type selfTask struct {
	delay        int64            // 延迟时间
	rouletteSite map[string]int64 // 在每个时间轮的位置
	key          int64            // 定时器唯一标识, 用于删除定时器
	Job          func()           // Job 延时任务回调函数
	crontab      *Crontab         // 重复调用时间表
	jobName      string           // 任务标记。打印给用户看的
}

func (st *selfTask) RunJob() error {
	st.Job()
	return nil
}
func (st *selfTask) GetSchedule() *Crontab {
	return st.crontab
}
func (st *selfTask) GetJobName() string {
	return st.jobName
}

//	type Task struct {
//		Job     func(interface{}) // 需要执行的任务
//		JobData interface{}       // 任务参数
//		Repeat  bool              // 是否需要重复执行 false 单次任务 true 重复任务
//		Crontab Crontab           // 计划执行时间
//		JobName string            // 任务名称
//	}
type Task interface {
	RunJob() error
	GetSchedule() *Crontab
	GetJobName() string
}
