package timeWheel

import (
	"container/list"
	"time"
)

type timeLevel int

const (
	errorLevel timeLevel = iota - 2
	zero
	year
	month
	day
	hour
	minute
	second
	millisecond
)

var (
	initYear        int
	assignTaskIndex = []timeLevel{year, month, day, hour, minute, second, millisecond}
)

// TaskData 回调函数参数类型
/*
时分秒都是 0-59 一共60个刻度
年 无上限先给个 10年刻度
月 1 - 12 12个刻度
日 1 - 31 不等
*/
type roulette struct {
	name           timeLevel     // 分别按照年月日时分秒来命名
	slots          []*list.List  // 时间轮槽
	slotNum        int           // 槽数量
	currentPos     int           // 当前指针指向哪一个槽
	taskKeyMap     map[int64]int // 任务在第几个槽中保存，只保存存在的，查不到就默认不存在
	beforeRoulette *roulette     // 上层的轮盘
	afterRoulette  *roulette     // 下层轮盘
	clock          *clock        // 时钟
}

func newRoulette(model timeLevel, initPointer int) *roulette {
	// 因为 月份和日都是从1开始的，所以最大列表需要比最大的要长一个
	var (
		maxNum  int
		ListLen int
	)
	if model == year {
		maxNum = 11
	} else if model == month {
		maxNum = 12
	} else if model == day {
		maxNum = 31
	} else if model == hour {
		maxNum = 24
	} else if model == minute {
		maxNum = 60
	} else if model == second {
		maxNum = 60
	} else if model == millisecond {
		maxNum = 100 // 暂时用不上
	} else {
		panic("model类型错误")
	}
	ListLen = maxNum
	if model == day || model == month || model == year {
		ListLen++
	}
	return &roulette{
		name:           model,
		slots:          make([]*list.List, ListLen),
		slotNum:        maxNum,
		currentPos:     initPointer,
		taskKeyMap:     map[int64]int{},
		beforeRoulette: nil,
		afterRoulette:  nil,
	}
}

// 当前轮盘时间是否已经到达上限。
// 到达上限后需要在上一个时间刻度+1
func (r *roulette) cycle() bool {
	/*
		年 不考虑
		月 12个月 最小数字 1 最大数字 12
		日 最小数字 1 每个月天数不等 roulette.getMonthDay 这个方法获取 具体天数
		时 24个刻度 最小数字 0 最大数字 23
		分 60个刻度 最小数字 0 最大数字 59
		秒 60个刻度 最小数字 0 最大数字 59
	*/
	if r.name == day {
		return r.currentPos == getMonthDay(r.beforeRoulette.beforeRoulette.currentPos, r.beforeRoulette.currentPos)
	}
	return r.currentPos == r.slotNum
}

// 主要函数入口
/*
定时调用该方法，从最底层的轮盘开始调用，、
相当于表盘中的秒针，每次都移动秒针，当秒针走完一圈就会带动分针走一格，依次类推
指针指向的格子中，所有的任务全部超时
当添加时间点就是当前时间点的时候，而且需要再次向下分配的情况。就会出现一个情况不会触发分配，只能等到循环一次之后才会被调用
已解决，原因在于list.list 在循环的时候需要删除循环出来的对象导致的，需要在删除前把下一个循环对象赋值给一个中间变量
*/
func (r *roulette) tickHandler() {
	r.currentPos++
	if r.cycle() {
		// 刻度归零，先让上层的时间轮指针动起来，如果有分配task的情况，先分配到下层时间轮，这样如果有零点触发的任务就会执行了
		if r.name == month || r.name == day {
			r.currentPos = 1
		} else {
			r.currentPos = 0
		}

		if r.beforeRoulette != nil {
			r.beforeRoulette.tickHandler()
		}
	}
	tasks := r.getTaskList(r.currentPos)
	if r.name == second && tasks != nil {
		// 如果是最底层的时间轮的话就执行所有的任务
		r.runTasks(tasks)
	} else {
		// 不是执行任务的时间轮就向下层分配任务
		// 到这只能是秒之前的时间刻度
		r.afterRoulette.assignTask(tasks)
	}

}

