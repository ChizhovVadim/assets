package reports

import (
	"fmt"
	"math"
	"os"
	"text/tabwriter"
	"time"

	"github.com/ChizhovVadim/assets/core"
)

type DividendReportService struct {
	myTradeStorage      core.MyTradeStorage
	securityInfoStorage core.SecurityInfoStorage
	myDividendStorage   core.MyDividendStorage
}

func NewDividendReportService(
	myTradeStorage core.MyTradeStorage,
	securityInfoStorage core.SecurityInfoStorage,
	myDividendStorage core.MyDividendStorage) *DividendReportService {
	return &DividendReportService{
		myTradeStorage:      myTradeStorage,
		securityInfoStorage: securityInfoStorage,
		myDividendStorage:   myDividendStorage,
	}
}

type DividendReport struct {
	Year          int
	Account       string
	Items         []DividendItem
	TotalExpected float64
	TotalReceived float64
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
		var security = d.SecurityCode
		if si, found := srv.securityInfoStorage.Read(d.SecurityCode); found {
			security = si.Title
		}
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
		}
		report.Items = append(report.Items, item)
	}
	for _, item := range report.Items {
		report.TotalExpected += item.Expected
		report.TotalReceived += item.Payment
	}
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
	var sum = math.Round(rate * float64(shares) * 100)
	var ndfl = math.Round(sum * dividendTaxRate(date) / 100)
	return sum/100 - ndfl
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

	var w = tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.AlignRight)
	fmt.Fprintf(w, "Security\tRecord\tRate\tShares\tExpected\tDate\tPayment\t\n")
	for _, item := range report.Items {
		fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\t%v\t%v\t\n",
			item.Security,
			item.RecordDate.Format(dateLayout),
			item.Rate, item.Shares, item.Expected,
			formatZeroDate(item.PaymentDate),
			item.Payment)
	}
	w.Flush()
	fmt.Printf("Ожидаемая сумма дивидендов: %v\n", report.TotalExpected)
	fmt.Printf("Дивидендов получено: %v\n", report.TotalReceived)
}

func formatZeroDate(d time.Time) string {
	if d.IsZero() {
		return ""
	}
	return d.Format(dateLayout)
}
