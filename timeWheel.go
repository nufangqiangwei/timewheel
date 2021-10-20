package timeWheel

import (
	"errors"
	"fmt"
	"github.com/deckarep/golang-set"
	"log"
	"math/rand"
	"time"
)

func init() {
	timeIndexMap = map[string]int{
		"year":   6,
		"month":  5,
		"day":    4,
		"hour":   3,
		"minute": 2,
		"second": 1,
	}
	timeList = []string{"year", "month", "day", "hour", "minute", "second"}
}

// task 延时任务
type task struct {
	delay        int64            // 延迟时间
	rouletteSite map[string]int64 // 在每个时间轮的位置
	key          int64            // 定时器唯一标识, 用于删除定时器
	Job          func()           // Job 延时任务回调函数
	crontab      *Crontab         // 重复调用时间表
	jobName      string           // 任务标记。打印给用户看的
}

type Task struct {
	Job     func(interface{}) // 需要执行的任务
	JobData interface{}       // 任务参数
	Repeat  bool              // 是否需要重复执行 false 单次任务 true 重复任务
	Crontab Crontab           // 计划执行时间
	JobName string            // 任务名称
}

// TimeWheel 时间轮
type TimeWheel struct {
	interval          time.Duration // 指针每隔多久往前移动一格
	ticker            *time.Ticker  // 时间间隔
	wheel             *roulette     // 时间轮
	rootWheel         *roulette     // 最上层时间轮
	taskKeySet        mapset.Set    //taskKey集合
	addTaskChannel    chan task     // 新增任务channel
	removeTaskChannel chan int64    // 删除任务channel
	stopChannel       chan bool     // 停止定时器channel
	running           bool
}

// 配置信息
type WheelConfig struct {
	Model        string      // 最小时间单位
	tickInterval int64       // 每次移动指针的时差
	BeatSchedule []Task      // 任务
	IsRun        bool        // 是否直接启动，默认需要手动调用 TimeWheel.Start
	Log          *log.Logger // 打印日志
}

// NewTimeWheel 调用实例，需要全局唯一，
// model: 模式，就是时间轮层数 年月日时分秒 year, month, day, hour, minute, second
// tickInterval:每次转动的时间间隔
// 使用方法
//		tw := NewTimeWheel(&WheelConfig{})
//		_ = tw.AppendOnceFunc(oneCallback, 1, 10)
//		err := tw.AppendCycleFunc(callbackFunc, 2, Crontab{
//			Second: "/5",
//		})
//		if err != nil {
//			tw.Stop()
//			println(err.Error())
//			return
//		}
// 工作大致说明
// TimeWheel.Start() 开始入口 ，通过监听*time.Ticker 每秒执行一次 TimeWheel.wheel.tickHandler() 这个方法
// 该方法每次执行都会在时间上 +1秒 ，每一个时间指针都指向一个list.List 链表，链表内存有 task 对象，被指针指到的链表，其内部所有的 task 都到了
// 执行时间，
func NewTimeWheel(config *WheelConfig) *TimeWheel {
	var (
		rootRoulette *roulette // 根节点
		snapRoulette *roulette // 当前
		lastRoulette *roulette // 上一个轮盘
	)
	config = DefaultWheelConfig(config)
	tw := &TimeWheel{
		interval:          time.Duration(config.tickInterval),
		addTaskChannel:    make(chan task, 10),
		removeTaskChannel: make(chan int64),
		stopChannel:       make(chan bool),
		taskKeySet:        mapset.NewSet(),
	}

	logObject = config.Log
	//ti := time.Now()
	//timeMap := map[string]int{
	//	"year":   ti.Year(),
	//	"month":  int(ti.Month()),
	//	"day":    ti.Day(),
	//	"hour":   ti.Hour(),
	//	"minute": ti.Minute(),
	//	"second": ti.Second(),
	//}
	timeMap := map[string]int{
		"year":   2021,
		"month":  9,
		"day":    10,
		"hour":   11,
		"minute": 59,
		"second": 59,
	}
	initYear = int64(timeMap["year"])
	for _, defaultModel := range timeList {
		snapRoulette = newRoulette(defaultModel, timeMap[defaultModel])
		if defaultModel == config.Model {
			lastRoulette.afterRoulette = snapRoulette
			snapRoulette.beforeRoulette = lastRoulette
			break
		}
		if defaultModel == "year" {
			rootRoulette = snapRoulette
			lastRoulette = snapRoulette
			continue
		}
		lastRoulette.afterRoulette = snapRoulette
		snapRoulette.beforeRoulette = lastRoulette
		lastRoulette = snapRoulette
	}
	tw.wheel = snapRoulette
	tw.rootWheel = rootRoulette

	if config.IsRun {
		go tw.Start()
	}
	if config.BeatSchedule != nil {
		go func() {
			var err error
			for _, schedule := range config.BeatSchedule {
				if schedule.Repeat {
					_, err = tw.AppendCycleFunc(schedule.Job, schedule.JobData, schedule.JobName, schedule.Crontab)
				} else {
					_, err = tw.AppendOnceFunc(schedule.Job, schedule.JobData, schedule.JobName, schedule.Crontab)
				}
				if err != nil {
					printLog("%s 任务添加出错 %s", schedule.JobName, err.Error())
				}
			}
		}()

	}
	return tw
}

