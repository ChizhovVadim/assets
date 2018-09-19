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
	var lastItem = items[len(items)-1]
	var result = 1.1
	for iteration := 0; iteration < 100; iteration++ {
		var prevResult = result
		var s = 0.0
		for i := 0; i < len(items)-1; i++ {
			s += items[i].Sum / math.Pow(result, items[i].period)
		}
		result = math.Pow(-lastItem.Sum/s, 1/lastItem.period)
		if math.Abs(result-prevResult) < 0.0005 {
			break
		}
	}
	return result
}

func calculatePeriodSums(cashflows []DateSum) []periodSum {
	var max, min = maxAndMinDateOfDateSums(cashflows)
	var step = yearsBetween(min, max) / 20
	var m = make(map[float64]float64)
	for _, c := range cashflows {
		var period = round(yearsBetween(min, c.Date), step)
		m[period] += c.Sum
	}
	var result []periodSum
	for k, v := range m {
		result = append(result, periodSum{k, v})
	}
	var maxIndex = 0
	for i, item := range result {
		if item.Sum > result[maxIndex].Sum {
			maxIndex = i
		}
	}
	result[maxIndex], result[len(result)-1] = result[len(result)-1], result[maxIndex]
	return result
}

func maxAndMinDateOfDateSums(cashflows []DateSum) (max, min time.Time) {
	max = cashflows[0].Date
	min = max
	for _, item := range cashflows[1:] {
		if item.Date.Before(min) {
			min = item.Date
		}
		if item.Date.After(max) {
			max = item.Date
		}
	}
	return max, min
}

func round(x, unit float64) float64 {
	return math.Round(x/unit) * unit
}
