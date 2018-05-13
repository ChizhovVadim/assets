package dal

import (
	"time"

	"github.com/ChizhovVadim/assets/core"
)

type myDividendStorage struct {
	items []core.DividendSchedule
	err   error
}

func NewMyDividendStorage(path string) *myDividendStorage {
	items, err := loadMyDividends(path)
	return &myDividendStorage{items, err}
}

func (srv *myDividendStorage) Read() ([]core.DividendSchedule, error) {
	return srv.items, srv.err
}

func (srv *myDividendStorage) ReadReceivedDividends(account string,
	start, finish time.Time) ([]core.ReceivedDividend, error) {
	if srv.err != nil {
		return nil, srv.err
	}
	var result []core.ReceivedDividend
	for _, item := range srv.items {
		if item.ReceivedDividend == nil {
			continue
		}
		var item = *item.ReceivedDividend
		if !item.Date.Before(start) &&
			!item.Date.After(finish) &&
			(account == "" || item.Account == account) {
			result = append(result, item)
		}
	}
	return result, nil
}

func loadMyDividends(path string) ([]core.DividendSchedule, error) {
	const DateLayout = "2006-01-02"
	type myDividend struct {
		Account      string  `xml:",attr"`
		SecurityCode string  `xml:"Name,attr"`
		RecordDate   string  `xml:",attr"`
		Rate         float64 `xml:",attr"`
		RecieveDate  string  `xml:",attr"`
		RecieveSum   float64 `xml:",attr"`
	}
	var obj = struct {
		Items []myDividend `xml:"Dividend"`
	}{}
	var err = decodeXmlFile(path, &obj)
	if err != nil {
		return nil, err
	}
	var dividends []core.DividendSchedule
	for _, item := range obj.Items {
		if item.RecordDate == "" {
			continue
		}
		recordDate, err := time.Parse(DateLayout, item.RecordDate)
		if err != nil {
			return nil, err
		}
		var receivedDividend *core.ReceivedDividend
		if item.RecieveDate != "" {
			recieveDate, err := time.Parse(DateLayout, item.RecieveDate)
			if err != nil {
				return nil, err
			}
			receivedDividend = &core.ReceivedDividend{
				Account: item.Account,
				Date:    recieveDate,
				Sum:     item.RecieveSum,
			}
		}
		dividends = append(dividends, core.DividendSchedule{
			SecurityCode:     item.SecurityCode,
			Rate:             item.Rate,
			RecordDate:       recordDate,
			ReceivedDividend: receivedDividend,
		})
	}
	return dividends, nil
}
