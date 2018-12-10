package reports

import (
	"math"
	"time"

	"github.com/ChizhovVadim/assets/core"
)

type currencyConverter struct {
	historyCandleStorage core.HistoryCandleStorage
	codeTo               string
	candles              []core.HistoryCandle
}

func (srv *currencyConverter) Convert(d time.Time, v float64) float64 {
	if srv.codeTo == "" {
		return v
	}
	var candle, err = srv.Candle(d)
	if err != nil {
		return math.NaN()
	}
	return v / candle.C
}

func (srv *currencyConverter) Candle(d time.Time) (core.HistoryCandle, error) {
	if srv.candles == nil {
		var candles, err = srv.historyCandleStorage.Read(srv.codeTo)
		if err != nil {
			return core.HistoryCandle{}, err
		}
		srv.candles = candles
	}
	var index = -1
	for i := range srv.candles {
		var item = &srv.candles[i]
		if item.DateTime.After(d) {
			break
		}
		index = i
	}
	if index == -1 {
		return core.HistoryCandle{}, core.ErrNoData
	}
	return srv.candles[index], nil
}
