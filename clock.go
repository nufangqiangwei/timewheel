package timeWheel

import "time"

type clock struct {
	year   *roulette
	month  *roulette
	day    *roulette
	hour   *roulette
	minute *roulette
	second *roulette
	loc    *time.Location
	wheel  *TimeWheel
}

func newClock() *clock {
	nowTime := time.Now()
	timetable := &clock{
		year:   newRoulette(year, nowTime.Year()),
		month:  newRoulette(month, int(nowTime.Month())),
		day:    newRoulette(day, nowTime.Day()),
		hour:   newRoulette(hour, nowTime.Hour()),
		minute: newRoulette(minute, nowTime.Minute()),
		second: newRoulette(second, nowTime.Second()),
	}
	// 先就使用本地时区
	timetable.loc = nowTime.Location()
	// 将每个轮盘与上一个轮盘关联
	timetable.second.beforeRoulette = timetable.minute
	timetable.minute.beforeRoulette = timetable.hour
	timetable.hour.beforeRoulette = timetable.day
	timetable.day.beforeRoulette = timetable.month
	timetable.month.beforeRoulette = timetable.year
	// 将每个轮盘与下一个轮盘关联
	timetable.year.afterRoulette = timetable.month
	timetable.month.afterRoulette = timetable.day
	timetable.day.afterRoulette = timetable.hour
	timetable.hour.afterRoulette = timetable.minute
	timetable.minute.afterRoulette = timetable.second
	//
	timetable.year.clock = timetable
	timetable.month.clock = timetable
	timetable.day.clock = timetable
	timetable.hour.clock = timetable
	timetable.minute.clock = timetable
	timetable.second.clock = timetable

	return timetable
}

func (t *clock) resetTime() {
	nowTime := time.Now()
	t.year.currentPos = nowTime.Year()
	t.month.currentPos = int(nowTime.Month())
	t.day.currentPos = nowTime.Day()
	t.hour.currentPos = nowTime.Hour()
	t.minute.currentPos = nowTime.Minute()
	t.second.currentPos = nowTime.Second()
}

func (t *clock) tickHandler() {
	t.second.tickHandler()
}

func (t *clock) getNowTime() time.Time {
	yearPoint := t.year.currentPos
	monthPoint := t.month.currentPos
	dayPoint := t.day.currentPos
	hourPoint := t.hour.currentPos
	minutePoint := t.minute.currentPos
	secondPoint := t.second.currentPos
	return time.Date(yearPoint, time.Month(monthPoint), dayPoint, hourPoint, minutePoint, secondPoint, 0, t.loc)
}

func (t *clock) addTask(task Task) error {
	schedule := task.GetSchedule()
	if schedule == nil {
		return ZeroScheduleError(task.GetJobName())
	}
	var lastRunTime time.Time
	if task.GetLastRunTime() == nil {
		lastRunTime = t.getNowTime()
		task.SetLastRunTime(lastRunTime)
	} else {
		lastRunTime = *(task.GetLastRunTime())
	}

	runTime := schedule.NextRunTime(lastRunTime)
	t.year.addTask(task, runTime)
	return nil
}
