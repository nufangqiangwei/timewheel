package Timewheel

import (
	"errors"
	"fmt"
	"math/rand"
	"time"
)

//回调任务
type job struct {
	appendTime int64             // 添加时候的时间戳毫秒
	outTime    int64             // 超时时间 单位毫秒 就是多久后执行回调
	callback   func(interface{}) // 回调函数
	params     interface{}       //回调函数参数
	key        int64             // 任务唯一id
	// 计算后的超时时间，根据这个值去依次放到指定的位置
	timeIndex map[string]int64
	//Year        int8 //年
	//Month       int8 //月
	//Day         int8 //日
	//Hour        int64 //小时
	//Minute      int64 //分
	//Second      int64 //秒
	//Millisecond int64 //毫秒
}

// 时间转盘
type Wheel struct {
	nextWheel     *Wheel               //下级轮盘
	roulette      int64                //轮盘长度
	roulettePoint int64                //当前指针位置
	rouletteData  [100]*map[int64]*job //当前轮盘任务
	mark          string               //备注
}

//处理逻辑
type TimeWheel struct {
	wheel             *Wheel             // 时间盘
	addTaskChannel    chan *job          // 新增任务channel
	removeTaskChannel chan [5]int64      // 删除任务channel
	stopChannel       chan bool          // 停止定时器channel
	ticker            *time.Ticker       // 定时器
	KeyMap            map[int64][5]int64 //所有的key
}

// 根据所需的量初始化相应数量的map
func newMap(count int) (result [100]*map[int64]*job) {
	for i := 0; i <= count; i++ {
		if i == 100 {
			return
		}
		result[i] = &map[int64]*job{}
	}
	return
}

// 初始化时间转盘
func newWheel() *Wheel {
	currentTime := time.Now()
	year := Wheel{
		nextWheel:     new(Wheel),
		roulette:      100,
		roulettePoint: int64(currentTime.Year()),
		rouletteData:  newMap(100),
		mark:          "Year",
	}
	month := Wheel{
		nextWheel:     &year,
		roulette:      12,
		roulettePoint: int64(currentTime.Month()),
		rouletteData:  newMap(12),
		mark:          "Month",
	}
	day := Wheel{
		nextWheel:     &month,
		roulette:      30,
		roulettePoint: int64(currentTime.Day()),
		rouletteData:  newMap(30),
		mark:          "Day",
	}
	hour := Wheel{
		nextWheel:     &day,
		roulette:      24,
		roulettePoint: int64(currentTime.Hour()),
		rouletteData:  newMap(24),
		mark:          "Hour",
	}
	second := Wheel{
		nextWheel:     &hour,
		roulette:      60,
		roulettePoint: int64(currentTime.Minute()),
		rouletteData:  newMap(60),
		mark:          "Minute",
	}
	minute := Wheel{
		nextWheel:     &second,
		roulette:      60,
		roulettePoint: int64(currentTime.Second()),
		rouletteData:  newMap(60),
		mark:          "Second",
	}
	return &Wheel{
		nextWheel:     &minute,
		roulette:      100,
		roulettePoint: int64(currentTime.Nanosecond()) / 10000000,
		rouletteData:  newMap(100),
		mark:          "Millisecond",
	}
}

// 获取多久后过期 传入时间单位 10毫秒
func getWheelIndex(outTime int64) (Hour int64, Minute int64, Second int64, millisecond int64) {
	var a int64
	Hour = outTime / 360000
	a = outTime % 360000
	Minute = a / 60000
	a = a % 60000
	Second = a / 1000
	millisecond = a % 100
	return
}

