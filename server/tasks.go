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
	"github.com/yushizhao/hubble/emailwrapper"
	"github.com/yushizhao/hubble/models"
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
	PnLReport := "PnL." + now + ".csv"
	BalanceReport := "Balance." + now + ".csv"

	// Balance
	fileBalance, err := os.Create(BalanceReport)
	if err != nil {
		return err
	}
	defer fileBalance.Close()
	writerBalance := csv.NewWriter(fileBalance)

	// PnL
	filePnL, err := os.Create(PnLReport)
	if err != nil {
		return err
	}
	defer filePnL.Close()
	writerPnL := csv.NewWriter(filePnL)

	var assets []string
	var totalPortfolios []models.Portfolio

	for _, a := range Memo.RealtimeAccounts {
		if a.PhysicalAccount == "TOTAL" {
			totalPortfolios = a.LogicalAccount
		}
	}

	for _, p := range totalPortfolios {
		if p.ClientCode == "TOTAL" {
			for k, _ := range p.PnLComponent {
				assets = append(assets, k)
			}
		}
	}

	// columns
	err = writerPnL.Write(append([]string{"", "TOTAL"}, assets...))
	if err != nil {
		return err
	}

	err = writerBalance.Write(append([]string{"", "TOTAL"}, assets...))
	if err != nil {
		return err
	}

	// rows
	for _, p := range totalPortfolios {

		rBalance := []string{p.ClientCode, strconv.FormatFloat(p.Value, 'g', -1, 64)}
		rPnL := []string{p.ClientCode, strconv.FormatFloat(p.PnL, 'g', -1, 64)}

		for _, asset := range assets {
			if v, ok := p.ValueComponent[asset]; ok {
				rBalance = append(rBalance, strconv.FormatFloat(v, 'g', -1, 64))
			} else {
				rBalance = append(rBalance, "")
			}
			if v, ok := p.PnLComponent[asset]; ok {
				rPnL = append(rPnL, strconv.FormatFloat(v, 'g', -1, 64))
			} else {
				rPnL = append(rPnL, "")
			}
		}

		err = writerBalance.Write(rBalance)
		if err != nil {
			return err
		}

		err = writerPnL.Write(rPnL)
		if err != nil {
			return err
		}
	}

	writerBalance.Flush()
	writerPnL.Flush()
	return emailwrapper.Send([]string{PnLReport, BalanceReport})
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
	writeReport := toolbox.NewTask("writeReport", config.Server.ReportSchedule, func() (err error) {
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
