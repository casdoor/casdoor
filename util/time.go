package util

import (
	"strconv"
	"time"
)

func GetCurrentTime() string {
	timestamp := time.Now().Unix()
	tm := time.Unix(timestamp, 0)
	return tm.Format(time.RFC3339)
}

func GetCurrentUnixTime() string {
	return strconv.FormatInt(time.Now().UnixNano(), 10)
}
