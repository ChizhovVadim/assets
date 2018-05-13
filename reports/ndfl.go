package reports

import (
	"fmt"
	"math"
	"os"
	"text/tabwriter"
	"time"

	"github.com/ChizhovVadim/assets/core"
)

type NdflReportService struct {
	myTradeStorage       core.MyTradeStorage
	historyCandleStorage core.HistoryCandleStorage
	securityInfoStorage  core.SecurityInfoStorage
}

func NewNdflReportService(
	myTradeStorage core.MyTradeStorage,
	historyCandleStorage core.HistoryCandleStorage,
	securityInfoStorage core.SecurityInfoStorage) *NdflReportService {
	return &NdflReportService{
		myTradeStorage:       myTradeStorage,
		historyCandleStorage: historyCandleStorage,
		securityInfoStorage:  securityInfoStorage,
	}
}

type NdflReport struct {
	Year              int
	Account           string
	Trades            []ClosedMyTrade
	TotalPnL          float64
	Ndfl              float64
	NdflWithDeduction float64
}

type ClosedMyTrade struct {
	SecurityCode string
	OpenDate     time.Time
	CloseDate    time.Time
	OpenPrice    float64
	ClosePrice   float64
	Volume       int
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (srv *NdflReportService) BuildNdflReport(year int, account string) (NdflReport, error) {
	var tt, err = srv.myTradeStorage.Read(account)
	if err != nil {
		return NdflReport{}, err
	}
	var _, closedTrades = splitOpenAndClosedTrades(tt)
	closedTrades = filterClosedTrades(closedTrades, year)
	var totalPnL = totalPnL(closedTrades)
	var ndlf = computeNdfl(totalPnL)
	var pnlDeduction = computePnLDeduction(closedTrades)
	var ndflDeduction = computeNdfl(totalPnL - pnlDeduction)
	var report = NdflReport{
		Year:              year,
		Account:           account,
		Trades:            closedTrades,
		TotalPnL:          totalPnL,
		Ndfl:              ndlf,
		NdflWithDeduction: ndflDeduction,
	}
	return report, nil
}

type TaxFreeReport struct {
	Account     string
	Date        time.Time
	Items       []TaxFreeReportItem
	TotalAmount float64
}

type TaxFreeReportItem struct {
	SecuirtyCode string
	Volume       int
	Price        float64
	Amount       float64
}

func (srv *NdflReportService) BuildTaxFreeReport(account string, date time.Time) (TaxFreeReport, error) {
	var tt, err = srv.myTradeStorage.Read(account)
	if err != nil {
		return TaxFreeReport{}, err
	}
	//TODO отдельный отчет, что можем продать без налога без 3 летней льготы
	var openTrades, _ = splitOpenAndClosedTrades(tt)
	var m = make(map[string]int)
	for _, t := range openTrades {
		if t.ExecutionDate.Year() >= 2014 &&
			t.ExecutionDate.AddDate(3, 0, 0).Before(date) {
			m[t.SecurityCode] += t.Volume
		}
	}
	var report = TaxFreeReport{
		Account: account,
		Date:    date,
	}
	var total = 0.0
	for k, v := range m {
		var c, _ = srv.historyCandleStorage.Last(k)
		var amount = c.C * float64(v)
		total += amount
		report.Items = append(report.Items, TaxFreeReportItem{
			SecuirtyCode: k,
			Volume:       v,
			Price:        c.C,
			Amount:       amount,
		})
	}
	report.TotalAmount = total
	return report, nil
}

func PrintTaxFreeReport(report TaxFreeReport) {
	fmt.Printf("Отчет по 3-летней льготе '%v' на дату %v\n",
		report.Account, report.Date.Format("2006-01-02"))

	var w = tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.AlignRight)
	fmt.Fprintf(w, "Security\tVolume\tPrice\tAmount\t\n")
	for _, item := range report.Items {
		fmt.Fprintf(w, "%v\t%v\t%v\t%.f\t\n",
			item.SecuirtyCode, item.Volume, item.Price, item.Amount)
	}
	w.Flush()
	fmt.Printf("ИТОГО: %.f\n", report.TotalAmount)
}

func splitOpenAndClosedTrades(tt []core.MyTrade) ([]core.MyTrade, []ClosedMyTrade) {
	//TODO sort tt by date
	type info struct {
		buyTrades  []core.MyTrade
		sellTrades []core.MyTrade
	}
	var m = make(map[string]*info)
	for _, t := range tt {
		var item, found = m[t.SecurityCode]
		if !found {
			item = &info{}
			m[t.SecurityCode] = item
		}
		if t.Volume > 0 {
			item.buyTrades = append(item.buyTrades, t)
		} else {
			item.sellTrades = append(item.sellTrades, t)
		}
	}
	var closedTrades []ClosedMyTrade
	for _, v := range m {
		for _, sellTrade := range v.sellTrades {
			var sellVolume = -sellTrade.Volume
			for len(v.buyTrades) > 0 && sellVolume > 0 {
				var buyTrade = v.buyTrades[0]
				var volume = minInt(buyTrade.Volume, sellVolume)
				closedTrades = append(closedTrades, ClosedMyTrade{
					SecurityCode: buyTrade.SecurityCode,
					OpenDate:     buyTrade.ExecutionDate,
					CloseDate:    sellTrade.ExecutionDate,
					OpenPrice:    buyTrade.Price,
					ClosePrice:   sellTrade.Price,
					Volume:       volume,
				})
				sellVolume -= volume
				if volume >= buyTrade.Volume {
					v.buyTrades = v.buyTrades[1:]
				} else {
					v.buyTrades[0].Volume -= volume
				}
			}
		}
	}
	var openTrades []core.MyTrade
	for _, v := range m {
		openTrades = append(openTrades, v.buyTrades...)
	}
	return openTrades, closedTrades
}

func filterClosedTrades(source []ClosedMyTrade, year int) []ClosedMyTrade {
	var result []ClosedMyTrade
	for _, t := range source {
		if t.CloseDate.Year() == year {
			result = append(result, t)
		}
	}
	return result
}

func totalPnL(closedTrades []ClosedMyTrade) float64 {
	var sum = 0.0
	for _, t := range closedTrades {
		sum += (t.ClosePrice - t.OpenPrice) * float64(t.Volume) //TODO finished trades comission!
	}
	return sum
}

func computePnLDeduction(closedTrades []ClosedMyTrade) float64 {
	var sum = 0.0
	for _, t := range closedTrades {
		if t.OpenDate.Year() >= 2014 &&
			t.OpenDate.AddDate(3, 0, 0).Before(t.CloseDate) {
			sum += (t.ClosePrice - t.OpenPrice) * float64(t.Volume)
		}
	}
	if sum <= 0 {
		return 0
	}
	//TODO Учесть макс размер вычета
	return sum
}

func computeNdfl(pnl float64) float64 {
	if pnl < 0 {
		return 0
	}
	return math.Round(pnl * 0.13)
}

func PrintNdflReport(report NdflReport) {
	fmt.Printf("Отчет '%v' НДФЛ за %v год\n",
		report.Account, report.Year)
	fmt.Printf("Доход: %.f\n", report.TotalPnL)
	fmt.Printf("НДФЛ: %.f\n", report.Ndfl)
	fmt.Printf("НДФЛ с 3 летней льготой: %.f\n", report.NdflWithDeduction)
}
