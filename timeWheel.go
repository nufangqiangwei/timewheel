package timeWheel

import (
	"container/list"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	mapset "github.com/deckarep/golang-set"
)

// TimeWheel 时间轮
type TimeWheel struct {
	interval          time.Duration // 指针每隔多久往前移动一格
	ticker            *time.Ticker  // 时间间隔
	*clock                          // 时钟
	taskKeySet        mapset.Set    //taskKey集合
	addTaskChannel    chan Task     // 新增任务channel
	removeTaskChannel chan int64    // 删除任务channel
	stopChannel       chan bool     // 停止定时器channel
	running           bool          // 是否正在运行
	config            *WheelConfig
}

// WheelConfig 配置信息
type WheelConfig struct {
	BeatSchedule []Task      // 任务
	IsRun        bool        // 是否直接启动，默认需要手动调用 TimeWheel.Start
	Log          *log.Logger // 打印日志
}

// NewTimeWheel 调用实例，需要全局唯一，
// model: 模式，就是时间轮层数 年月日时分秒 year, month, day, hour, minute, second
// tickInterval:每次转动的时间间隔
// 使用方法
//
//	tw := NewTimeWheel(&WheelConfig{})
//	_ = tw.AppendOnceFunc(oneCallback, 1, 10)
//	err := tw.AppendCycleFunc(callbackFunc, 2, Crontab{
//		Second: "/5",
//	})
//	if err != nil {
//		tw.Stop()
//		println(err.Error())
//		return
//	}
//
// 工作大致说明
// TimeWheel.Start() 开始入口 ，通过监听*time.Ticker 每秒执行一次 TimeWheel.wheel.tickHandler() 这个方法
// 该方法每次执行都会在时间上 +1秒 ，每一个时间指针都指向一个list.List 链表，链表内存有 task 对象，被指针指到的链表，其内部所有的 task 都到了
// 执行时间，
func NewTimeWheel(config *WheelConfig) *TimeWheel {
	config = DefaultWheelConfig(config)
	tw := &TimeWheel{
		interval:          time.Second,
		addTaskChannel:    make(chan Task, 10),
		removeTaskChannel: make(chan int64),
		stopChannel:       make(chan bool),
		taskKeySet:        mapset.NewSet(),
		config:            config,
	}

	logObject = config.Log

	if config.IsRun {
		go tw.Start()
	}

	return tw
}

func DefaultWheelConfig(config *WheelConfig) *WheelConfig {
	if config == nil {
		config = &WheelConfig{}
	}
	return config
}

// Start 开始
func (tw *TimeWheel) Start() {
	if tw.running {
		printLog("已启动，无需再次启动")
		return
	}
	if tw.clock == nil {
		tw.clock = newClock()
		tw.clock.wheel = tw
	} else {
		tw.clock.resetTime()
	}
	printLog("定时器启动,当前时间 %s", tw.clock.getNowTime().Format("2006-01-02 15:04:05"))
	tw.running = true
	if tw.ticker == nil {
		tw.ticker = time.NewTicker(tw.interval)
	}
	if tw.config.BeatSchedule != nil {
		go func() {
			// 防止队列阻塞，所以使用goroutine
			var err error
			for _, task := range tw.config.BeatSchedule {
				_, err = tw.AppendTask(task)
				if err != nil {
					printLog("%s 任务添加出错 %s", task.GetJobName(), err.Error())
				}
			}
		}()
	}
	for {
		select {
		case <-tw.ticker.C:
			tw.clock.tickHandler()
		case task := <-tw.addTaskChannel:
			//printLog("收到一个任务id为 %d 的任务，在 %s 调用 当前时间的 %d 秒后调用", task.key, task.crontab.beforeRunTime.PrintTime(), task.delay)
			err := tw.clock.addTask(task)
			if err != nil {
				printLog("%s 任务添加出错 %s", task.GetJobName(), err.Error())
			}
		case key := <-tw.removeTaskChannel:
			tw.clock.year.removeTask(key)
		case <-tw.stopChannel:
			tw.ticker.Stop()
			tw.running = false
			return
		}
	}
}

// Stop 停止
func (tw *TimeWheel) Stop() {
	tw.stopChannel <- true
}

// AppendOnceFunc 添加单次任务
// job 回调函数
// jobData 回调函数的调用参数
// jobName 任务标记，打印出给用户查看
// expiredTime Crontab 对象
func (tw *TimeWheel) AppendOnceFunc(job func(), jobName string, delay int64, crontab string) (taskKey int64, err error) {
	if !tw.running {
		return 0, errors.New("定时器尚未启动，请先调用 Start 启动定时器")
	}
	functionName := getFunctionName(job)
	if jobName == "" {
		jobName = functionName
	}
	taskKey = tw.randomTaskKey()
	printLog("添加%s任务", jobName)
	task := &selfTask{
		jobName:  jobName,
		key:      taskKey,
		funcName: functionName,
	}
	if delay > 0 {
		task.expiredTime = expiredTime(delay)
	} else if crontab != "" {
		task.crontab = &Crontab{Spec: crontab}
	}
	task.Job = func() {
		// 不需要重复执行的任务，在这里将调度对象至空
		task.schedule = nil
		job()
	}

	tw.addTask(task)
	return
}

