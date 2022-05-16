package utils

import "time"

var (
	nowTS  int64
	nowStr string
)

// NowTS 获取当前时间
func NowTS() int64 {
	return nowTS
}

// NowStr 获取当前时间字符串
func NowStr() string {
	return nowStr
}

func init() {
	initTS()
	go func() {
		for range time.Tick(time.Second) {
			initTS()
		}
	}()
}

func initTS() {
	nowTS = time.Now().Unix()
	nowStr = time.Now().Format("2006-01-02 15:04:05")
}
