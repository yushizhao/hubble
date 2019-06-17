package server

import (
	"sync"

	"github.com/yushizhao/hubble/models"
)

type MEMO struct {
	// Exchanges           []string
	// Symbols             []string
	SymbolsMapExchanges map[string][]string
	SymbolsMapLastTrade map[string]models.TRADE
	Accounts            []models.Account
	RealtimeAccounts    []models.Account

	// LockExchanges           sync.RWMutex // Write in marketData/STATUS call, read in nowhere
	// LockSymbols             sync.RWMutex // Write in UpdateFromDepth, read in nowhere
	LockSymbolsMapExchanges sync.RWMutex // Write in UpdateFromDepth, read in DEPTH call
	LockSymbolsMapLastTrade sync.RWMutex // Write in UpdateFromDepth & SubscribeTrade, read in TRADE call
	LockAccounts            sync.RWMutex // Write in TaskWriteReport & StartServer, read in UpdateAccount
	LockRealtimeAccounts    sync.RWMutex // Write in UpdateAccount, read in ACCOUNT call & TaskWriteReport
}

var Memo MEMO

func init() {
	Memo.SymbolsMapExchanges = make(map[string][]string)
	Memo.SymbolsMapLastTrade = make(map[string]models.TRADE)
}
