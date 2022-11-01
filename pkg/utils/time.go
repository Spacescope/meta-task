package utils

import "time"

// 获取当天起始时间: 00:00:00
func GetStartOfTheDay() time.Time {
	timestr := time.Now().Format("2006-01-02")
	t, _ := time.Parse("2006-01-02", timestr)
	return t
}

func GetEndOfTheDay() time.Time {
	return GetStartOfTheDay().Add(23*time.Hour + 59*time.Minute + 59*time.Second)
}
