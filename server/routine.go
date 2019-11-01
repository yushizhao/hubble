package server

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/wonderivan/logger"
	"github.com/yushizhao/hubble/models"
)

func UpdateFromDepth() {

	for {
		logger.Info("run updateFromDepth at: %s\n", time.Now())

		rawKeys, err := MarketDataSource.Keys("DEPTHx*")
		if err != nil {
			logger.Error(err)
			time.Sleep(5 * time.Minute)
			continue
		}

		// symbol list
		var tmpSymbols []string
		// update SymbolsMapExchanges, directly replace the order map
		newSME := make(map[string][]string)

		Memo.LockSymbolsMapExchanges.Lock()
		for _, rawKey := range rawKeys {
			keys := strings.Split(rawKey[7:], ".")
			tmpSymbols = append(tmpSymbols, keys[0])
			newSME[keys[0]] = append(newSME[keys[0]], keys[1])
		}
		Memo.SymbolsMapExchanges = newSME
		Memo.LockSymbolsMapExchanges.Unlock()

		// logger.Info("%v\n", Memo.SymbolsMapExchanges)

		// dummy trade
		tmpTrade := models.TRADE{
			Exchange:   "",
			Time:       "",
			TimeArrive: "",
			Direction:  "",
			LastPx:     -1,
			Qty:        -1,
		}
		// update SymbolsMapLastTrade, move trade data from the order map to the new map
		newSMT := make(map[string]models.TRADE)

		Memo.LockSymbolsMapLastTrade.Lock()
		for _, sym := range tmpSymbols {

			if val, ok := Memo.SymbolsMapLastTrade[sym]; ok {
				newSMT[sym] = val
			} else {
				tmpTrade.Symbol = sym
				newSMT[sym] = tmpTrade
			}

		}
		Memo.SymbolsMapLastTrade = newSMT
		Memo.LockSymbolsMapLastTrade.Unlock()

		// logger.Info("%v\n", Memo.SymbolsMapLastTrade)
		time.Sleep(2 * time.Hour)
	}

}

func SubscribeTrade() {
	psc, err := MarketDataSource.PSub("TRADEx*")
	if err != nil {
		logger.Error(err)
	}

	for {
		switch v := psc.Receive().(type) {
		case redis.Message:
			var t models.TRADE
			// logger.Info("%s: message: %s\n", v.Channel, v.Data)
			err := json.Unmarshal(v.Data, &t)
			if err != nil {
				logger.Error(err)
				logger.Debug(string(v.Data))
				break
			}
			Memo.LockSymbolsMapLastTrade.Lock()
			Memo.SymbolsMapLastTrade[t.Symbol] = t
			Memo.LockSymbolsMapLastTrade.Unlock()
		case redis.Subscription:
			// logger.Info("%s: %s %d\n", v.Channel, v.Kind, v.Count)
		case error:
			logger.Warn(v)
			logger.Warn("Reconnect in 1 sec.")
			psc, err = MarketDataSource.PSub("TRADEx*")
			if err != nil {
				logger.Error(err)
				break
			}
		}
	}
}

