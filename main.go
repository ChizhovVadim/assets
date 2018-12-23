package main

import (
	"log"
	"os/user"
	"path"

	"github.com/ChizhovVadim/assets/dal"
	"github.com/ChizhovVadim/assets/reports"
)

func main() {
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
	securityInfoDirectory := dal.NewSecurityInfoDirectory(securityInfoStorage)
	myTradeStorage := dal.NewMyTradeStorage(path.Join(assetsDir, "trades.csv"))
	myDividendStorage := dal.NewMyDividendStorage(path.Join(assetsDir, "Dividends.xml"))
	historyCandleStorage := dal.NewHistoryCandleStorage(path.Join(homeDir, "TradingData/Portfolio"))

	historyCandleService := dal.NewHistoryCandleService(historyCandleStorage,
		dal.NewHistoryCandleProvider(securityInfoDirectory))
	periodReportService := reports.NewPeriodReportService(myTradeStorage, historyCandleStorage, securityInfoDirectory, myDividendStorage)
	dividendReportService := reports.NewDividendReportService(myTradeStorage, securityInfoDirectory, myDividendStorage)
	ndflReportService := reports.NewNdflReportService(myTradeStorage, historyCandleStorage, securityInfoDirectory)
	quoteReportService := reports.NewQuoteReportService(historyCandleStorage, securityInfoDirectory)

	controller := &controller{
		homeDir:               homeDir,
		historyCandleService:  historyCandleService,
		periodReportService:   periodReportService,
		dividendReportService: dividendReportService,
		ndflReportService:     ndflReportService,
		quoteReportService:    quoteReportService,
	}

	runCommands([]command{
		command{"update", controller.updateHandler},
		command{"period", controller.periodHandler},
		command{"dividend", controller.dividendHandler},
		command{"ndfl", controller.ndflHandler},
		command{"taxfree", controller.taxfreeHandler},
		command{"import", controller.importHandler},
		command{"quote", controller.quoteHandler},
	})
}
