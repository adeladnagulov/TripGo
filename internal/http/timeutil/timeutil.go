package timeutil

import "time"

func FormatTime(t time.Time) time.Time {
	return t.UTC().Truncate(time.Second)
}

func FormatTimePtr(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	result := t.UTC().Truncate(time.Second)
	return &result
}
