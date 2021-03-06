package reports

import (
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/ChizhovVadim/assets/core"
)

type PeriodReportService struct {
	myTradeStorage        core.MyTradeStorage
	historyCandleStorage  core.HistoryCandleStorage
	securityInfoDirectory core.SecurityInfoDirectory
	myDividendStorage     core.MyDividendStorage
}

func NewPeriodReportService(
	myTradeStorage core.MyTradeStorage,
	historyCandleStorage core.HistoryCandleStorage,
	securityInfoDirectory core.SecurityInfoDirectory,
	myDividendStorage core.MyDividendStorage) *PeriodReportService {
	return &PeriodReportService{
		myTradeStorage:        myTradeStorage,
		historyCandleStorage:  historyCandleStorage,
		securityInfoDirectory: securityInfoDirectory,
		myDividendStorage:     myDividendStorage,
	}
}

type PeriodReport struct {
	Start        time.Time
	Finish       time.Time
	Account      string
	Items        []PeriodItem
	AmountStart  float64
	AmountBuy    float64
	AmountSell   float64
	AmountChange float64
	AmountFinish float64
	Dividends    float64
	Comissions   float64
	PnL          float64
	Irr          float64
	Benchmark    float64
}

type PeriodItem struct {
	SecurityCode string
	Title        string
	PriceStart   float64
	PriceFinish  float64
	VolumeStart  int
	VolumeBuy    int
	VolumeSell   int
	VolumeFinish int
	AmountStart  float64
	AmountBuy    float64
	AmountSell   float64
	AmountChange float64
	AmountFinish float64
	Weight       float64
	Comissions   float64
}

type PeriodReportRequest struct {
	Start    time.Time
	Finish   time.Time
	Account  string
	Brief    bool
	Currency string
}

func (srv *PeriodReportService) BuildPeriodReport(r PeriodReportRequest) (PeriodReport, error) {
	var tt, err = srv.myTradeStorage.Read(r.Account)
	if err != nil {
		return PeriodReport{}, err
	}
	var curConv = &currencyConverter{
		codeTo:               r.Currency,
		historyCandleStorage: srv.historyCandleStorage,
	}
	var m = make(map[string]*PeriodItem)
	var cashflows []DateSum
	for _, t := range tt {
		//TODO compare date parts only?
		if t.ExecutionDate.After(r.Finish) {
			continue
		}
		var item, found = m[t.SecurityCode]
		if !found {
			item = &PeriodItem{SecurityCode: t.SecurityCode}
			m[t.SecurityCode] = item
		}
		if t.ExecutionDate.Before(r.Start) {
			item.VolumeStart += t.Volume
			item.VolumeFinish += t.Volume
		} else {
			item.VolumeFinish += t.Volume
			item.Comissions += curConv.Convert(t.ExecutionDate, t.BrokerComission+t.ExchangeComission)
			if t.Volume > 0 {
				item.VolumeBuy += t.Volume
				var amount = curConv.Convert(t.ExecutionDate, float64(t.Volume)*t.Price)
				item.AmountBuy += amount
				cashflows = append(cashflows, DateSum{t.ExecutionDate, -amount})
			} else {
				item.VolumeSell -= t.Volume
				var amount = curConv.Convert(t.ExecutionDate, float64(-t.Volume)*t.Price)
				item.AmountSell += amount
				cashflows = append(cashflows, DateSum{t.ExecutionDate, amount})
			}
		}
	}
	var items []PeriodItem
	for _, v := range m {
		if v.VolumeStart != 0 ||
			v.VolumeFinish != 0 ||
			v.VolumeBuy != 0 ||
			v.VolumeSell != 0 {
			if v.VolumeStart != 0 {
				c0, _ := srv.historyCandleStorage.CandleBeforeDate(v.SecurityCode, r.Start)
				v.PriceStart = c0.C
				v.AmountStart = curConv.Convert(r.Start, v.PriceStart*float64(v.VolumeStart))
			}
			if v.VolumeFinish != 0 {
				c1, _ := srv.historyCandleStorage.CandleByDate(v.SecurityCode, r.Finish)
				v.PriceFinish = c1.C
				v.AmountFinish = curConv.Convert(r.Finish, v.PriceFinish*float64(v.VolumeFinish))
			}
			v.AmountChange = v.AmountFinish - v.AmountStart - (v.AmountBuy - v.AmountSell)
			v.Title = securityTitle(v.SecurityCode, srv.securityInfoDirectory)
			items = append(items, *v)
		}
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].VolumeFinish == 0 && items[j].VolumeFinish == 0 {
			return items[i].Title < items[j].Title
		}
		return items[i].AmountFinish > items[j].AmountFinish
	})
	var result = PeriodReport{
		Start:   r.Start,
		Finish:  r.Finish,
		Account: r.Account,
		Items:   items,
	}
	for _, item := range items {
		result.AmountStart += item.AmountStart
		result.AmountBuy += item.AmountBuy
		result.AmountSell += item.AmountSell
		result.AmountFinish += item.AmountFinish
		result.Comissions += item.Comissions
	}
	for i := range items {
		items[i].Weight = items[i].AmountFinish / result.AmountFinish
	}
	cashflows = append(cashflows, DateSum{r.Start, -result.AmountStart})
	cashflows = append(cashflows, DateSum{r.Finish, result.AmountFinish})
	dd, err := srv.myDividendStorage.ReadReceivedDividends(r.Account, r.Start, r.Finish)
	if err != nil {
		return PeriodReport{}, err
	}
	for _, d := range dd {
		var dividend = curConv.Convert(d.Date, d.Sum)
		result.Dividends += dividend
		cashflows = append(cashflows, DateSum{d.Date, dividend})
	}
	result.AmountChange = result.AmountFinish - result.AmountStart - (result.AmountBuy - result.AmountSell)
	result.PnL = result.AmountChange + result.Dividends - result.Comissions
	result.Irr = InternalRateOfReturn(cashflows)
	if years := yearsBetween(r.Start, r.Finish); years < 1 {
		result.Irr = math.Pow(result.Irr, years)
	}
	result.Benchmark = srv.calculateBenchmark(r.Start, r.Finish)
	if r.Brief {
		var briefItems []PeriodItem
		for _, item := range items {
			if item.VolumeFinish != 0 {
				briefItems = append(briefItems, item)
			}
		}
		result.Items = briefItems
	}
	return result, nil
}

