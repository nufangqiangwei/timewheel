package timeWheel

import (
	"fmt"
	"log"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"unicode"
)

var (
	logObject *log.Logger
	timeList  = []string{"year", "month", "day", "hour", "minute", "second"}
)

// 是否是闰年
func leapYear(year int) bool {
	//是否是闰年
	if year%100 == 0 {
		if year%400 == 0 {
			return true
		} else {
			return false
		}
	}
	if year%4 == 0 {
		return true
	} else {
		return false
	}
}

// 获取指定月份的天数
func getMonthDay(year, month int) int {
	if month == 4 || month == 6 || month == 9 || month == 11 {
		// 30天的月份 月份从1开始计数
		return 30
	}
	if month == 1 || month == 3 || month == 5 || month == 7 || month == 8 || month == 10 || month == 12 {
		// 31天的月份 月份从1开始计数
		return 31
	}
	if leapYear(year) {
		return 29
	} else {
		return 28
	}
}

// 切割字符串，根据输入将执行的时间点以切片的形式返回
func splitArgs(args, timeType string) []int {
	/*
		按逗号分割每个时间点，每个时间点可以传入一个整数直接指定时间点。也可以传一个计算式，来计算具体的时间点
		计算式如下
			- 省略号 类似range 从第一个数到最后一个数，全范围
			例：
				0-10 结果时间点是[0,1,2,3,4,5,6,7,8,9]
			/ 除号 将前面的时间点对后面的底取余数，当余数为0 的被除数就是目标时间点
			例子：
				0-10/2 [0,1,2,3,4,5,6,7,8,9]这些数字对2取余数 结果是[0,2,4,6,8]
				/2 == 0 当前面省略的时候，会更具当前时间刻度的范围取全量来计算
	*/

	if args == "" {
		return []int{0}
	}
	s := strings.Split(args, ",")
	var result []int
	for _, timeStr := range s {
		onlyInt := true   // 只有数字，无需计算
		xIndex := -1      // /字符出现的位置
		yIndex := -1      // -字符出现的位置
		var cache []int   // 对每个逗号中内容的缓存
		var special []int // 计算式 例： 0-10 = [0,-45,10]; 0-10/2 = [0,-45,10,-47,2]; /2 = [-47,2]
		for _, a := range timeStr {
			// 判断a字符是否是数字，不是数字，代表这是个计算式
			if !unicode.IsDigit(a) {
				onlyInt = false
				if a == 47 {
					// 当前字符是 / 字符
					if xIndex != -1 {
						// 相同字符只能出现一次
						panic(fmt.Sprintf("Crontab解析执行时间异常：%s字符格式错误%s", timeType, args))
					}
					if len(cache) > 0 {
						// 如果之前传入了数字，需要把之前的数字保存
						special = append(special, listToInt(cache))
					}
					special = append(special, -47)
					cache = cache[:0]
					xIndex = len(special) - 1
				} else if a == 45 {
					// 当前字符是 - 字符
					if yIndex != -1 {
						// 相同字符只能出现一次
						panic(fmt.Sprintf("Crontab解析执行时间异常：%s字符格式错误%s", timeType, args))
					}
					special = append(special, listToInt(cache))
					special = append(special, -45)
					cache = cache[:0]
					yIndex = len(special) - 1
				} else {
					panic(fmt.Sprintf("Crontab只能解析 - / 这两个字符不能解析 %s 字符", string(a)))
				}
				continue
			}
			cache = append(cache, strToInt(a))
		}
		start, end := timeRange(timeType)

		if !onlyInt {
			// timeStr 是一个计算式，需要计算出结果
			// 计算顺序，如果有 - 先计算范围 在计算余数。范围需要更具当前刻度限制
			// 获取当前刻度的起止时间点

			var e, f int
			// 把最后一个数字添加到计算式中
			special = append(special, listToInt(cache))
			if yIndex != -1 {
				// 确定起止点
				if xIndex == 0 {
					e, f = start, end
				} else {
					e, f = special[0], special[2]
					if e < start {
						e = start
					}
					if f > end {
						f = end + 1
					}
				}

				// 如果有计算余数，那就把这个范围放在special中的 / 字符前面
				if xIndex != -1 {
					if xIndex < yIndex {
						panic("Crontab解析异常：余数计算的被除数，只能是整数")
					}
					special[0] = start
					special[2] = end
				} else {
					// 没有余数计算，直接循环一遍
					for e < f {
						result = append(result, e)
						e++
					}
				}
			}

			if xIndex != -1 {
				if xIndex == 0 {
					e, f = start, end+1
				} else {
					e, f = special[0], special[2]
					if e < start {
						e = start
					}
					if f > end {
						f = end + 1
					}
				}

				base := special[xIndex+1]
				// 根据起始时间计算可被底数整除的时间点
				for e < f {
					if e%base == 0 {
						result = append(result, e)
					}
					e++
				}
			}
		} else {
			// 当前逗号内的只是数字，直接输出
			ii := listToInt(cache)
			if start <= ii && ii <= end {
				result = append(result, ii)
			}
		}
	}
	sort.Ints(result)
	return result
}

// rune 映射成数字
func strToInt(a rune) int {
	switch a {
	case 48:
		return 0
	case 49:
		return 1
	case 50:
		return 2
	case 51:
		return 3
	case 52:
		return 4
	case 53:
		return 5
	case 54:
		return 6
	case 55:
		return 7
	case 56:
		return 8
	case 57:
		return 9
	}
	panic("")
}

// 数字切片转整数
func listToInt(data []int) int {
	l := len(data) - 1
	result := 0
	for index, value := range data {
		result += pow(10, l-index) * value
	}
	return result
}

// 计算次方
func pow(x, n int) int {
	ret := 1 // 结果初始为0次方的值，整数0次方为1。如果是矩阵，则为单元矩阵。
	for n != 0 {
		if n%2 != 0 {
			ret = ret * x
		}
		n /= 2
		x = x * x
	}
	return ret
}

// 每个时间轮盘的起止时间点
func timeRange(timeType string) (start, end int) {
	if initYear == 0 {
		panic("请先初始化开始年份")
	}
	switch timeType {
	case "year":
		return int(initYear), int(initYear) + 10
	case "month":
		return 1, 12
	case "day":
		return 1, 31 // 2月份的处理放在计算时间间隔的时候
	case "hour":
		return 0, 23
	case "minute":
		return 0, 59
	case "second":
		return 0, 59
	}
	panic("错误时间类型")
}

// 获取函数名
func getFunctionName(i interface{}, seps ...rune) string {
	// 获取函数名称
	fn := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()

	// 用 seps 进行分割
	fields := strings.FieldsFunc(fn, func(sep rune) bool {
		for _, s := range seps {
			if sep == s {
				return true
			}
		}
		return false
	})

	// fmt.Println(fields)

	if size := len(fields); size > 0 {
		return fields[size-1]
	}
	return ""
}

// 打印日志
func printLog(format string, v ...interface{}) {
	if logObject != nil {
		logObject.Printf(fmt.Sprintf("%s\n", format), v...)
	} else {
		fmt.Printf(fmt.Sprintf("%s\n", format), v...)
	}
}

// 判断是否是等差数列
func isGradeList(abc []int) bool {
	var timeInt int
	if abc == nil || len(abc) == 1 {
		return true
	}
	timeInt = abc[0] - abc[1]
	for i := 0; i < len(abc)-1; i++ {
		a, b := abc[i], abc[i+1]
		if timeInt != a-b {
			return false
		}
		timeInt = a - b
	}
	return true
}

type TimeOut struct {
	s string
}
