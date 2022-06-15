package xtime

import (
	"time"
)

// Seconds根據秒參數創建 time.Duration
func Seconds(seconds int64) time.Duration {
	return time.Duration(seconds) * time.Second
}

// NowAfter 從現在開始返回一個 unix 日期字符串加上持續時間
func NowAfter(unix time.Duration) string {
	date := time.Now().Add(unix)
	return FormatUnix(date)
}

// FormatUnix將 time.Time 格式化為 unix 日期字符串
func FormatUnix(date time.Time) string {
	return date.Format(time.UnixDate)
}

// FormatUnixInt將 int 時間戳格式化為 unix 日期字符串
func FormatUnixInt(unix int64) string {
	return FormatUnix(time.Unix(unix, 0))
}
