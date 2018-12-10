package reports

import (
	"math"
	"time"
)

type DateSum struct {
	Date time.Time
	Sum  float64
}

type periodSum struct {
	period float64
	Sum    float64
}

func InternalRateOfReturn(cashflows []DateSum) float64 {
	var items = calculatePeriodSums(cashflows)
	return minimizeF(func(x float64) float64 {
		return math.Abs(npv(items, x))
	}, 0.11, 5, 0.1, 4)
}

func calculatePeriodSums(cashflows []DateSum) []periodSum {
	var now = time.Now()
	var result []periodSum
	for _, c := range cashflows {
		result = append(result, periodSum{
			period: yearsBetween(now, c.Date),
			Sum:    c.Sum,
		})
	}
	return result
}

func npv(source []periodSum, rate float64) float64 {
	var sum = 0.0
	for _, item := range source {
		sum += item.Sum * math.Pow(rate, -item.period)
	}
	return sum
}

func minimizeF(f func(float64) float64, start, end, delta float64, steps int) float64 {
	var bestError = f(start)
	for i := 0; i < steps; i++ {
		var curr = start - delta
		for curr < end {
			curr = curr + delta
			var thisError = f(curr)
			if thisError <= bestError {
				bestError = thisError
				start = curr
			}
		}
		// Narrow our search to [best - delta, best + delta]
		end = start + delta
		start = start - delta
		delta = delta / 10.0
	}
	return start
}
