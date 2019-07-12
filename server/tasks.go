package server

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"strconv"
	"time"

	"github.com/wonderivan/logger"

	"github.com/astaxie/beego/toolbox"
	"github.com/yushizhao/hubble/config"
)

// func TaskUpdateFromDepth() {
// 	updateFromDepth := toolbox.NewTask("updateFromDepth", "0 0 */2 * * *", func() error {
// 		// this task will run every 2 hours
// 		logger.Info("run updateFromDepth at: %s\n", time.Now())

// 		rawKeys, err := MarketDataSource.Keys("DEPTHx*")
// 		if err != nil {
// 			logger.Error(err)
// 			return err
// 		}

// 		// symbol list
// 		var tmpSymbols []string
// 		// update SymbolsMapExchanges, directly replace the order map
// 		newSME := make(map[string][]string)

// 		Memo.LockSymbolsMapExchanges.Lock()
// 		for _, rawKey := range rawKeys {
// 			keys := strings.Split(rawKey[7:], ".")
// 			tmpSymbols = append(tmpSymbols, keys[0])
// 			newSME[keys[0]] = append(newSME[keys[0]], keys[1])
// 		}
// 		Memo.SymbolsMapExchanges = newSME
// 		Memo.LockSymbolsMapExchanges.Unlock()

// 		logger.Info("%v\n", Memo.SymbolsMapExchanges)

// 		// dummy trade
// 		tmpTrade := models.TRADE{
// 			Exchange:   "",
// 			Time:       "",
// 			TimeArrive: "",
// 			Direction:  "",
// 			LastPx:     -1,
// 			Qty:        -1,
// 		}
// 		// update SymbolsMapLastTrade, move trade data from the order map to the new map
// 		newSMT := make(map[string]models.TRADE)

// 		Memo.LockSymbolsMapLastTrade.Lock()
// 		for _, sym := range tmpSymbols {

// 			if val, ok := Memo.SymbolsMapLastTrade[sym]; ok {
// 				newSMT[sym] = val
// 			} else {
// 				tmpTrade.Symbol = sym
// 				newSMT[sym] = tmpTrade
// 			}

// 		}
// 		Memo.SymbolsMapLastTrade = newSMT
// 		Memo.LockSymbolsMapLastTrade.Unlock()

// 		logger.Info("%v\n", Memo.SymbolsMapLastTrade)

// 		return nil
// 	})

// 	toolbox.AddTask("updateFromDepth", updateFromDepth)
// }

func Report() error {
	now := time.Now().Format(time.RFC3339)
	file, err := os.Create("PnL." + now + ".csv")
	if err != nil {
		return err
	}

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, a := range Memo.RealtimeAccounts {
		if a.PhysicalAccount == "TOTAL" {
			for _, p := range a.LogicalAccount {
				err := writer.Write([]string{p.ClientCode, strconv.FormatFloat(p.PnL, 'g', -1, 64)})
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func Backup() error {

	now := time.Now().Format(time.RFC3339)

	b, err := json.Marshal(Memo.RealtimeAccounts)
	if err != nil {
		return err
	}

	bak, err := os.Create(now + ".bak.json")
	if err != nil {
		return err
	}
	_, err = bak.Write(b)
	if err != nil {
		return err
	}

	err = os.Remove("account.json")
	if err != nil {
		return err
	}

	f, err := os.Create("account.json")
	if err != nil {
		return err
	}

	_, err = f.Write(b)
	if err != nil {
		return err
	}

	return nil
}

func TaskWriteReport() {
	writeReport := toolbox.NewTask("writeReport", config.Conf.ReportSchedule, func() (err error) {
		logger.Info("run writeReport at: %s\n", time.Now())

		Memo.LockRealtimeAccounts.RLock()
		Memo.LockAccounts.Lock()
		defer Memo.LockRealtimeAccounts.RUnlock()
		defer Memo.LockAccounts.Unlock()

		// log&backup realtimeAccounts
		err = Backup()
		if err != nil {
			logger.Error(err)
		}

		// reset Accounts
		Memo.Accounts = Memo.RealtimeAccounts

		// write report
		err = Report()
		if err != nil {
			logger.Error(err)
		}

		return nil
	})
	toolbox.AddTask("writeReport", writeReport)
}
