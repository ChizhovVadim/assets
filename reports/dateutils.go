package reports

import (
	"time"
)

func yearsBetween(start, finish time.Time) float64 {
	return float64(finish.Sub(start)/(24*time.Hour)) / 365.25
}
