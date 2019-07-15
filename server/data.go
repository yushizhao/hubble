package server

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	"github.com/wonderivan/logger"
	"github.com/yushizhao/hubble/models"
)

type MEMO struct {
	InvitationCode map[string]int64
	// Exchanges           []string
	// Symbols             []string
	SymbolsMapExchanges map[string][]string
	SymbolsMapLastTrade map[string]models.TRADE
	Accounts            []models.Account
	RealtimeAccounts    []models.Account
	GalaxyStatusMemo    models.GalaxyStatus
	StrategyStatusMap   map[string]models.StrategyStatus

	LockInvitationCode sync.RWMutex // Write in POST /user/SignUp & /user/Invite, Read in POST /user/SignUp
	// LockExchanges           sync.RWMutex // Write in marketData/STATUS call, read in nowhere
	// LockSymbols             sync.RWMutex // Write in UpdateFromDepth, read in nowhere
	LockSymbolsMapExchanges sync.RWMutex // Write in UpdateFromDepth, read in DEPTH call
	LockSymbolsMapLastTrade sync.RWMutex // Write in UpdateFromDepth & SubscribeTrade, read in TRADE call
	LockAccounts            sync.RWMutex // Write in TaskWriteReport & StartServer, read in UpdateAccount
	LockRealtimeAccounts    sync.RWMutex // Write in UpdateAccount, read in ACCOUNT call & TaskWriteReport
	LockGalaxyStatusMemo    sync.RWMutex // Write in UpdateGalaxy, read in GET /galaxy/STATUS
	LockStrategyStatusMap   sync.RWMutex // Write in UpdateGalaxy, read in GET /galaxy/STRATEGY
}

const ISSUER = "Hubble"

var Memo MEMO

func init() {
	Memo.SymbolsMapExchanges = make(map[string][]string)
	Memo.SymbolsMapLastTrade = make(map[string]models.TRADE)
	Memo.StrategyStatusMap = make(map[string]models.StrategyStatus)
	Memo.InvitationCode = make(map[string]int64)

	data, err := ioutil.ReadFile("account.json")
	if err != nil {
		logger.Warn(err)
	}
	// accounts should have Value
	err = json.Unmarshal(data, &Memo.Accounts)
	if err != nil {
		logger.Warn(err)
	}

	Memo.RealtimeAccounts = Memo.Accounts
	logger.Info(&Memo.Accounts)
}
