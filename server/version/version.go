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
	timeStr := "2006-01-02T15:04:05"
	t, err := time.Parse(timeStr, date)
	if err != nil {
		return
	}
	timestamp = t.Unix()
	return
}
func BuildVersion() (string, int64, error) {
	timestamp, err := convertToTimestamp(BuildTimestamp)
	if err != nil {
		return "", 0, err
	}
	return CommitHash, timestamp, nil
}
