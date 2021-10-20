package timeWheel

// Crontab 时间执行表 +++++++++++++++++++++++++++++++++++++++++++++++++++
// 字符串 按照给定的数字，当时间到给定的刻度就会执行
// 比如 Crontab{Minute:10,Second:30} 每个小时的十分三十秒的时候就会执行
// 支持一次传入多个时间点 Crontab{Minute:"10,11,12",Second:30}每个小时的10：30，11：30，12：30三个时间点执行
// 连续时间点可以用-表示 Minute:"10-12" 代表 Minute:"10,11,12"
// 也可以 /5表示当时间点可以被5整除的时候就执行任务 前面可以写自己指定的时间段 默认的是当前时间段的起止 比如 秒 就是0-59
// 10-20/5 表示当时间点在 10 15 20 这三个时间点执行任务
// 参照python的celery的crontab实现的
//
// 也可以直接传一个间隔时间，会以当前的时间点为起点，向后推迟到目标时间点执行任务
type Crontab struct {
	ExpiredTime   int64 // 用户直接指定延迟时间
	Second        string
	Minute        string
	Hour          string
	Day           string
	Month         string
	Year          string
	second        []int
	minute        []int
	hour          []int
	day           []int
	month         []int
	year          []int
	yearIndex     int
	monthIndex    int
	dayIndex      int
	hourIndex     int
	minuteIndex   int
	secondIndex   int
	timeDict      map[string][]int
	init          bool
	beforeRunTime timestamp //调用时间
	isSetting     bool
}

//初始化配置信息
func (c *Crontab) initConfig() {

	var (
		abc []int
		i   int
	)

	c.init = true
	settingExpiredTime := c.ExpiredTime != 0
	settingTimeStamp := c.Second != "" || c.Minute != "" || c.Hour != "" || c.Day != "" || c.Month != "" || c.Year != ""
	if settingExpiredTime && settingTimeStamp {
		panic("延迟时间和执行时间点，只需要设置一项")
	}
	if !settingExpiredTime && !settingTimeStamp {
		panic("需要设置一项 延迟时间或执行时间点。")
	}
	/*
		如果该时间刻度有指定时间，并且回调函数没有指定，那就在当前刻度就是用户配置的最小时间刻度，回调函数就从这里开始执行
		如果该时间刻度没有指定时间，而且回调函数有指定，那就当前的时间刻度所有值都遍历一边

		如果该时间刻度有指定时间，并且回调函数有指定，那就刻度赋值
		如果时间刻度没有指定时间，而且回调函数未指定，对应的时间刻度赋值
	*/
	if settingTimeStamp {
		abc = splitArgs(c.Second, "second")
		if c.Second != "" {
			c.isSetting = true
		} else {
			c.second = abc
		}

		abc = splitArgs(c.Minute, "minute")
		if c.Minute != "" && !c.isSetting {
			c.isSetting = true
			c.minute = abc
		} else if c.Minute == "" && c.isSetting {
			i = 0
			for i < 60 {
				c.minute = append(c.minute, i)
				i++
			}
		} else {
			c.minute = abc
		}

		abc = splitArgs(c.Hour, "hour")
		if c.Hour != "" && !c.isSetting {
			c.isSetting = true
			c.hour = abc
		} else if c.Hour == "" && c.isSetting {
			i = 1
			for i < 24 {
				c.hour = append(c.hour, i)
				i++
			}
		} else {
			c.hour = abc
		}

		abc = splitArgs(c.Day, "day")
		if c.Day != "" && !c.isSetting {
			c.isSetting = true
			c.day = abc
		} else if c.Day == "" && c.isSetting {
			i = 1
			for i <= 31 {
				c.day = append(c.day, i)
				i++
			}
		} else {
			c.day = abc
		}

		abc = splitArgs(c.Month, "month")
		if c.Month != "" && !c.isSetting {
			c.isSetting = true
			c.month = abc
		} else if c.Month == "" && c.isSetting {
			i = 1
			for i <= 12 {
				c.month = append(c.month, i)
				i++
			}
		} else {
			c.month = abc
		}

		abc = splitArgs(c.Year, "year")
		if c.Year != "" && !c.isSetting {
			c.isSetting = true
			c.year = abc
		} else if c.Year == "" && c.isSetting {
			i = int(initYear)
			maxYear := int(initYear) + 10
			for i <= maxYear {
				c.year = append(c.year, i)
				i++
			}
		} else {
			c.year = abc
		}

		c.beforeRunTime = timestamp{
			year:   c.year[0],
			month:  c.month[0],
			day:    c.day[0],
			hour:   c.hour[0],
			minute: c.minute[0],
			second: c.second[0],
		}
	}

}

// 获取执行延时
func (c *Crontab) getNextExecTime(TimeDict timestamp) int64 {
	if !c.init && c.ExpiredTime == 0 {
		// 初始化 解析字符串
		c.initConfig()
		for !TimeDict.isFutureTime(c.beforeRunTime) {
			c.recentTime()
		}
		return 0
	}
	// 如果是直接指定了延迟时间，那就返回指定的值
	if c.ExpiredTime != 0 {
		return c.ExpiredTime
	}

	// 生成下一个执行时间点
	//c.beforeRunTime.Second(c)
	for !TimeDict.isFutureTime(c.beforeRunTime) {
		c.recentTime()
	}

	return 0
}

// 根据数组计算下一个时间
func (c *Crontab) recentTime() {
	t := &(c.beforeRunTime)
	if c.secondIndex >= len(c.second)-1 {
		c.secondIndex = 0
		t.second = c.second[c.secondIndex]
	} else {
		c.secondIndex++
		t.second = c.second[c.secondIndex]
		return
	}

	if c.minuteIndex >= len(c.minute)-1 {
		c.minuteIndex = 0
		t.minute = c.minute[c.minuteIndex]
	} else {
		c.minuteIndex++
		t.minute = c.minute[c.minuteIndex]
		return
	}

	if c.hourIndex >= len(c.hour)-1 {
		c.hourIndex = 0
		t.hour = c.hour[c.hourIndex]
	} else {
		c.hourIndex++
		t.hour = c.hour[c.hourIndex]
		return
	}

	if c.dayIndex >= len(c.day)-1 {
		c.dayIndex = 0
		t.day = c.day[c.dayIndex]
	} else {
		c.dayIndex++
		t.day = c.day[c.dayIndex]
		return
	}

	if c.monthIndex >= len(c.month)-1 {
		c.monthIndex = 0
		t.month = c.month[c.monthIndex]
	} else {
		c.monthIndex++
		t.month = c.month[c.monthIndex]
		return
	}

	if c.yearIndex >= len(c.year)-1 {
		panic(TimeOut{s: "已到指定的时间尽头"})
	} else {
		c.yearIndex++
		t.year = c.year[c.yearIndex]
		return
	}
	return
}