func UpdateAccount() {
	psc, err := TradingSource.PSub("*.Monitor")
	if err != nil {
		logger.Error(err)
	}

	for {
		switch v := psc.Receive().(type) {
		case redis.Message:
			// logger.Info("%s: message: %s\n", v.Channel, v.Data)

			var inAccounts []models.InAccount
			err := json.Unmarshal(v.Data, &inAccounts)
			if err != nil {
				logger.Error(err)
				logger.Debug(string(v.Data))
				break
			}

			var tmpAccounts []models.Account
			for _, a := range inAccounts {
				tmpAccounts = append(tmpAccounts, a.ToAccount())
			}

			var fairValue models.FairValue
			b, err := MarketDataSource.Get("FAIRVALUE")
			if err != nil {
				logger.Error(err)
				break
			}
			err = json.Unmarshal(b, &fairValue)
			if err != nil {
				logger.Error(err)
				logger.Debug(string(b))
				break
			}

			Memo.LockAccounts.RLock()
			for i, _ := range tmpAccounts {

				err := tmpAccounts[i].EstimateValue(fairValue)
				if err != nil {
					logger.Error(err)
					Memo.LockAccounts.RUnlock()
					break
				}

				for _, ain := range Memo.Accounts {
					err = tmpAccounts[i].CalculatePnl(ain)
					if err != nil {
						logger.Error(err)
						Memo.LockAccounts.RUnlock()
						break
					}
				}

				err = tmpAccounts[i].LogicalTotal()
				if err != nil {
					logger.Error(err)
					Memo.LockAccounts.RUnlock()
					break
				}
			}
			Memo.LockAccounts.RUnlock()

			tmpAccounts, err = models.PhysicalTotal(tmpAccounts)
			if err != nil {
				logger.Error(err)
				break
			}

			Memo.LockRealtimeAccounts.Lock()
			Memo.RealtimeAccounts = tmpAccounts
			Memo.LockRealtimeAccounts.Unlock()

		case redis.Subscription:
			// logger.Info("%s: %s %d\n", v.Channel, v.Kind, v.Count)
		case error:
			logger.Warn(v)
			logger.Warn("Reconnect in 1 sec.")
			psc, err = MarketDataSource.PSub("TRADEx*")
			if err != nil {
				logger.Error(err)
				break
			}
		}
	}
}

