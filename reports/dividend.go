package reports

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/ChizhovVadim/assets/core"
)

type DividendReportService struct {
	myTradeStorage        core.MyTradeStorage
	securityInfoDirectory core.SecurityInfoDirectory
	myDividendStorage     core.MyDividendStorage
}

func NewDividendReportService(
	myTradeStorage core.MyTradeStorage,
	securityInfoDirectory core.SecurityInfoDirectory,
	myDividendStorage core.MyDividendStorage) *DividendReportService {
	return &DividendReportService{
		myTradeStorage:        myTradeStorage,
		securityInfoDirectory: securityInfoDirectory,
		myDividendStorage:     myDividendStorage,
	}
}

type DividendReport struct {
	Year      int
	Account   string
	Items     []DividendItem
	ToReceive float64
}

type DividendItem struct {
	Security    string
	RecordDate  time.Time
	Rate        float64
	Shares      int
	Expected    float64
	PaymentDate time.Time
	Payment     float64
}

func (srv *DividendReportService) BuildDividendReport(year int,
	account string) (DividendReport, error) {
	tt, err := srv.myTradeStorage.Read(account)
	if err != nil {
		return DividendReport{}, err
	}
	dd, err := srv.myDividendStorage.Read()
	if err != nil {
		return DividendReport{}, err
	}
	var report = DividendReport{
		Year:    year,
		Account: account,
	}
	for _, d := range dd {
		if d.RecordDate.Year() != year {
			continue
		}
		var shares = calculateShares(tt, d.RecordDate, d.SecurityCode, account)
		if shares == 0 {
			continue
		}
		var security = securityTitle(d.SecurityCode, srv.securityInfoDirectory)
		var item = DividendItem{
			Security:   security,
			RecordDate: d.RecordDate,
			Rate:       d.Rate,
			Shares:     shares,
			Expected:   calculateExpectedDividend(d.Rate, shares, d.RecordDate), // or RecieveDate if exists?
		}
		if d.ReceivedDividend != nil {
			item.PaymentDate = d.ReceivedDividend.Date
			item.Payment = d.ReceivedDividend.Sum
		} else {
			report.ToReceive += item.Expected
		}
		report.Items = append(report.Items, item)
	}
	sort.Slice(report.Items, func(i, j int) bool {
		return report.Items[i].RecordDate.Before(report.Items[j].RecordDate)
	})
	return report, nil
}

func calculateShares(tt []core.MyTrade, date time.Time, securityCode, account string) int {
	var shares = 0
	for _, t := range tt {
		if t.SecurityCode == securityCode &&
			(account == "" || t.Account == account) &&
			!t.ExecutionDate.After(date) {
			shares += t.Volume
		}
	}
	return shares
}

func calculateExpectedDividend(rate float64, shares int, date time.Time) float64 {
	var sum = math.Round(rate*float64(shares)*100) / 100
	var ndfl = math.Round(sum * dividendTaxRate(date))
	return math.Round((sum-ndfl)*100) / 100
}

func dividendTaxRate(d time.Time) float64 {
	if d.Year() <= 2014 {
		return 0.09
	}
	return 0.13
}

func PrintDividendReport(report DividendReport) {
	fmt.Printf("Дивиденды '%v' в %v году\n",
		report.Account,
		report.Year)

	var w = newTabWriter()
	fmt.Fprintf(w, "Security\tRecord\tRate\tShares\tExpected\tDate\tPayment\t\n")
	for _, item := range report.Items {
		fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\t%v\t%v\t\n",
			item.Security,
			item.RecordDate.Format(dateLayout),
			item.Rate, item.Shares, item.Expected,
			formatZeroDate(item.PaymentDate),
			formatZeroFloat64(item.Payment))
	}
	w.Flush()
	fmt.Printf("Сумма дивидендов к получению: %.f\n", report.ToReceive)
}

func formatZeroFloat64(v float64) string {
	if v == 0 {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

func formatZeroDate(d time.Time) string {
	if d.IsZero() {
		return ""
	}
	return d.Format(dateLayout)
}
