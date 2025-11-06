package utils

import "time"

func TruncateToDateUTC(t time.Time) time.Time {
    utc := t.UTC()
    return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
}
