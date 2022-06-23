package utils

import "time"

func StringPtr(s string) *string {
	return &s
}

func TimePtr(t time.Time) *time.Time {
	return &t
}

func Int64Ptr(i int64) *int64 {
	return &i
}
