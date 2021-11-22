package util

import (
	"strconv"
	"time"
)

func GetFirstDayOfMonth() time.Time {
	now := time.Now()
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()

	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)

	return firstOfMonth
}

func GetUTCTime(t time.Time) float64 {
	zoneStr := t.Format("-07")
	byteNow := []byte(zoneStr)
	result, _ := strconv.Atoi(string(byteNow[1:]))
	if string(byteNow[0]) == "-" {
		return float64(result) * -1.0
	}
	return float64(result)
}
