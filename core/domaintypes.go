package core

import (
	"errors"
	"time"
)

var (
	ErrNotImplemented = errors.New("not implemented")
	ErrNoData         = errors.New("no data")
)

type HistoryCandle struct {
	DateTime      time.Time
	O, H, L, C, V float64
}

type MyTrade struct {
	SecurityCode      string
	DateTime          time.Time
	ExecutionDate     time.Time
	Price             float64
	Volume            int
	ExchangeComission float64
	BrokerComission   float64
	Account           string
}

type DividendSchedule struct {
	SecurityCode     string
	RecordDate       time.Time
	Rate             float64
	ReceivedDividend *ReceivedDividend
}

type ReceivedDividend struct {
	Account string
	Date    time.Time
	Sum     float64
}

type SecurityInfo struct {
	SecurityCode string `xml:"Name,attr"`
	Title        string `xml:",attr"`
	Number       string `xml:",attr"`
	FinamCode    int    `xml:",attr"`
	Sector       string `xml:",attr"`
}

type MyTradeStorage interface {
	Read(account string) ([]MyTrade, error)
	Update(trades []MyTrade) error
}

type MyDividendStorage interface {
	ReadReceivedDividends(account string, start, finish time.Time) ([]ReceivedDividend, error)
	Read() ([]DividendSchedule, error)
}

type HistoryCandleStorage interface {
	CandleBeforeDate(securityCode string, date time.Time) (HistoryCandle, error)
	CandleByDate(securityCode string, date time.Time) (HistoryCandle, error)
	Last(securityCode string) (HistoryCandle, error)
	Update(securityCode string, candles []HistoryCandle) error
}

type SecurityInfoStorage interface {
	Read(securityCode string) (SecurityInfo, bool)
}