func DefaultWheelConfig(config *WheelConfig) *WheelConfig {
	if config == nil {
		config = &WheelConfig{}
	}
	if config.Model == "" {
		config.Model = "second"
	}
	if config.tickInterval == 0 {
		config.tickInterval = int64(time.Millisecond)
		//config.tickInterval = int64(time.Second)
	}
	return config
}

// 开始
func (tw *TimeWheel) Start() {
	if tw.running {
		printLog("已启动，无需再次启动")
		return
	}
	printLog("定时器启动,当前时间 %s", tw.PrintTime())
	tw.running = true
	tw.ticker = time.NewTicker(tw.interval)
	for {
		select {
		case <-tw.ticker.C:
			tw.wheel.tickHandler()
		case task := <-tw.addTaskChannel:
			//printLog("收到一个任务id为 %d 的任务，在 %s 调用 当前时间的 %d 秒后调用", task.key, task.crontab.beforeRunTime.PrintTime(), task.delay)
			tw.wheel.addTask(&task)
		case key := <-tw.removeTaskChannel:
			tw.rootWheel.removeTask(key)
		case <-tw.stopChannel:
			tw.ticker.Stop()
			tw.running = false
			return
		}
	}
}

// 停止
func (tw *TimeWheel) Stop() {
	tw.stopChannel <- true
}

// 添加单次任务
// job 回调函数
// jobData 回调函数的调用参数
// jobName 任务标记，打印出给用户查看
// expiredTime Crontab 对象
func (tw *TimeWheel) AppendOnceFunc(job func(interface{}), jobData interface{}, jobName string, expiredTime Crontab) (taskKey int64, err error) {
	//if !tw.running {
	//	return 0, errors.New("定时器尚未启动，请先调用 Start 启动定时器")
	//}
	timeParams := expiredTime.getNextExecTime(tw.getTimeDict())
	if timeParams > 10*365*24*60*60 {
		return 0, errors.New("时间最长不能超过十年")
	}
	taskKey = tw.randomTaskKey()
	if jobName == "" {
		jobName = getFunctionName(job)
	}
	tw.addTask(job, jobData, &expiredTime, jobName, taskKey, false)
	return
}

