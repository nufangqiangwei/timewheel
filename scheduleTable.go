package timeWheel

import (
	"strconv"
	"time"
)

type ZeroScheduleError string

func (z ZeroScheduleError) Error() string {
	return string(z + "任务没有指定时间表，无法执行")
}

type Schedule interface {
	NextRunTime(time.Time) time.Time
	MarshalJSON() ([]byte, error)
}

// expiredTime 用户直接指定延迟时间
type expiredTime int64

func (e *expiredTime) NextRunTime(nowTime time.Time) time.Time {
	if *e > 0 {
		return nowTime.Add(time.Duration(*e) * time.Second)
	}
	// 无延迟
	return time.Date(9999, 12, 31, 0, 0, 0, 0, time.Local)
}
func (e *expiredTime) MarshalJSON() ([]byte, error) {
	return []byte("固定延时：" + strconv.FormatInt(int64(*e), 10) + "秒"), nil
}