// 时间转盘步进
func (wh *Wheel) stepping() (result map[int64]*job) {
	wh.roulettePoint++
	if wh.roulettePoint == wh.roulette {
		if wh.nextWheel != nil {
			for k, v := range wh.nextWheel.stepping() {
				index := v.timeIndex[wh.mark]
				(*wh.rouletteData[index])[k] = v
			}
		}
		wh.roulettePoint = 0
	}

	if wh.rouletteData[wh.roulettePoint] == nil {
		result = map[int64]*job{}
	} else {
		result = *(wh.rouletteData[wh.roulettePoint])
	}
	wh.rouletteData[wh.roulettePoint] = &map[int64]*job{}
	return
}

//获取时间转盘的当前时间
func (wh Wheel) getTime(li *[7]int64, index int) {

	//if wh.nextWheel == nil {
	//	return ""
	//}
	li[index] = wh.roulettePoint
	index++
	if index == 7 {
		return
	}
	wh.nextWheel.getTime(li, index)
	//if after == "" {
	//	return fmt.Sprintf("%v", wh.roulettePoint)
	//}
	//return fmt.Sprintf("%s：%v", after, wh.roulettePoint)

	//wh.nextWheel.getTime()
}

// 删除任务
func (wh *Wheel) removeTask(li [5]int64, index int) {
	if index == 5 {
		return
	}
	if li[index] == 0 {
		index++
		wh.nextWheel.removeTask(li, index)
	} else {
		_, ok := (*wh.rouletteData[li[index]])[li[4]]
		if ok {
			delete(*wh.rouletteData[li[index]], li[4])
		} else {
			index++
			wh.nextWheel.removeTask(li, index)
		}
	}
}

// 开始定时器 需手动使用go执行
func (tw *TimeWheel) Start() {
	//tw.wheel = newWheel()
	for {
		select {
		case <-tw.ticker.C:
			tw.tickHandler()
		case task := <-tw.addTaskChannel:
			tw.addTask(task)
		case key := <-tw.removeTaskChannel:
			tw.wheel.removeTask(key, 0)
		case <-tw.stopChannel:
			tw.ticker.Stop()
			return
		}
	}
}

// 添加任务 outTime 过期时间 单位毫秒 callBack 回调函数 params 回调函数的参数
// 返回 key 改任务的key删除任务的时候使用该参数删除 err 现在无法添加超过一天的任务
func (tw *TimeWheel) AddTask(outTime int64, callBack func(interface{}), params interface{}) (key int64, err error) {
	err = nil
	Hour, Minute, Second, millisecond := getWheelIndex(outTime)
	if Hour > 24 {
		err = errors.New("无法添加超过一天的任务")
		return
	}
	a := make(map[string]int64)
	a["Hour"] = Hour
	a["Minute"] = Minute
	a["Second"] = Second
	a["Millisecond"] = millisecond
	key = rand.Int63()
	for {
		_, ok := tw.KeyMap[key]
		if ok {
			key = rand.Int63()
		} else {
			break
		}
	}
	newJob := job{
		appendTime: time.Now().Unix() / 1e6,
		outTime:    outTime,
		callback:   callBack,
		timeIndex:  a,
		params:     params,
		key:        key,
	}
	tw.addTaskChannel <- &newJob
	return
}

// 删除任务
func (tw TimeWheel) DeleteTask(key int64) error {
	timeList, ok := tw.KeyMap[key]
	if ok {
		tw.removeTaskChannel <- timeList
		return nil
	} else {
		return errors.New("该任务不存在")
	}
}

// 转盘步进一格 并执行超时任务
func (tw TimeWheel) tickHandler() {
	//after := tw.wheel.nextWheel.nextWheel.roulettePoint
	workonFunc := tw.wheel.stepping()
	for _, i := range workonFunc {
		go i.callback(i.params)
		delete(tw.KeyMap, i.key)
	}
	//if after != tw.wheel.nextWheel.nextWheel.roulettePoint {
	//	println(time.Now().Format("2006-01-02 15:04:05"))
	//	println(tw.GetTime())
	//}
}

// 获取当前时间 按指定格式返回 尚未完成这个。目前只能返回固定格式
func (tw TimeWheel) GetTime() string {
	li := &[7]int64{}
	tw.wheel.getTime(li, 0)

	return fmt.Sprintf("%v-%v-%v %v:%v:%v:%v", li[6], li[5], li[4], li[3], li[2], li[1], li[0])

}

