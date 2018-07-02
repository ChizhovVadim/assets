package reports

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/ChizhovVadim/assets/core"
)

type QuoteReportService struct {
	historyCandleStorage  core.HistoryCandleStorage
	securityInfoDirectory core.SecurityInfoDirectory
}

func NewQuoteReportService(
	historyCandleStorage core.HistoryCandleStorage,
	securityInfoDirectory core.SecurityInfoDirectory) *QuoteReportService {
	return &QuoteReportService{
		historyCandleStorage:  historyCandleStorage,
		securityInfoDirectory: securityInfoDirectory,
	}
}

type QuoteReport struct {
	Start  time.Time
	Finish time.Time
	Items  []QuoteItem
}

type QuoteItem struct {
	SecurityCode string
	Title        string
	PriceStart   float64
	PriceFinish  float64
	Change       float64
}

func (srv *QuoteReportService) BuildQuoteReport(start, finish time.Time,
	securityCodes []string) (QuoteReport, error) {
	var years = yearsBetween(start, finish)
	var items []QuoteItem
	for _, securityCode := range securityCodes {
		priceStart, _ := srv.historyCandleStorage.CandleBeforeDate(securityCode, start)
		priceFinish, _ := srv.historyCandleStorage.CandleByDate(securityCode, finish)
		title := securityTitle(securityCode, srv.securityInfoDirectory)
		change := priceFinish.C / priceStart.C
		if years > 1 {
			change = math.Pow(change, 1.0/years)
		}
		items = append(items, QuoteItem{
			SecurityCode: securityCode,
			Title:        title,
			PriceStart:   priceStart.C,
			PriceFinish:  priceFinish.C,
			Change:       change,
		})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Change > items[j].Change
	})
	var report = QuoteReport{
		Start:  start,
		Finish: finish,
		Items:  items,
	}
	return report, nil
}

func PrintQuoteReport(report QuoteReport) {
	fmt.Printf("Изменение котировок с %v по %v\n",
		report.Start.Format(dateLayout),
		report.Finish.Format(dateLayout))

	var w = newTabWriter()
	fmt.Fprintf(w, "Security\tPrice\tChange\t\n")
	for _, item := range report.Items {
		fmt.Fprintf(w, "%v\t%v\t%.1f\t\n",
			item.Title, item.PriceFinish, (item.Change-1)*100)
	}
	w.Flush()
}
