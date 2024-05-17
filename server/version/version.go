package version

import (
	"time"
)

var (
	Version        = "dev"
	CommitHash     = "n/a"
	BuildTimestamp = "n/a"
)

func convertToTimestamp(date string) (timestamp int64, err error) {
	timeStr := "2024-05-17T11:34:30"
	t, err := time.Parse(date, timeStr)
	if err != nil {
		return
	}
	timestamp = t.Unix()
	return
}
func BuildVersion() (string, int64) {
	timestamp, _ := convertToTimestamp(BuildTimestamp)
	return CommitHash, timestamp
}