// 执行所有已到期的任务
func (r *roulette) runTasks(taskList *list.List) {
	var next *list.Element
	for e := taskList.Front(); e != nil; e = next {
		next = e.Next()
		task := e.Value.(Task)
		r.runTask(task)
		taskList.Remove(e)
	}
}

// 判断任务是否需要立即执行，不需要的话就在本层轮盘中放置任务
func (r *roulette) assignTask(tasks *list.List) {
	if tasks == nil {
		return
	}
	nowTime := r.clock.getNowTime()
	var (
		next        *list.Element
		nextRunTime time.Time
	)
	for e := tasks.Front(); e != nil; e = next {
		next = e.Next()
		task := e.Value.(Task)
		if task.GetLastRunTime() != nil {
			nextRunTime = task.GetSchedule().NextRunTime(*(task.GetLastRunTime()))
		} else {
			nextRunTime = task.GetSchedule().NextRunTime(nowTime)
		}
		if r.addTask(task, nextRunTime) {
			delete(r.taskKeyMap, task.GetTaskKey())
			tasks.Remove(e)
		}
	}
}

// 添加任务，如果在已经是秒这一层了就直接执行，并返回是否执行的状态
func (r *roulette) addTask(task Task, nextRunTime time.Time) bool {
	futurePointer := getCurrentRoulettePointer(r.name, nextRunTime)
	difference := futurePointer - r.currentPos
	//println(r.name, "添加任务，预计执行时间为：", nextRunTime.Format("2006-01-02 15:04:05"), "当前轮盘指针为：", r.currentPos, "预计轮盘指针为：", futurePointer, "差值为：", difference)
	if difference > 0 {
		r.getTaskList(futurePointer).PushBack(task)
	} else if r.name == second {
		r.runTask(task)
		return true
	} else {
		return r.afterRoulette.addTask(task, nextRunTime)
	}
	return false
}

// 删除尚未到期的任务
func (r *roulette) removeTask(taskKey int64) {
	taskIndex, ok := r.taskKeyMap[taskKey]
	if !ok {
		if r.name != second {
			r.afterRoulette.removeTask(taskKey)
		}
		return
	}

	l := r.getTaskList(taskIndex)
	var next *list.Element
	for e := l.Front(); e != nil; e = next {
		next = e.Next()
		task := e.Value.(Task)
		if task.GetTaskKey() == taskKey {
			l.Remove(e)
		}
	}

}

func (r *roulette) getTaskList(index int) *list.List {
	if r.name == year {
		index = index - initYear
		if index > 10 {
			index = 11
		}
	}
	tasks := r.slots[index]
	if tasks == nil {
		tasks = list.New()
	}
	r.slots[index] = tasks
	return tasks
}

func (r *roulette) runTask(task Task) {
	go func() {
		defer func() {
			panicErr := recover()
			if panicErr != nil {
				_, ok := panicErr.(TimeOut)
				if ok {
					printLog("%s 任务所有指定时间全部执行完毕，任务不在执行", task.GetJobName())
					return
				}
				printLog("%s 执行%s函数出错。错误信息为：%s", time.Now().Format("2006-01-02 15:04:05"), task.GetJobName(), panicErr)
			}
		}()
		printLog("执行 %s 函数", task.GetJobName())
		task.RunJob()
		if task.GetSchedule() != nil {
			r.clock.wheel.addTask(task)
		}
	}()
	delete(r.taskKeyMap, task.GetTaskKey())
}

func getCurrentRoulettePointer(level timeLevel, notTime time.Time) int {
	if level == year {
		return notTime.Year()
	} else if level == month {
		return int(notTime.Month())
	} else if level == day {
		return notTime.Day()
	} else if level == hour {
		return notTime.Hour()
	} else if level == minute {
		return notTime.Minute()
	} else if level == second {
		return notTime.Second()
	}
	return 0
}
