package dal

import (
	"log"
	"time"

	"github.com/ChizhovVadim/assets/core"
)

type HistoryCandleProvider interface {
	Load(securityCode string, beginDate, endDate time.Time) ([]core.HistoryCandle, error)
}

type HistoryCandleService struct {
	historyCandleStorage  core.HistoryCandleStorage
	historyCandleProvider HistoryCandleProvider
	startHistoryDate      time.Time
}

func NewHistoryCandleService(
	historyCandleStorage core.HistoryCandleStorage,
	historyCandleProvider HistoryCandleProvider) *HistoryCandleService {
	var startHistoryDate = time.Date(2013, time.January, 1, 0, 0, 0, 0, time.Local)
	return &HistoryCandleService{historyCandleStorage, historyCandleProvider, startHistoryDate}
}

func (srv *HistoryCandleService) UpdateHistoryCandles(securityCodes []string) error {
	log.Println("Обновляем исторические котировки...")
	for i, securityCode := range securityCodes {
		if i > 0 {
			time.Sleep(1000 * time.Millisecond)
		}
		var err = srv.UpdateHistoryCandlesBySecurityCode(securityCode)
		if err != nil {
			log.Printf("update failed first %v %v", securityCode, err)
			time.Sleep(1000 * time.Millisecond)
			err = srv.UpdateHistoryCandlesBySecurityCode(securityCode)
			if err != nil {
				log.Printf("update failed second %v %v", securityCode, err)
			}
		}
	}
	log.Println("Исторические котировки обновлены.")
	return nil
}

func (srv *HistoryCandleService) UpdateHistoryCandlesBySecurityCode(securityCode string) error {
	var startDate time.Time
	if last, err := srv.historyCandleStorage.Last(securityCode); err != nil {
		if err == core.ErrNoData {
			startDate = srv.startHistoryDate
		} else {
			return err
		}
	} else {
		startDate = last.DateTime
	}
	var candles, err = srv.historyCandleProvider.Load(securityCode, startDate, time.Now())
	if err != nil {
		return err
	}
	return srv.historyCandleStorage.Update(securityCode, candles)
}
