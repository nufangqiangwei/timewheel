package timeWheel

import (
	"fmt"
	"time"
)

// timestamp 时间对象++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
type timestamp struct {
	year   int
	month  int
	day    int
	hour   int
	minute int
	second int
}

type timestampTimeType string

const (
	timestampYear   timestampTimeType = "year"
	timestampMonth  timestampTimeType = "month"
	timestampDay    timestampTimeType = "day"
	timestampHour   timestampTimeType = "hour"
	timestampMinute timestampTimeType = "minute"
	timestampSecond timestampTimeType = "second"
)

// 指定的时间刻度+1
func (t *timestamp) addUp(timeType timestampTimeType) {
	switch timeType {
	case timestampYear:
		t.year++
	case timestampMonth:
		t.month++
		if t.month == 13 {
			t.month = 1
			t.year++
		}
	case timestampDay:
		t.day++
		if t.day == getMonthDay(t.year, t.month)+1 {
			t.day = 1
			t.addUp(timestampMonth)
		}
	case timestampHour:
		t.hour++
		if t.hour == 24 {
			t.hour = 0
			t.addUp(timestampDay)
		}
	case timestampMinute:
		t.minute++
		if t.minute == 60 {
			t.minute = 0
			t.addUp(timestampHour)
		}
	case timestampSecond:
		t.second++
		if t.second == 60 {
			t.second = 0
			t.addUp(timestampMinute)
		}
	default:
		panic("错误类型")
	}
}

// 指定的时间刻度-1
func (t *timestamp) backDown(timeType string) {
	switch timeType {
	case "year":
		t.year--
	case "month":
		t.month--
		if t.month == 0 {
			t.month = 12
			t.year--
		}
	case "day":
		t.day--
		if t.day == 0 {
			t.backDown("month")
			t.day = getMonthDay(t.year, t.month)
		}
	case "hour":
		t.hour--
		if t.hour == -1 {
			t.hour = 23
			t.backDown("day")
		}
	case "minute":
		t.minute--
		if t.minute == -1 {
			t.minute = 59
			t.backDown("hour")
		}
	case "second":
		t.second--
		if t.second == -1 {
			t.second = 59
			t.backDown("minute")
		}
	default:
		panic("错误类型")
	}
}

// 获取时间差
func (t timestamp) stamp(datetime timestamp) int64 {
	// t 目标天数
	// 当前时间
	// 计算datetime时间到t 时间相差的秒数
	// 只能计算t比datetime大的时间
	days := -1
	stamp := 0
	// year ==============================
	a := datetime.year
	//if t.year+t.month+t.day < datetime.year+datetime.month+datetime.day {
	//	a = t.year
	//	b = datetime.year
	//}
	for {
		if a >= t.year {
			break
		}
		a++
		if leapYear(a) {
			days += 366
		} else {
			days += 365
		}
	}
	// t ==================================
	a = 1
	tDay := 0
	for {
		if a == t.month {
			tDay += t.day
			break
		}
		tDay += getMonthDay(t.year, a)
		a++
	}
	// datetime ===========================
	a = 1
	datetimeDay := 0
	for {
		if a == datetime.month {
			datetimeDay += datetime.day
			break
		}
		datetimeDay += getMonthDay(t.year, a)
		a++
	}
	days += tDay - datetimeDay
	if leapYear(datetime.year) {
		days++
	}
	stamp += days * (24 * 60 * 60)
	a = t.hour*3600 + t.minute*60 + t.second
	b := datetime.hour*3600 + datetime.minute*60 + datetime.second
	stamp += a - b
	x := a - b
	if a < b {
		// 如果t的时分秒比datetime的时分秒小，那需要加一天
		stamp += 86400
		x += 86400
	}
	return int64(stamp)
}

// 使用时间戳获取时间差
func (t timestamp) sysStamp(datetime timestamp) int64 {
	datetimeStamp, _ := time.ParseInLocation("2006-01-02 15:04:05", datetime.PrintTime(), time.Local)
	selfStamp, _ := time.ParseInLocation("2006-01-02 15:04:05", t.PrintTime(), time.Local)

	return selfStamp.Unix() - datetimeStamp.Unix()
}

// 格式化时间字符串 2020-01-01 12:00:00
func (t timestamp) PrintTime() string {
	return fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", t.year, t.month, t.day, t.hour, t.minute, t.second)
}

// 传入的时间是否是将来的时间，就是判断传入的时间比当前时间大
func (t timestamp) isFutureTime(ti timestamp) bool {
	// 如果时间相同那就返回 false
	if t.year == ti.year && t.month == ti.month && t.day == ti.day && t.hour == ti.hour && t.minute == ti.minute && t.second == ti.second {
		return false
	}
	// 从年月日时分秒开始，有一个比现在的大那就是未来的时间
	if t.year < ti.year && t.month < ti.month && t.day < ti.day && t.hour < ti.hour && t.minute < ti.minute && t.second < ti.second {
		return true
	}
	tTimeStamp, _ := time.Parse("2006-01-02 15:04:05", t.PrintTime())
	tiTimeStamp, _ := time.Parse("2006-01-02 15:04:05", ti.PrintTime())
	return tTimeStamp.Before(tiTimeStamp)
}

// 在当前时间上前进多少秒
func (t *timestamp) addTime(second int64) {
	for second > 0 {
		t.addUp(timestampSecond)
		second--
	}
}
