package dal

import (
	"encoding/csv"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ChizhovVadim/assets/core"
)

type historyCandleStorage struct {
	historyCandlesFolder string
}

func NewHistoryCandleStorage(historyCandlesFolder string) *historyCandleStorage {
	return &historyCandleStorage{historyCandlesFolder}
}

func (srv *historyCandleStorage) fileName(securityCode string) string {
	return filepath.Join(srv.historyCandlesFolder, securityCode+".txt")
}

func (srv *historyCandleStorage) CandleBeforeDate(securityCode string, date time.Time) (core.HistoryCandle, error) {
	var cc, err = srv.readAll(securityCode)
	if err != nil {
		return core.HistoryCandle{}, err
	}
	var index = lastIndexOfHistoryCandle(cc, func(c core.HistoryCandle) bool {
		return c.DateTime.Before(date)
	})
	if index == -1 {
		return core.HistoryCandle{}, core.ErrNoData
	}
	return cc[index], nil
}

func (srv *historyCandleStorage) CandleByDate(securityCode string, date time.Time) (core.HistoryCandle, error) {
	var cc, err = srv.readAll(securityCode)
	if err != nil {
		return core.HistoryCandle{}, err
	}
	var index = lastIndexOfHistoryCandle(cc, func(c core.HistoryCandle) bool {
		return !c.DateTime.After(date)
	})
	if index == -1 {
		return core.HistoryCandle{}, core.ErrNoData
	}
	return cc[index], nil
}

func (srv *historyCandleStorage) Last(securityCode string) (core.HistoryCandle, error) {
	var cc, err = srv.readAll(securityCode)
	if err != nil {
		return core.HistoryCandle{}, err
	}
	if len(cc) == 0 {
		return core.HistoryCandle{}, core.ErrNoData
	}
	return cc[len(cc)-1], nil
}

func (srv *historyCandleStorage) readAll(securityCode string) ([]core.HistoryCandle, error) {
	file, err := os.Open(srv.fileName(securityCode))
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var result []core.HistoryCandle
	csv := csv.NewReader(file)
	csv.Read() // skip fst line
	for {
		rec, err := csv.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		candle, err := parseHistoryCandle(rec)
		if err != nil {
			return nil, err
		}
		result = append(result, candle)
	}
	return result, nil
}

func (srv *historyCandleStorage) writeAll(securityCode, filename string, source []core.HistoryCandle) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	csv := csv.NewWriter(f)
	err = csv.Write(strings.Split("<TICKER>,<PER>,<DATE>,<TIME>,<OPEN>,<HIGH>,<LOW>,<CLOSE>,<VOL>", ","))
	if err != nil {
		return err
	}
	for _, c := range source {
		record := []string{
			securityCode,
			"5",
			c.DateTime.Format("20060102"),
			strconv.Itoa(100 * (100*c.DateTime.Hour() + c.DateTime.Minute())), //TODO 000000
			strconv.FormatFloat(c.O, 'f', -1, 64),
			strconv.FormatFloat(c.H, 'f', -1, 64),
			strconv.FormatFloat(c.L, 'f', -1, 64),
			strconv.FormatFloat(c.C, 'f', -1, 64),
			strconv.FormatFloat(c.V, 'f', -1, 64),
		}
		err := csv.Write(record)
		if err != nil {
			return err
		}
	}
	csv.Flush()
	return csv.Error()
}

func lastIndexOfHistoryCandle(source []core.HistoryCandle,
	condition func(item core.HistoryCandle) bool) int {
	var result = -1
	for i, item := range source {
		if !condition(item) {
			return result
		}
		result = i
	}
	return result
}

func (srv *historyCandleStorage) Update(securityCode string, candles []core.HistoryCandle) error {
	var currentCandles, err = srv.readAll(securityCode)
	if err != nil {
		return err
	}
	var d = date(candles[0].DateTime)
	var index = lastIndexOfHistoryCandle(currentCandles, func(c core.HistoryCandle) bool {
		return date(c.DateTime).Before(d)
	})
	currentCandles = append(currentCandles[:index+1], candles...)
	//TODO write to temp file first?
	return srv.writeAll(securityCode, srv.fileName(securityCode), currentCandles)
}

func parseHistoryCandle(record []string) (candle core.HistoryCandle, err error) {
	d, err := time.Parse("20060102", record[2])
	if err != nil {
		return
	}
	t, err := strconv.Atoi(record[3])
	if err != nil {
		return
	}
	var hour = t / 10000
	var min = (t / 100) % 100
	d = d.Add(time.Duration(hour)*time.Hour + time.Duration(min)*time.Minute)
	o, err := strconv.ParseFloat(record[4], 64)
	if err != nil {
		return
	}
	h, err := strconv.ParseFloat(record[5], 64)
	if err != nil {
		return
	}
	l, err := strconv.ParseFloat(record[6], 64)
	if err != nil {
		return
	}
	c, err := strconv.ParseFloat(record[7], 64)
	if err != nil {
		return
	}
	v, err := strconv.ParseFloat(record[8], 64)
	if err != nil {
		return
	}
	candle = core.HistoryCandle{d, o, h, l, c, v}
	return
}

func date(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}