// 添加重复任务
// job 回调函数
// jobData 回调函数的调用参数
// jobName 任务标记，打印出给用户查看
// expiredTime Crontab 对象
func (tw *TimeWheel) AppendCycleFunc(job func(interface{}), jobData interface{}, jobName string, expiredTime Crontab) (taskKey int64, err error) {
	//if !tw.running {
	//	return 0, errors.New("定时器尚未启动，请先调用 Start 启动定时器")
	//}
	timeParams := expiredTime.getNextExecTime(tw.getTimeDict())
	//fmt.Printf("重复任务下次执行时间: %d\n", timeParams)

	if timeParams > 10*365*24*60*60 {
		return 0, errors.New("时间最长不能超过十年")
	}
	taskKey = tw.randomTaskKey()
	if jobName == "" {
		jobName = getFunctionName(job)
	}
	printLog("%s 初次添加%s任务", expiredTime.beforeRunTime.PrintTime(), jobName)
	tw.addTask(job, jobData, &expiredTime, jobName, taskKey, true)
	return
}

// 统一处理回调函数，如果想在执行回调函数的时候做什么事情，就在这修改
func (tw *TimeWheel) addTask(job func(interface{}), jobData interface{}, crontab *Crontab, jobName string, taskKey int64, isCycle bool) {
	var taskJob func()
	if isCycle {
		taskJob = func() {
			defer func() {
				panicErr := recover()
				if panicErr != nil {
					_, ok := panicErr.(TimeOut)
					if ok {
						printLog("%s 任务所有指定时间全部执行完毕，任务不在执行", jobName)
						return
					}
					printLog("%s 执行%s函数出错。错误信息为：%s", time.Now().Format("2006-01-02 15:04:05"), jobName, panicErr)
				}
			}()
			printLog("执行 %s 函数", jobName)
			//printLog("执行 %s 函数,定时器时间是%s ", jobName, tw.PrintTime())
			crontab.getNextExecTime(tw.getTimeDict())

			tw.addTask(job, jobData, crontab, jobName, taskKey, true)
			job(jobData)
			printLog("%s 任务执行完成。预计在 % s再次调用：%s", jobName, crontab.beforeRunTime.PrintTime(), jobName)
		}
	} else {
		taskJob = func() {
			defer func() {
				panicErr := recover()
				if panicErr != nil {
					_, ok := panicErr.(TimeOut)
					if ok {
						printLog("%s 任务所有指定时间全部执行完毕，任务不在执行", jobName)
						return
					}
					printLog("%s 执行%s函数出错。错误信息为：%s", time.Now().Format("2006-01-02 15:04:05"), jobName, panicErr)
				}
			}()
			printLog("执行 %s 函数", jobName)
			tw.taskKeySet.Remove(taskKey)
			job(jobData)
		}
	}
	tw.taskKeySet.Add(taskKey)
	tw.addTaskChannel <- task{
		delay: crontab.ExpiredTime,
		rouletteSite: map[string]int64{
			"year":   int64(crontab.beforeRunTime.year),
			"month":  int64(crontab.beforeRunTime.month),
			"day":    int64(crontab.beforeRunTime.day),
			"hour":   int64(crontab.beforeRunTime.hour),
			"minute": int64(crontab.beforeRunTime.minute),
			"second": int64(crontab.beforeRunTime.second),
		},
		key:     taskKey,
		Job:     taskJob,
		crontab: crontab,
		jobName: jobName,
	}
}

// 删除指定的回调任务
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
func (tw *TimeWheel) getTime() (year, month, day, hour, minute, second int) {
	year = tw.rootWheel.getYearRoulette().currentPos
	month = tw.rootWheel.getMonthRoulette().currentPos
	day = tw.rootWheel.getDayRoulette().currentPos
	hour = tw.rootWheel.getHourRoulette().currentPos
	minute = tw.rootWheel.getMinuteRoulette().currentPos
	second = tw.rootWheel.getSecondRoulette().currentPos
	return
}

//  获取当前定时器时间 字符串
func (tw *TimeWheel) PrintTime() string {
	year, month, day, hour, minute, second := tw.getTime()
	return fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", year, month, day, hour, minute, second)
}

// 获取当前定时器时间 timestamp 对象
func (tw *TimeWheel) getTimeDict() (result timestamp) {
	result = timestamp{}
	result.year, result.month, result.day, result.hour, result.minute, result.second = tw.getTime()
	return result
}
