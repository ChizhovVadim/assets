package main

import (
	"path"
	"strconv"
	"time"

	"github.com/ChizhovVadim/assets/dal"
	"github.com/ChizhovVadim/assets/reports"
)

const dateLayout = "2006-01-02"
const micexIndex = "MICEXINDEXCF"
const MOEXRussiaTotalReturnIndex = "MCFTR"
const USDCbrf = "USDCB"

var etfTickers = []string{
	"FXUS",
	"FXDE",
	"FXUK",
	"FXMM",
	"FXRU",
	"FXRB",
	"FXGD",
	"iFXIT",
	"FXJP",
	"FXAU",
	"FXCN",
	"FXRL",
}

// https://app2.msci.com/eqb/custom_indexes/russia_performance.html
var msciRussiaTickers = []string{
	"GAZP",
	"LKOH",
	"SBER",
	"MGNT",
	"SNGS",
	"SNGSP",
	"GMKN",
	"NVTK",
	"ROSN",
	"MTSS",
	"VTBR",
	"TATN",
	"TRNFP",
	"ALRS",
	"MOEX",
	"CHMF",
	"PHOR",
	"IRAO",
	"NLMK",
	"MAGN",
	"PLZL",
}

type controller struct {
	homeDir               string
	historyCandleService  *dal.HistoryCandleService
	periodReportService   *reports.PeriodReportService
	dividendReportService *reports.DividendReportService
	ndflReportService     *reports.NdflReportService
	quoteReportService    *reports.QuoteReportService
}

func (c *controller) updateHandler(args commandArgs) error {
	securityCodes := getTickersByType(c.periodReportService, args.params["type"])
	return c.historyCandleService.UpdateHistoryCandles(securityCodes)
}

func (c *controller) periodHandler(args commandArgs) error {
	var r = reports.PeriodReportRequest{}
	r.Brief = true
	r.Currency = args.params["cur"] // example: "USDCB"
	r.Account = args.params["account"]
	start, err := time.Parse(dateLayout, args.params["start"])
	if err != nil {
		start = firstDayOfYear(time.Now())
	}
	r.Start = start
	finish, err := time.Parse(dateLayout, args.params["finish"])
	if err != nil {
		//finish = time.Now()
		finish = today()
	}
	r.Finish = finish

	report, err := c.periodReportService.BuildPeriodReport(r)
	if err != nil {
		return err
	}
	reports.PrintPeriodReport(report)
	return nil
}

func (c *controller) dividendHandler(args commandArgs) error {
	account := args.params["account"]
	year, err := strconv.Atoi(args.params["year"])
	if err != nil {
		year = time.Now().Year()
	}

	report, err := c.dividendReportService.BuildDividendReport(year, account)
	if err != nil {
		return err
	}
	reports.PrintDividendReport(report)
	return nil
}

func (c *controller) ndflHandler(args commandArgs) error {
	account := args.params["account"]
	year, err := strconv.Atoi(args.params["year"])
	if err != nil {
		year = time.Now().Year()
	}

	report, err := c.ndflReportService.BuildNdflReport(year, account)
	if err != nil {
		return err
	}
	reports.PrintNdflReport(report)
	return nil
}

func (c *controller) taxfreeHandler(args commandArgs) error {
	account := args.params["account"]
	date, err := time.Parse(dateLayout, args.params["date"])
	if err != nil {
		date = time.Now()
	}

	report, err := c.ndflReportService.BuildPlannedTaxReport(account, date)
	if err != nil {
		return err
	}
	reports.PrintPlannedTaxReport(report)
	return nil
}

func (c *controller) importHandler(args commandArgs) error {
	importTradeService := dal.NewSberbankImportTradeService()
	tt, err := importTradeService.LoadTrades(path.Join(c.homeDir, "src.txt"))
	if err != nil {
		return err
	}
	importTradeStorage := dal.NewMyTradeStorage(path.Join(c.homeDir, "dst.txt"))
	return importTradeStorage.Update(tt)
}

func (c *controller) quoteHandler(args commandArgs) error {
	var r = reports.QuoteReportRequest{}
	start, err := time.Parse(dateLayout, args.params["start"])
	if err != nil {
		start = firstDayOfYear(time.Now())
	}
	r.Start = start
	finish, err := time.Parse(dateLayout, args.params["finish"])
	if err != nil {
		finish = time.Now()
	}
	r.Finish = finish
	r.SecurityCodes = getTickersByType(c.periodReportService, args.params["type"])
	r.Currency = args.params["cur"]

	report, err := c.quoteReportService.BuildQuoteReport(r)
	if err != nil {
		return err
	}
	reports.PrintQuoteReport(report)
	return nil
}

func firstDayOfYear(d time.Time) time.Time {
	return time.Date(d.Year(), 1, 1, 0, 0, 0, 0, d.Location())
}

func today() time.Time {
	y, m, d := time.Now().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, &time.Location{})
}

func getTickersByType(periodReportService *reports.PeriodReportService,
	securityType string) []string {
	switch securityType {
	case "":
		securityCodes, _ := periodReportService.GetHoldingTickers()
		securityCodes = append(securityCodes, micexIndex, USDCbrf)
		return securityCodes
	case "etf":
		return etfTickers
	case "stock":
		return msciRussiaTickers
	}
	return nil
}
