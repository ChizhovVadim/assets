package reports

import (
	"math"
	"sort"
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
	var startDate = date(minDateOfDateSums(cashflows))
	var m = make(map[int]float64)
	for _, c := range cashflows {
		var period = daysBetween(startDate, date(c.Date))
		m[period] += c.Sum
	}
	var result []periodSum
	for k, v := range m {
		result = append(result, periodSum{float64(k) / 365.25, v})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].period < result[j].period
	})
	return result
}

func minDateOfDateSums(cashflows []DateSum) time.Time {
	var result = cashflows[0].Date
	for _, item := range cashflows[1:] {
		if item.Date.Before(result) {
			result = item.Date
		}
	}
	return result
}

func daysBetween(start, finish time.Time) int {
	return int(finish.Sub(start) / (24 * time.Hour))
}
