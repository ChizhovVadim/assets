package dal

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ChizhovVadim/assets/core"
)

const (
	myTradeStorageDateTimeLayout = "2006-01-02T15:04:05"
	myTradeStorageDateLayout     = "2006-01-02"
)

type myTradeStorage struct {
	path string
}

func NewMyTradeStorage(path string) *myTradeStorage {
	return &myTradeStorage{path}
}

func (srv *myTradeStorage) Read(account string) ([]core.MyTrade, error) {
	file, err := os.Open(srv.path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	csv := csv.NewReader(file)
	csv.Read() // skip fst line
	var result []core.MyTrade
	for {
		rec, err := csv.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		t, err := parseMyTrade(rec)
		if err != nil {
			return nil, err
		}
		if account == "" || strings.EqualFold(t.Account, account) {
			result = append(result, t)
		}
	}
	return result, nil
}

func (srv *myTradeStorage) Update(trades []core.MyTrade) error {
	file, err := os.Create(srv.path)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	for _, t := range trades {
		rec := []string{
			t.SecurityCode,
			t.DateTime.Format(myTradeStorageDateTimeLayout),
			t.ExecutionDate.Format(myTradeStorageDateLayout),
			strconv.FormatFloat(t.Price, 'g', -1, 64),
			strconv.Itoa(t.Volume),
			strconv.FormatFloat(t.ExchangeComission, 'g', -1, 64),
			strconv.FormatFloat(t.BrokerComission, 'g', -1, 64),
			t.Account,
		}
		err = writer.Write(rec)
		if err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}

func parseMyTrade(record []string) (core.MyTrade, error) {
	if len(record) < 8 {
		return core.MyTrade{}, fmt.Errorf("parseMyTrade len %v", record)
	}
	securityCode := record[0]
	d, err := time.Parse(myTradeStorageDateTimeLayout, record[1])
	if err != nil {
		return core.MyTrade{}, err
	}
	execDate, err := time.Parse(myTradeStorageDateLayout, record[2])
	if err != nil {
		return core.MyTrade{}, err
	}
	price, err := strconv.ParseFloat(record[3], 64)
	if err != nil {
		return core.MyTrade{}, err
	}
	volume, err := strconv.Atoi(record[4])
	if err != nil {
		return core.MyTrade{}, err
	}
	exCom, err := strconv.ParseFloat(record[5], 64)
	if err != nil {
		return core.MyTrade{}, err
	}
	brCom, err := strconv.ParseFloat(record[6], 64)
	if err != nil {
		return core.MyTrade{}, err
	}
	account := record[7]
	return core.MyTrade{securityCode, d, execDate, price, volume, exCom, brCom, account}, nil
}
