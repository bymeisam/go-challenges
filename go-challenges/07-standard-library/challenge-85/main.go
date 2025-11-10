package main

import "time"

func GetCurrentTime() time.Time {
	return time.Now()
}

func AddDuration(t time.Time, hours int) time.Time {
	return t.Add(time.Duration(hours) * time.Hour)
}

func FormatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func ParseTime(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}

func TimeDifference(t1, t2 time.Time) time.Duration {
	return t1.Sub(t2)
}

func main() {}
