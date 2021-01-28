package main

import (
	"fmt"
	"math/rand"
	"time"
	"timeWhell/Timewheel"
)

var keys []int64

func main() {
	dta := Timewheel.NewTimeWheel()
	println("开始")
	//println(dta.GetTime())
	//println(time.Now().Format("2006-01-02 15:04:05"))
	//go dta.Start()
	//time.Sleep(time.Second * 10)
	//
	//for {
	//	a := rand.Int()
	//	time.Sleep(time.Second * 5)
	//	if a%3 == 0 {
	//		key, _ := dta.AddTask(10000, huidiao, time.Now().Format("2006-01-02 15:04:05"))
	//		keys = append(keys, key)
	//	} else if a%3 == 1 {
	//
	//	} else if len(keys) > 1 {
	//		x := keys[0]
	//		keys = keys[1:]
	//		err := dta.DeleteTask(x)
	//		if err != nil {
	//			fmt.Printf("%x删除成功\n", x)
	//		} else {
	//			println("删除失败")
	//		}
	//	}
	//
	//}
	go func() {
		time.Sleep(time.Second * 10)
		for {
			a := rand.Int()
			time.Sleep(time.Second * 5)
			if a%3 == 0 {
				key, _ := dta.AddTask(10000, huidiao, time.Now().Format("2006-01-02 15:04:05"))
				keys = append(keys, key)
			} else if a%3 == 1 {

			} else if len(keys) > 1 {
				x := keys[0]
				keys = keys[1:]
				_ = dta.DeleteTask(x)
				//if err != nil {
				//	fmt.Printf("%x删除成功\n", x)
				//} else {
				//	println("删除失败")
				//}
			}

		}
	}()
	dta.Start()
}

func huidiao(inter interface{}) {
	switch inter.(type) {
	case string:
		fmt.Println("当前时间是：", time.Now().Format("2006-01-02 15:04:05"), " 添加时间是：", inter.(string))
		break
	case int:
		fmt.Println("int", inter.(int))
		break
	case float64:
		fmt.Println("float64", inter.(float64))
		break
	}
}