// AppendCycleFunc 添加重复任务
// job 回调函数
// jobData 回调函数的调用参数
// jobName 任务标记，打印出给用户查看
// expiredTime Crontab 对象
func (tw *TimeWheel) AppendCycleFunc(job func(), jobName string, delay int64, crontab string) (taskKey int64, err error) {
	//if !tw.running {
	//	return 0, errors.New("定时器尚未启动，请先调用 Start 启动定时器")
	//}
	taskKey = tw.randomTaskKey()
	if jobName == "" {
		jobName = getFunctionName(job)
	}
	printLog("添加%s任务", jobName)
	task := &selfTask{
		Job:     job,
		jobName: jobName,
		key:     taskKey,
	}
	if delay > 0 {
		task.expiredTime = expiredTime(delay)
	} else if crontab != "" {
		task.crontab = &Crontab{Spec: crontab}
	}
	tw.addTask(task)
	return
}

func (tw *TimeWheel) AppendTask(task Task) (taskKey int64, err error) {
	if !tw.running {
		return 0, errors.New("定时器尚未启动，请先调用 Start 启动定时器")
	}
	taskKey = tw.randomTaskKey()
	task.SetTaskKey(taskKey)
	tw.addTask(task)
	return
}

// 统一处理回调函数，如果想在执行回调函数的时候做什么事情，就在这修改。会出现多线程同时调用这个方法。需要考虑线程安全
func (tw *TimeWheel) addTask(task Task) {
	tw.addTaskChannel <- task
}

// RemoveTask 删除指定的回调任务
func (tw *TimeWheel) RemoveTask(taskKey int64) {
	if !tw.taskKeySet.Contains(taskKey) {
		// taskKey 不存在
		return
	}
	tw.removeTaskChannel <- taskKey
}

// 获取随机数字
func (tw *TimeWheel) randomTaskKey() (key int64) {
	rand.Seed(time.Now().Unix())
	for {
		key = rand.Int63()
		if !tw.taskKeySet.Contains(key) {
			return key
		}
	}

}

// 解析 AppendOnceFunc 传入的 expiredTime 参数
func (tw *TimeWheel) expiredTimeParsing(timeParams interface{}) (int64, error) {
	if timeInt, intOk := timeParams.(int); intOk {
		return int64(timeInt), nil
	} else if timeInt64, IntOk := timeParams.(int64); IntOk {
		return timeInt64, nil
	} else if timeStr, StrOk := timeParams.(string); StrOk {
		stamp, err := time.ParseInLocation("2006-01-02 15:04:05", timeStr, time.Local)
		if err != nil {
			return 0, err
		}
		return stamp.Unix(), nil
	}
	return 0, errors.New("过期时间类型错误,目前只支持int,int64,string类型")
}

// 获取当前定时器时间 集合
func (tw *TimeWheel) getTime() time.Time {
	return tw.clock.getNowTime()
}

// PrintTime 获取当前定时器时间 字符串
func (tw *TimeWheel) PrintTime() string {
	return tw.clock.getNowTime().Format("2006-01-02 15:04:05")
}

type ManageTask struct {
	Name          string
	TaskId        int64
	ScheduleTable Schedule
	JobName       string
	FuncName      string
	ExecNumber    int64
	LastRunTime   *time.Time
	NextRunTime   time.Time
}

func (m *ManageTask) MarshalJSON() ([]byte, error) {
	var scheduleTable, lastRunTime string
	switch st := m.ScheduleTable.(type) {
	case *expiredTime:
		scheduleTable = fmt.Sprintf("固定延时：%d秒", st)
	case expiredTime:
		scheduleTable = fmt.Sprintf("固定延时：%d秒", st)
	case *Crontab:
		scheduleTable = st.String()
	case *specSchedule:
		scheduleTable = st.ToString()
	case *ConstantDelaySchedule:
		scheduleTable = st.ToString()
	default:
		scheduleTable = "未知的类型"
	}
	if m.LastRunTime != nil {
		lastRunTime = m.LastRunTime.Format("2006-01-02 15:04:05")
	} else {
		lastRunTime = "未执行"
	}
	result := []byte(fmt.Sprintf(`{"name":"%s","taskId":%d,"scheduleTable":"%s","jobName":"%s","execNumber":%d,"lastRunTime":"%s","nextRunTime":"%s"}`, m.Name, m.TaskId, scheduleTable, m.JobName, m.ExecNumber, lastRunTime, m.NextRunTime.Format("2006-01-02 15:04:05")))
	return result, nil
}

func (tw *TimeWheel) GetAllTask() []ManageTask {
	var result []ManageTask
	result = make([]ManageTask, 0)
	result = append(result, getRouletteTaskInfo(tw.year)...)
	result = append(result, getRouletteTaskInfo(tw.month)...)
	result = append(result, getRouletteTaskInfo(tw.day)...)
	result = append(result, getRouletteTaskInfo(tw.hour)...)
	result = append(result, getRouletteTaskInfo(tw.minute)...)
	result = append(result, getRouletteTaskInfo(tw.second)...)
	return nil
}

func getRouletteTaskInfo(r *roulette) (result []ManageTask) {
	var next *list.Element
	result = make([]ManageTask, 0)
	for i := 0; i < len(r.slots); i++ {
		if r.slots[i] == nil {
			continue
		}
		for e := r.slots[i].Front(); e != nil; e = next {
			next = e.Next()
			task := e.Value.(Task)
			result = append(result, ManageTask{
				Name:          task.GetJobName(),
				TaskId:        task.GetTaskKey(),
				ScheduleTable: task.GetSchedule(),
				JobName:       task.GetJobName(),
				FuncName:      getFunctionName(task),
				ExecNumber:    task.GetExecNumber(),
				LastRunTime:   task.GetLastRunTime(),
				NextRunTime:   task.GetSchedule().NextRunTime(r.clock.getNowTime()),
			})
		}
	}
	return result
}