// 添加任务
func (tw *TimeWheel) addTask(task *job) {
	//defer func() {
	//	if err := recover(); err != nil {
	//		fmt.Println(err)
	//	}
	//}()
	kMap := &([5]int64{4: task.key})
	wheelMap := [4]*Wheel{
		tw.wheel.nextWheel.nextWheel.nextWheel,
		tw.wheel.nextWheel.nextWheel,
		tw.wheel.nextWheel,
		tw.wheel,
	}
	taskTimeIndex := [4]int64{
		task.timeIndex["Hour"],
		task.timeIndex["Minute"],
		task.timeIndex["Second"],
		task.timeIndex["Millisecond"],
	}
	now := true
	for index, i := range taskTimeIndex {
		if i != 0 {
			now = false
			nowWheel := wheelMap[index]
			timeIndex := nowWheel.roulettePoint + i
			if timeIndex > nowWheel.roulette {
				timeIndex = timeIndex - nowWheel.roulette
			}
			(*nowWheel.rouletteData[timeIndex])[task.key] = task
			fmt.Printf("当前指针位置%d，添加位置%d\n", nowWheel.roulettePoint, timeIndex)
			kMap[index] = timeIndex
			for xx := index + 1; xx < 4; xx++ {
				kMap[xx] = taskTimeIndex[xx]
			}
			break
		}
	}
	if now {
		go task.callback(task.params)
		return
	}
	//if task.timeIndex["Hour"] != 0 {
	//	nowWhell := tw.wheel.nextWheel.nextWheel.nextWheel
	//	index := nowWhell.roulettePoint + task.timeIndex["Hour"]
	//	if index > nowWhell.roulette {
	//		(*nowWhell.rouletteData[index-nowWhell.roulette])[task.key] = task
	//	} else {
	//		(*nowWhell.rouletteData[index])[task.key] = task
	//	}
	//} else if task.timeIndex["Minute"] != 0 {
	//	nowWhell := tw.wheel.nextWheel.nextWheel
	//	index := nowWhell.roulettePoint + task.timeIndex["Minute"]
	//	if index > nowWhell.roulette {
	//		(*nowWhell.rouletteData[index-nowWhell.roulette])[task.key] = task
	//	} else {
	//		(*nowWhell.rouletteData[index])[task.key] = task
	//	}
	//} else if task.timeIndex["Second"] != 0 {
	//	nowWhell := tw.wheel.nextWheel
	//	index := nowWhell.roulettePoint + task.timeIndex["Second"]
	//	if index > nowWhell.roulette {
	//		(*nowWhell.rouletteData[index-nowWhell.roulette])[task.key] = task
	//	} else {
	//		(*nowWhell.rouletteData[index])[task.key] = task
	//	}
	//} else if task.timeIndex["Millisecond"] != 0 {
	//	nowWhell := tw.wheel
	//	index := nowWhell.roulettePoint + task.timeIndex["Millisecond"]
	//	if index > nowWhell.roulette {
	//		(*nowWhell.rouletteData[index-nowWhell.roulette])[task.key] = task
	//	} else {
	//		(*nowWhell.rouletteData[index])[task.key] = task
	//	}
	//} else {
	//	go task.callback(task.params)
	//	return
	//}
	tw.KeyMap[task.key] = *kMap
}

// 获取定时器对象
func NewTimeWheel() *TimeWheel {
	return &TimeWheel{
		//wheel: new(Wheel),
		wheel:             newWheel(),
		addTaskChannel:    make(chan *job, 100),
		removeTaskChannel: make(chan [5]int64, 100),
		stopChannel:       make(chan bool, 100),
		ticker:            time.NewTicker(time.Millisecond * 10),
		KeyMap:            map[int64][5]int64{},
	}
}
