package timeWheel

import (
	"container/list"
)

var (
	initYear     int64
	timeIndexMap map[string]int
)

// TaskData 回调函数参数类型
/*
时分秒都是 0-59 一共60个刻度
年 无上限先给个 10年刻度
月 1 - 12 12个刻度
日 1 - 31 不等
*/
type roulette struct {
	name           string
	slots          []*list.List  // 时间轮槽
	slotNum        int           // 槽数量
	currentPos     int           // 当前指针指向哪一个槽
	isLastRoulette bool          // 最底层的时间轮盘,既最小刻度轮盘
	taskKeyMap     map[int64]int // 任务在第几个槽中保存，只保存存在的，查不到就默认不存在
	beforeRoulette *roulette     // 上层的轮盘
	afterRoulette  *roulette     // 下层轮盘
}

// 当前时间是否已经到达上限。
// 到达上限后需要在上一个时间刻度+1
func (r *roulette) cycle() bool {
	/*
		年 不考虑
		月 12个月 最小数字 1 最大数字 12
		日 每个月天数不等 roulette.getMonthDay 这个方法获取 具体天数
		时 24个刻度 最小数字 0 最大数字 23
		分 60个刻度 最小数字 0 最大数字 59
		秒 60个刻度 最小数字 0 最大数字 59
	*/
	if r.name == "month" {
		return r.currentPos == r.slotNum+1
	}
	if r.name == "day" {
		return r.currentPos == getMonthDay(r.beforeRoulette.beforeRoulette.currentPos, r.beforeRoulette.currentPos)+1
	}

	return r.currentPos == r.slotNum
}

// 主要函数入口
/*
定时调用该方法，从最底层的轮盘开始调用，、
相当于表盘中的秒针，每次都移动秒针，当秒针走完一圈就会带动分针走一格，依次类推
指针指向的格子中，所有的任务全部超时
todo 当添加时间点就是当前时间点的时候，而且需要再次向下分配的情况。就会出现一个情况不会触发分配，只能等到循环一次之后才会被调用
todo 已解决，原因在于list.list 在循环的时候需要删除循环出来的对象导致的，需要在删除前把下一个循环对象赋值给一个中间变量
*/
func (r *roulette) tickHandler() {
	r.currentPos++
	if r.cycle() {
		// 刻度归零，先让上层的时间轮指针动起来，如果有分配task的情况，先分配到下层时间轮，这样如果有零点触发的任务就会执行了
		if r.name == "month" || r.name == "day" {
			r.currentPos = 1
		} else {
			r.currentPos = 0
		}

		if r.beforeRoulette != nil {
			r.beforeRoulette.tickHandler()
		}
	}
	tasks := r.getTaskList(int64(r.currentPos))
	if r.isLastRoulette && tasks != nil {
		// 如果是最底层的时间轮的话就执行所有的任务
		r.runTask(tasks)
	} else {
		// 不是执行任务的时间轮就向下层分配任务
		// 到这只能是秒之前的时间刻度
		r.afterRoulette.assignTask(tasks)
	}

}

// 执行所有已到期的任务
func (r roulette) runTask(taskList *list.List) {
	var next *list.Element
	for e := taskList.Front(); e != nil; e = next {
		next = e.Next()
		task := e.Value.(*task)
		go task.Job()
		delete(r.taskKeyMap, task.key)
		taskList.Remove(e)
	}
}

// 判断任务是否需要立即执行，不需要的话就在本层轮盘中放置任务
func (r *roulette) assignTask(tasks *list.List) {
	if tasks == nil {
		return
	}
	afterName := r.name
	runTaskList := list.New()
	var next *list.Element
	for e := tasks.Front(); e != nil; e = next {
		next = e.Next()
		task := e.Value.(*task)
		index, ok := task.rouletteSite[afterName]
		if !ok {
			//需要立即执行任务
			runTaskList.PushBack(task)
		}
		r.getTaskList(index).PushBack(task)
		//if r.slots[index] == nil {
		//	r.slots[index] = list.New()
		//}
		//r.slots[index].PushBack(task)
		//printLog("%s轮盘重分配一个任务,在%d时候调用%s函数，当前指针在%d", r.name, index, task.jobName, r.currentPos)
		r.taskKeyMap[task.key] = int(index)
		// 删除上层轮盘中的task key 标识
		delete(r.beforeRoulette.taskKeyMap, task.key)
		tasks.Remove(e)
	}
}

