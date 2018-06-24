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

func (srv *PeriodReportService) BuildPeriodReport(start, finish time.Time,
	account string) (PeriodReport, error) {
	var tt, err = srv.myTradeStorage.Read(account)
	if err != nil {
		return PeriodReport{}, err
	}
	var m = make(map[string]*PeriodItem)
	var cashflows []DateSum
	for _, t := range tt {
		if t.ExecutionDate.After(finish) {
			continue
		}
		var item, found = m[t.SecurityCode]
		if !found {
			item = &PeriodItem{SecurityCode: t.SecurityCode}
			m[t.SecurityCode] = item
		}
		if t.ExecutionDate.Before(start) {
			item.VolumeStart += t.Volume
			item.VolumeFinish += t.Volume
		} else {
			item.VolumeFinish += t.Volume
			item.Comissions += t.BrokerComission + t.ExchangeComission
			if t.Volume > 0 {
				item.VolumeBuy += t.Volume
				var amount = float64(t.Volume) * t.Price
				item.AmountBuy += amount
				cashflows = append(cashflows, DateSum{t.ExecutionDate, -amount})
			} else {
				item.VolumeSell -= t.Volume
				var amount = float64(-t.Volume) * t.Price
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
				c0, _ := srv.historyCandleStorage.CandleBeforeDate(v.SecurityCode, start)
				v.PriceStart = c0.C
				v.AmountStart = v.PriceStart * float64(v.VolumeStart)
			}
			if v.VolumeFinish != 0 {
				c1, _ := srv.historyCandleStorage.CandleByDate(v.SecurityCode, finish)
				v.PriceFinish = c1.C
				v.AmountFinish = v.PriceFinish * float64(v.VolumeFinish)
			}
			v.AmountChange = v.AmountFinish - v.AmountStart - (v.AmountBuy - v.AmountSell)
			if info, found := srv.securityInfoDirectory.Read(v.SecurityCode); found {
				v.Title = info.Title
			} else {
				v.Title = v.SecurityCode
			}
			items = append(items, *v)
		}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Title < items[j].Title
	})
	var r = PeriodReport{
		Start:   start,
		Finish:  finish,
		Account: account,
		Items:   items,
	}
	for _, item := range items {
		r.AmountStart += item.AmountStart
		r.AmountBuy += item.AmountBuy
		r.AmountSell += item.AmountSell
		r.AmountFinish += item.AmountFinish
		r.Comissions += item.Comissions
	}
	for i := range items {
		items[i].Weight = items[i].AmountFinish / r.AmountFinish
	}
	cashflows = append(cashflows, DateSum{start, -r.AmountStart})
	cashflows = append(cashflows, DateSum{finish, r.AmountFinish})
	dd, err := srv.myDividendStorage.ReadReceivedDividends(account, start, finish)
	if err != nil {
		return PeriodReport{}, err
	}
	for _, d := range dd {
		r.Dividends += d.Sum
		cashflows = append(cashflows, DateSum{d.Date, d.Sum})
	}
	r.AmountChange = r.AmountFinish - r.AmountStart - (r.AmountBuy - r.AmountSell)
	r.PnL = r.AmountChange + r.Dividends - r.Comissions
	r.Irr = InternalRateOfReturn(cashflows)
	if years := yearsBetween(start, finish); years < 1 {
		r.Irr = math.Pow(r.Irr, years)
	}

	return r, nil
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
