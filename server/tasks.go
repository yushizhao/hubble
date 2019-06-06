package server

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