// 删除尚未到期的任务
func (r *roulette) removeTask(taskKey int64) {
	taskIndex, ok := r.taskKeyMap[taskKey]
	if !ok {
		if r.isLastRoulette {
			return
		}
		r.afterRoulette.removeTask(taskKey)
	}
	//l := r.slots[taskIndex]
	//if l == nil {
	//	return
	//}
	l := r.getTaskList(int64(taskIndex))
	var next *list.Element
	for e := l.Front(); e != nil; e = next {
		next = e.Next()
		task := e.Value.(*task)
		if task.key == taskKey {
			l.Remove(e)
		}
	}

}

//添加task 根据超时时间计算多久后执行任务 最大时间十年，超过十年会触发panic
func (r *roulette) addTask(task *task) {
	if task.delay == 0 {
		r.addTaskByTimeStamp(task)
	} else {
		r.addTaskByInt(task)
	}

}

//year 2021
//month 8
//day 27
//hour 22
//minute 30
//second 10
//
//300000
//
//300,000 / 60 = 5000 0  位置 10 + 0
//5000 / 60 = 83 20  位置 30 + 20
//83 / 24 = 3 11 位置 22 + 11 -> 9 ↓ 1
//3 / 31 = 0 3 位置 27 + 3 + 1
//
//year 2021
//month 8
//day 31
//hour 9
//minute 50
//second 10
//
//余数是在本级中的位置，商是上层需要的计算的数字
func (r *roulette) addTaskByInt(task *task) {
	slotNum := int64(r.slotNum)
	currentPos := int64(r.currentPos)

	nowRouletteSite := task.delay % slotNum // 余数
	nextCircle := task.delay / slotNum      // 商
	if nowRouletteSite+currentPos >= slotNum {
		// 如果超过当前 最大刻度，在上级时间中加一 本级中计算差值
		nextCircle++
		nowRouletteSite = nowRouletteSite + currentPos - slotNum
	} else {
		nowRouletteSite += currentPos
	}
	task.delay = nextCircle
	task.rouletteSite[r.name] = nowRouletteSite
	if nextCircle != 0 {
		//if r.name == "year" {
		//	panic("无法添加超过十年的任务")
		//}
		r.beforeRoulette.addTaskByInt(task)
		return
	}
	//printLog("%s轮盘添加一个任务,在%d时候调用%s函数，当前指针在%d", r.name, nowRouletteSite, task.jobName, r.currentPos)

	if nowRouletteSite == currentPos && r.name == "second" {
		// 如果执行时间就是当前时间立即调用
		go task.Job()
		return
	}
	r.getTaskList(nowRouletteSite).PushBack(task)
	r.taskKeyMap[task.key] = int(nowRouletteSite)
}

