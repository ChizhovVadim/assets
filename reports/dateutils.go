package reports

import (
	"time"
)

func date(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func yearsBetween(start, finish time.Time) float64 {
	return float64(finish.Sub(start)/(24*time.Hour)) / 365.25
}
