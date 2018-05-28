package dal

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/ChizhovVadim/assets/core"
)

type importTradeService struct {
	securityInfoStorage core.SecurityInfoStorage
}

func NewImportTradeService(securityInfoStorage core.SecurityInfoStorage) *importTradeService {
	return &importTradeService{
		securityInfoStorage: securityInfoStorage,
	}
}

func (srv *importTradeService) LoadTrades(fileName string) ([]core.MyTrade, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.Read() // skip fst line
	var result []core.MyTrade
	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		t, err := parseMyTradeSberbank(rec)
		if err != nil {
			return nil, err
		}
		result = append(result, t)
	}
	return result, nil
}

func parseMyTradeSberbank(line []string) (core.MyTrade, error) {
	const (
		DateTimeLayout = "2006-01-02 15:04:05"
	)
	if len(line) <= 18 {
		return core.MyTrade{}, fmt.Errorf("failed parseMyTradeSberbank %v", line)
	}
	account := line[0]
	date, err := time.Parse(DateTimeLayout, line[4])
	if err != nil {
		return core.MyTrade{}, err
	}
	executionDate, err := time.Parse(DateTimeLayout, line[5]) //Date Layout?
	if err != nil {
		return core.MyTrade{}, err
	}
	securityCode, err := parseSecurityCodeSberbank(line[6])
	if err != nil {
		return core.MyTrade{}, err
	}
	volume, err := strconv.Atoi(line[8])
	if err != nil {
		return core.MyTrade{}, err
	}
	if line[7] == "продажа" {
		volume *= -1
	}
	price, err := strconv.ParseFloat(line[11], 64)
	if err != nil {
		return core.MyTrade{}, err
	}
	exchangeCommisssion, err := strconv.ParseFloat(line[17], 64)
	if err != nil {
		return core.MyTrade{}, err
	}
	brokerCommisssion, err := strconv.ParseFloat(line[18], 64)
	if err != nil {
		return core.MyTrade{}, err
	}
	return core.MyTrade{
		Account:           account,
		DateTime:          date,
		ExecutionDate:     executionDate,
		SecurityCode:      securityCode,
		Volume:            volume,
		Price:             price,
		ExchangeComission: exchangeCommisssion,
		BrokerComission:   brokerCommisssion,
	}, nil
}

func parseSecurityCodeSberbank(securityName string) (string, error) {
	return "", core.ErrNotImplemented
}
