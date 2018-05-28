package main

import (
	"log"
	"os"
	"os/user"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/ChizhovVadim/assets/dal"
	"github.com/ChizhovVadim/assets/reports"
)

var tickers = []string{
	"MICEXINDEXCF",
	//MSCI Russia
	"GAZP", "LKOH", "SBER", "MGNT", "GMKN", "NVTK", "ROSN", "SNGSP",
	"MTSS", "ALRS", "CHMF", "MOEX", "HYDR", "PHOR",
	"NLMK", "SNGS", "IRAO", "MAGN", "PLZL", "TATN", "VTBR",
	//Прочие
	"TATNP", "AFKS",
}

type CliContext struct {
	CommandName string
	Flags       map[string]string
}

func main() {
	const dateLayout = "2006-01-02"

	curUser, err := user.Current()
	if err != nil {
		log.Print(err)
		return
	}
	homeDir := curUser.HomeDir
	if homeDir == "" {
		log.Print("Current user home dir empty.")
		return
	}

	assetsDir := path.Join(homeDir, "Projects/Assets/Assets/Content")
	securityInfoStorage := dal.NewSecurityInfoStorage(path.Join(assetsDir, "StockSettings.xml"))
	myTradeStorage := dal.NewMyTradeStorage(path.Join(assetsDir, "trades.csv"))
	myDividendStorage := dal.NewMyDividendStorage(path.Join(assetsDir, "Dividends.xml"))
	historyCandleStorage := dal.NewHistoryCandleStorage(path.Join(homeDir, "TradingData/Portfolio"))

	historyCandleService := dal.NewHistoryCandleService(historyCandleStorage,
		dal.NewHistoryCandleProvider(securityInfoStorage))
	periodReportService := reports.NewPeriodReportService(myTradeStorage, historyCandleStorage, securityInfoStorage, myDividendStorage)
	dividendReportService := reports.NewDividendReportService(myTradeStorage, securityInfoStorage, myDividendStorage)
	ndflReportService := reports.NewNdflReportService(myTradeStorage, historyCandleStorage, securityInfoStorage)

	var commands = map[string]func(ctx CliContext) error{
		"update": func(ctx CliContext) error {
			return historyCandleService.UpdateHistoryCandles(tickers)
		},
		"period": func(ctx CliContext) error {
			account := ctx.Flags["account"]
			start, err := time.Parse(dateLayout, ctx.Flags["start"])
			if err != nil {
				start = firstDayOfYear(time.Now())
			}
			finish, err := time.Parse(dateLayout, ctx.Flags["finish"])
			if err != nil {
				finish = time.Now()
			}

			report, err := periodReportService.BuildPeriodReport(start, finish, account)
			if err != nil {
				return err
			}
			reports.PrintPeriodReport(report)
			return nil
		},
		"dividend": func(ctx CliContext) error {
			account := ctx.Flags["account"]
			year, err := strconv.Atoi(ctx.Flags["year"])
			if err != nil {
				year = time.Now().Year()
			}

			report, err := dividendReportService.BuildDividendReport(year, account)
			if err != nil {
				return err
			}
			reports.PrintDividendReport(report)
			return nil
		},
		"ndfl": func(ctx CliContext) error {
			account := ctx.Flags["account"]
			year, err := strconv.Atoi(ctx.Flags["year"])
			if err != nil {
				year = time.Now().Year()
			}

			report, err := ndflReportService.BuildNdflReport(year, account)
			if err != nil {
				return err
			}
			reports.PrintNdflReport(report)
			return nil
		},
		"taxfree": func(ctx CliContext) error {
			account := ctx.Flags["account"]
			date, err := time.Parse(dateLayout, ctx.Flags["date"])
			if err != nil {
				date = time.Now()
			}

			report, err := ndflReportService.BuildPlannedTaxReport(account, date)
			if err != nil {
				return err
			}
			reports.PrintPlannedTaxReport(report)
			return nil
		},
		"import": func(ctx CliContext) error {
			importTradeService := dal.NewImportTradeService(securityInfoStorage)
			tt, err := importTradeService.LoadTrades(path.Join(homeDir, "src.txt"))
			if err != nil {
				return err
			}
			importTradeStorage := dal.NewMyTradeStorage(path.Join(homeDir, "dst.txt"))
			return importTradeStorage.Update(tt)
		},
	}
	var ctx = parseFlags()
	var cmd, found = commands[ctx.CommandName]
	if !found {
		log.Println("Command not found.")
		return
	}
	err = cmd(ctx)
	if err != nil {
		log.Print(err)
	}
}

func parseFlags() CliContext {
	var args = os.Args
	var cmdName = ""
	var flags = make(map[string]string)
	for i := 1; i < len(args); i++ {
		var arg = args[i]
		if strings.HasPrefix(arg, "-") {
			if i < len(args)-1 {
				var k = strings.TrimPrefix(arg, "-")
				var v = args[i+1]
				flags[k] = v
			}
		} else if cmdName == "" {
			cmdName = arg
		}
	}
	return CliContext{cmdName, flags}
}

func firstDayOfYear(d time.Time) time.Time {
	return time.Date(d.Year(), 1, 1, 0, 0, 0, 0, d.Location())
}
