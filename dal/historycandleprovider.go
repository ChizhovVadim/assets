package dal

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/ChizhovVadim/assets/core"
)

const (
	finamPeriodMinutes1 = 2
	finamPeriodMinutes5 = 3
	finamPeriodDay      = 8
)

type historyCandleProvider struct {
	securityInfoStorage core.SecurityInfoStorage
}

func NewHistoryCandleProvider(securityInfoStorage core.SecurityInfoStorage) *historyCandleProvider {
	return &historyCandleProvider{securityInfoStorage}
}

func (srv *historyCandleProvider) Load(security string,
	beginDate, endDate time.Time) ([]core.HistoryCandle, error) {

	var secInfo, found = srv.securityInfoStorage.Read(security)
	if !found {
		return nil, fmt.Errorf("securityCode not found %v", security)
	}
	var url, err = historyCandlesFinamUrl(secInfo.FinamCode, finamPeriodDay, beginDate, endDate)
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Timeout: 25 * time.Second,
	}
	return getHistoryCandles(client, url)
}

func historyCandlesFinamUrl(securityCode int, periodCode int,
	beginDate, endDate time.Time) (string, error) {
	baseUrl, err := url.Parse("http://export.finam.ru/data.txt?d=d&market=14&f=data.txt&e=.txt&cn=data&dtf=1&tmf=1&MSOR=0&sep=1&sep2=1&datf=1&at=1")
	if err != nil {
		return "", err
	}

	params, err := url.ParseQuery(baseUrl.RawQuery)
	if err != nil {
		return "", err
	}

	params.Set("em", strconv.Itoa(securityCode))
	params.Set("df", strconv.Itoa(beginDate.Day()))
	params.Set("mf", strconv.Itoa(int(beginDate.Month())-1))
	params.Set("yf", strconv.Itoa(beginDate.Year()))
	params.Set("dt", strconv.Itoa(endDate.Day()))
	params.Set("mt", strconv.Itoa(int(endDate.Month())-1))
	params.Set("yt", strconv.Itoa(endDate.Year()))
	params.Set("p", strconv.Itoa(periodCode))

	baseUrl.RawQuery = params.Encode()
	return baseUrl.String(), nil
}

func getHistoryCandles(client *http.Client, url string) ([]core.HistoryCandle, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:60.0) Gecko/20100101 Firefox/60.0")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http status %v", resp.Status)
	}

	var result []core.HistoryCandle
	csv := csv.NewReader(resp.Body)
	csv.Read() // skip fst line
	for {
		rec, err := csv.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		c, err := parseHistoryCandle(rec)
		if err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	if len(result) == 0 {
		return nil, core.ErrNoData
	}
	return result, nil
}