func (r *roulette) addTaskByTimeStamp(task *task) {
	execTimeStamp := task.crontab.beforeRunTime // 目标时间
	year := r.getYearRoulette()
	month := r.getMonthRoulette()
	day := r.getDayRoulette()
	hour := r.getHourRoulette()
	minute := r.getMinuteRoulette()
	second := r.getSecondRoulette()

	if execTimeStamp.year-year.currentPos != 0 {
		year.getTaskList(int64(execTimeStamp.year)).PushBack(task)
		year.taskKeyMap[task.key] = execTimeStamp.year
		return
	}
	if execTimeStamp.month-month.currentPos != 0 {
		month.getTaskList(int64(execTimeStamp.month)).PushBack(task)
		month.taskKeyMap[task.key] = execTimeStamp.month
		return
	}
	if execTimeStamp.day-day.currentPos != 0 {
		day.getTaskList(int64(execTimeStamp.day)).PushBack(task)
		day.taskKeyMap[task.key] = execTimeStamp.day
		return
	}
	if execTimeStamp.hour-hour.currentPos != 0 {
		hour.getTaskList(int64(execTimeStamp.hour)).PushBack(task)
		hour.taskKeyMap[task.key] = execTimeStamp.hour
		return
	}
	if execTimeStamp.minute-minute.currentPos != 0 {
		minute.getTaskList(int64(execTimeStamp.minute)).PushBack(task)
		minute.taskKeyMap[task.key] = execTimeStamp.minute
		return
	}
	if execTimeStamp.second-second.currentPos != 0 {
		second.getTaskList(int64(execTimeStamp.second)).PushBack(task)
		second.taskKeyMap[task.key] = execTimeStamp.second
		return
	}
	// 到这里的话那时间就是相同的立即执行该函数
	go task.Job()
}

/*
2021 10 17 12 42 10

2021 10 17 15 00 0



*/

func newRoulette(model string, initPointer int) *roulette {
	// 因为 月份和日都是从1开始的，所以最大列表需要比最大的要长一个
	var (
		maxNum         int
		ListLen        int
		isLastRoulette bool
	)
	if model == "year" {
		maxNum = 10
	} else if model == "month" {
		maxNum = 12
	} else if model == "day" {
		maxNum = 31
	} else if model == "hour" {
		maxNum = 24
	} else if model == "minute" {
		maxNum = 60
	} else if model == "second" {
		maxNum = 60
		isLastRoulette = true
	} else if model == "millisecond" {
		maxNum = 100 // 暂时用不上
	} else {
		panic("model类型错误")
	}
	ListLen = maxNum
	if model == "day" || model == "month" || model == "year" {
		ListLen++
	}
	return &roulette{
		name:           model,
		slots:          make([]*list.List, ListLen),
		slotNum:        maxNum,
		currentPos:     initPointer,
		isLastRoulette: isLastRoulette,
		taskKeyMap:     map[int64]int{},
		beforeRoulette: nil,
		afterRoulette:  nil,
	}
}

func (r *roulette) getTaskList(index int64) *list.List {
	if r.name == "year" {
		index = index - initYear
	}
	tasks := r.slots[index]
	if tasks == nil {
		tasks = list.New()
	}
	r.slots[index] = tasks
	return tasks
}

func (r *roulette) getYearRoulette() *roulette {
	if r.beforeRoulette == nil {
		return r
	}
	return r.beforeRoulette.getYearRoulette()
}
func (r *roulette) getMonthRoulette() *roulette {
	if r.name == "month" {
		return r
	}
	index := timeIndexMap[r.name]
	if index > 5 {
		return r.afterRoulette.getMonthRoulette()
	} else {
		return r.beforeRoulette.getMonthRoulette()
	}
}
func (r *roulette) getDayRoulette() *roulette {
	if r.name == "day" {
		return r
	}
	index := timeIndexMap[r.name]
	if index > 4 {
		return r.afterRoulette.getDayRoulette()
	} else {
		return r.beforeRoulette.getDayRoulette()
	}
}
func (r *roulette) getHourRoulette() *roulette {
	if r.name == "hour" {
		return r
	}
	index := timeIndexMap[r.name]
	if index > 3 {
		return r.afterRoulette.getHourRoulette()
	} else {
		return r.beforeRoulette.getHourRoulette()
	}
}
func (r *roulette) getMinuteRoulette() *roulette {
	if r.name == "minute" {
		return r
	}
	index := timeIndexMap[r.name]
	if index > 2 {
		return r.afterRoulette.getMinuteRoulette()
	} else {
		return r.beforeRoulette.getMinuteRoulette()
	}
}
func (r *roulette) getSecondRoulette() *roulette {
	if r.afterRoulette == nil {
		return r
	}
	return r.afterRoulette.getSecondRoulette()
}