func UpdateGalaxy() {
	psc, err := GalaxySource.Sub("GalaxyDetail", "StrategySummary", "StrategyPosition", "StrategyAccount", "StrategyMarket", "StrategyUserDefine", "StrategyTrade", "StrategyOrder")
	if err != nil {
		logger.Error(err)
	}

	for {
		switch v := psc.Receive().(type) {
		case redis.Message:
			// logger.Debug(v.Channel)
			// logger.Debug(string(v.Data))
			switch v.Channel {
			case "GalaxyDetail":
				var s models.GalaxyStatus
				err := json.Unmarshal(v.Data, &s)
				if err != nil {
					logger.Error(err)
					logger.Debug(string(v.Data))
					break
				}

				Memo.LockGalaxyStatusMemo.Lock()
				Memo.GalaxyStatusMemo = s
				Memo.LockGalaxyStatusMemo.Unlock()

			case "StrategySummary":
				var summary models.StrategySummary
				err := json.Unmarshal(v.Data, &summary)
				if err != nil {
					logger.Error(err)
					logger.Debug(string(v.Data))
					break
				}

				Memo.LockGalaxyStatusMemo.Lock()
				if _, ok := Memo.StrategyMessageSetMap[summary.StrategyName]; !ok {
					Memo.StrategyMessageSetMap[summary.StrategyName] = models.MakeStrategyMessageSet(summary.StrategyName)
				}
				err = Memo.StrategyMessageSetMap[summary.StrategyName].InsertSummary(summary)
				Memo.LockGalaxyStatusMemo.Unlock()

				if err != nil {
					logger.Error(err)
					break
				}

			case "StrategyPosition":
				var position models.StrategyPosition
				err := json.Unmarshal(v.Data, &position)
				if err != nil {
					logger.Error(err)
					logger.Debug(string(v.Data))
					break
				}

				Memo.LockGalaxyStatusMemo.Lock()
				if _, ok := Memo.StrategyMessageSetMap[position.StrategyName]; !ok {
					Memo.StrategyMessageSetMap[position.StrategyName] = models.MakeStrategyMessageSet(position.StrategyName)
				}
				err = Memo.StrategyMessageSetMap[position.StrategyName].InsertPosition(position)
				Memo.LockGalaxyStatusMemo.Unlock()

				if err != nil {
					logger.Error(err)
					break
				}

			case "StrategyAccount":
				var account models.StrategyAccount
				err := json.Unmarshal(v.Data, &account)
				if err != nil {
					logger.Error(err)
					logger.Debug(string(v.Data))
					break
				}

				Memo.LockGalaxyStatusMemo.Lock()
				if _, ok := Memo.StrategyMessageSetMap[account.StrategyName]; !ok {
					Memo.StrategyMessageSetMap[account.StrategyName] = models.MakeStrategyMessageSet(account.StrategyName)
				}
				err = Memo.StrategyMessageSetMap[account.StrategyName].InsertAccount(account)
				Memo.LockGalaxyStatusMemo.Unlock()

				if err != nil {
					logger.Error(err)
					break
				}

			case "StrategyMarket":
				var market models.StrategyMarket
				err := json.Unmarshal(v.Data, &market)
				if err != nil {
					logger.Error(err)
					logger.Debug(string(v.Data))
					break
				}

				Memo.LockGalaxyStatusMemo.Lock()
				if _, ok := Memo.StrategyMessageSetMap[market.StrategyName]; !ok {
					Memo.StrategyMessageSetMap[market.StrategyName] = models.MakeStrategyMessageSet(market.StrategyName)
				}
				err = Memo.StrategyMessageSetMap[market.StrategyName].InsertMarket(market)
				Memo.LockGalaxyStatusMemo.Unlock()

				if err != nil {
					logger.Error(err)
					break
				}

			case "StrategyUserDefine":
				var userDefineInput models.StrategyUserDefineInput
				err := json.Unmarshal(v.Data, &userDefineInput)
				if err != nil {
					logger.Error(err)
					logger.Debug(string(v.Data))
					break
				}
				userDefine := userDefineInput.Transform()
				Memo.LockGalaxyStatusMemo.Lock()
				if _, ok := Memo.StrategyMessageSetMap[userDefine.StrategyName]; !ok {
					Memo.StrategyMessageSetMap[userDefine.StrategyName] = models.MakeStrategyMessageSet(userDefine.StrategyName)
				}
				err = Memo.StrategyMessageSetMap[userDefine.StrategyName].InsertUserDefine(userDefine)
				Memo.LockGalaxyStatusMemo.Unlock()

				if err != nil {
					logger.Error(err)
					break
				}

			case "StrategyTrade":
				var trade models.StrategyTrade
				err := json.Unmarshal(v.Data, &trade)
				if err != nil {
					logger.Error(err)
					logger.Debug(string(v.Data))
					break
				}

				Memo.LockGalaxyStatusMemo.Lock()
				if _, ok := Memo.StrategyMessageSetMap[trade.StrategyName]; !ok {
					Memo.StrategyMessageSetMap[trade.StrategyName] = models.MakeStrategyMessageSet(trade.StrategyName)
				}
				err = Memo.StrategyMessageSetMap[trade.StrategyName].InsertTrade(trade)
				Memo.LockGalaxyStatusMemo.Unlock()

				if err != nil {
					logger.Error(err)
					break
				}

			case "StrategyOrder":
				var order models.StrategyOrder
				err := json.Unmarshal(v.Data, &order)
				if err != nil {
					logger.Error(err)
					logger.Debug(string(v.Data))
					break
				}

				Memo.LockGalaxyStatusMemo.Lock()
				if _, ok := Memo.StrategyMessageSetMap[order.StrategyName]; !ok {
					Memo.StrategyMessageSetMap[order.StrategyName] = models.MakeStrategyMessageSet(order.StrategyName)
				}
				err = Memo.StrategyMessageSetMap[order.StrategyName].InsertOrder(order)
				Memo.LockGalaxyStatusMemo.Unlock()

				if err != nil {
					logger.Error(err)
					break
				}

			}

		case redis.Subscription:
			// logger.Info("%s: %s %d\n", v.Channel, v.Kind, v.Count)

		case error:
			logger.Warn(v)
			logger.Warn("Reconnect in 1 sec.")
			psc, err = MarketDataSource.PSub("TRADEx*")
			if err != nil {
				logger.Error(err)
				break
			}
		}
	}
}