func (srv *PeriodReportService) calculateBenchmark(start, finish time.Time) float64 {
	const ticker = "MICEXINDEXCF"
	q0, _ := srv.historyCandleStorage.CandleBeforeDate(ticker, start)
	q1, _ := srv.historyCandleStorage.CandleByDate(ticker, finish)
	change := q1.C / q0.C
	if years := yearsBetween(start, finish); years > 1 {
		change = math.Pow(change, 1.0/years)
	}
	return change
}

func securityTitle(securityCode string,
	securityInfoDirectory core.SecurityInfoDirectory) string {
	info, found := securityInfoDirectory.Read(securityCode)
	if !found || info.Title == "" {
		return securityCode
	}
	return info.Title
}

const dateLayout = "2006-01-02"

func PrintPeriodReport(report PeriodReport) {
	fmt.Printf("Отчет '%v' с %v по %v\n",
		report.Account,
		report.Start.Format(dateLayout),
		report.Finish.Format(dateLayout))

	fmt.Printf("Стоимость активов на начало периода: %.f\n", report.AmountStart)
	fmt.Printf("Сумма зачисления: %.f\n", report.AmountBuy)
	fmt.Printf("Сумма списания: %.f\n", report.AmountSell)
	fmt.Printf("Изменение стоимости: %.f\n", report.AmountChange)
	fmt.Printf("Стоимость активов на конец периода: %.f\n", report.AmountFinish)
	fmt.Printf("Дивиденды: %.f\n", report.Dividends)
	fmt.Printf("Комиссия: %.f\n", report.Comissions)
	fmt.Printf("Доход: %.f\n", report.PnL)
	fmt.Printf("Доходность: %.1f%%\n", (report.Irr-1)*100)
	fmt.Printf("Доходность индекса: %.1f%%\n", (report.Benchmark-1)*100)

	var w = newTabWriter()
	fmt.Fprintf(w, "Security\tW1\tP1\tV0\tV+\tV-\tV1\tT1\t\n")
	for _, item := range report.Items {
		fmt.Fprintf(w, "%v\t%.1f\t%v\t%v\t%v\t%v\t%v\t%.f\t\n",
			item.Title, item.Weight*100, item.PriceFinish,
			formatZeroInt(item.VolumeStart), formatZeroInt(item.VolumeBuy), formatZeroInt(item.VolumeSell), formatZeroInt(item.VolumeFinish),
			item.AmountFinish)
	}
	w.Flush()
}

func formatZeroInt(v int) string {
	if v == 0 {
		return ""
	}
	return strconv.Itoa(v)
}

func newTabWriter() *tabwriter.Writer {
	return tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.AlignRight)
}

func (srv *PeriodReportService) GetHoldingTickers() ([]string, error) {
	var tt, err = srv.myTradeStorage.Read("")
	if err != nil {
		return nil, err
	}
	var m = make(map[string]int)
	for _, t := range tt {
		m[t.SecurityCode] += t.Volume
	}
	var result []string
	for k, v := range m {
		if v != 0 {
			result = append(result, k)
		}
	}
	return result, nil
}
