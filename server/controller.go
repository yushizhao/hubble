package server

import (
	"encoding/json"

	"github.com/yushizhao/authenticator/gawrapper"

	"github.com/astaxie/beego"
	"github.com/wonderivan/logger"
	"github.com/yushizhao/hubble/models"
)

type MainController struct {
	beego.Controller
}

func (this *MainController) Options() {
	// requestDump, err := httputil.DumpRequest(this.Ctx.Request, true)
	// if err != nil {
	// 	logger.Debug(err)
	// }
	// logger.Debug(string(requestDump))
	this.Data["json"] = map[string]interface{}{"status": 200, "message": "ok"}
	this.ServeJSON()
}

func (this *MainController) SignUp() {
	code := this.GetString("InvitationCode")

	Memo.LockInvitationCode.RLock()
	if code != Memo.InvitationCode {
		Memo.LockInvitationCode.RUnlock()
		this.Data["json"] = map[string]interface{}{"status": 400, "message": "Invalid InvitationCode"}
		this.ServeJSON()
		return
	}
	Memo.LockInvitationCode.RUnlock()

	name := this.GetString("UserName")

	if name == "" {
		this.Data["json"] = map[string]interface{}{"status": 400, "message": "Empty UserName"}
		this.ServeJSON()
		return
	}

	Memo.LockNameMapUsers.Lock()
	if _, ok := Memo.NameMapUsers[name]; ok {
		Memo.LockNameMapUsers.Unlock()
		this.Data["json"] = map[string]interface{}{"status": 400, "message": "Existing UserName"}
		this.ServeJSON()
		return
	}

	secret := gawrapper.GenerateSecret()

	var u models.User
	u.Secret = secret
	u.Role = ""
	Memo.NameMapUsers[name] = u

	Memo.LockNameMapUsers.Unlock()

	this.Data["json"] = map[string]interface{}{"status": 200, "message": gawrapper.NewOTPAuth(name, secret, ISSUER)}
	this.ServeJSON()
}

// @router /marketData/STATUS [get]
func (this *MainController) STATUS() {
	// this.Ctx.WriteString(models.MOCK_STATUS)
	// return
	exchanges, err := MarketDataSource.HKeys("STATUS")
	if err != nil {
		logger.Error(err)
	}

	exchangesMapStatus := make(map[string]models.STATUS)
	for _, ex := range exchanges {
		b, err := MarketDataSource.HGet("STATUS", ex)
		if err != nil {
			logger.Error(err)
		}
		var tmpStatus models.STATUS
		err = json.Unmarshal(b, &tmpStatus)
		if err != nil {
			logger.Error(err)
		}
		exchangesMapStatus[ex] = tmpStatus
	}

	jsonBytes, err := json.Marshal(exchangesMapStatus)

	if err != nil {
		logger.Error(err)
	}

	this.Ctx.ResponseWriter.Write(jsonBytes)
}

// @router /marketData/TRADEx [get]
func (this *MainController) TRADEx() {
	// this.Ctx.WriteString(models.MOCK_TRADEx)
	// return
	Memo.LockSymbolsMapLastTrade.RLock()
	jsonBytes, err := json.Marshal(Memo.SymbolsMapLastTrade)
	Memo.LockSymbolsMapLastTrade.RUnlock()

	if err != nil {
		logger.Error(err)
	}

	this.Ctx.ResponseWriter.Write(jsonBytes)
}

func (this *MainController) KLINE() {
	// this.Ctx.WriteString(models.MOCK_KLINE)
	// return
	ob := make(map[string]string)
	err := json.Unmarshal(this.Ctx.Input.RequestBody, &ob)
	if err != nil {
		logger.Debug(err)
	}

	sym := ob["Symbol"]
	ex := ob["Exchange"]

	b, err := MarketDataSource.HGet("KLINE", sym+"."+ex)
	if err != nil {
		logger.Error(err)
	}
	this.Ctx.ResponseWriter.Write(b)
}

func (this *MainController) DEPTHx() {
	// this.Ctx.WriteString(models.MOCK_DEPTHx)
	// return
	ob := make(map[string]string)
	err := json.Unmarshal(this.Ctx.Input.RequestBody, &ob)
	if err != nil {
		logger.Debug(err)
	}

	sym := ob["Symbol"]

	var depths []models.OutDEPTH
	for _, ex := range Memo.SymbolsMapExchanges[sym] {

		key := "DEPTHx|" + sym + "." + ex
		b, err := MarketDataSource.Get(key)
		if err != nil {
			logger.Error(err)
		}

		var tmpDepth models.DEPTH
		err = json.Unmarshal(b, &tmpDepth)
		if err != nil {
			logger.Error(err)
		}

		tmpOutDepth, err := tmpDepth.Output()
		if err != nil {
			logger.Error(err)
		}

		depths = append(depths, tmpOutDepth)
	}

	jsonBytes, err := json.Marshal(depths)

	if err != nil {
		logger.Error(err)
	}

	this.Ctx.ResponseWriter.Write(jsonBytes)
}

func (this *MainController) TSTATUS() {
	this.Ctx.WriteString(models.MOCK_TSTATUS)
	// return
}

func (this *MainController) MYORDERS() {
	this.Ctx.WriteString(models.MOCK_DEPTHx)
	// return
}

func (this *MainController) GALAXY() {
	this.Ctx.WriteString(models.MOCK_GALAXY)
	// return
}

func (this *MainController) SINGULARITY() {
	this.Ctx.WriteString(models.MOCK_SINGULARITY)
	// return
}

func (this *MainController) ACCOUNTNAME() {
	// this.Ctx.WriteString(models.MOCK_ACCOUNTNAME)
	// return
	tmp := make(map[string][]string)
	Memo.LockRealtimeAccounts.RLock()
	for _, a := range Memo.RealtimeAccounts {
		tmp["PhysicalAccount"] = append(tmp["PhysicalAccount"], a.PhysicalAccount)
	outer:
		for _, p := range a.LogicalAccount {
			for _, n := range tmp["LogicalAccount"] {
				if n == p.ClientCode {
					continue outer
				}
			}
			tmp["LogicalAccount"] = append(tmp["LogicalAccount"], p.ClientCode)
		}
	}
	Memo.LockRealtimeAccounts.RUnlock()
	b, err := json.Marshal(tmp)
	if err != nil {
		logger.Error(err)
	}
	logger.Info("%q", b)
	this.Ctx.ResponseWriter.Write(b)
}

func (this *MainController) ACCOUNT() {
	// this.Ctx.WriteString(models.MOCK_ACCOUNT)
	// return
	Memo.LockRealtimeAccounts.RLock()
	b, err := json.Marshal(Memo.RealtimeAccounts)
	Memo.LockRealtimeAccounts.RUnlock()

	if err != nil {
		logger.Error(err)
	}
	this.Ctx.ResponseWriter.Write(b)
}

func (this *MainController) GSTATUS() {
	// this.Ctx.WriteString(models.MOCK_GSTATUS)
	// return
	Memo.LockGalaxyStatusMemo.RLock()
	b, err := json.Marshal(Memo.GalaxyStatusMemo)
	Memo.LockGalaxyStatusMemo.RUnlock()

	if err != nil {
		logger.Error(err)
	}
	this.Ctx.ResponseWriter.Write(b)
}

func (this *MainController) STRATEGY() {
	// this.Ctx.WriteString(models.MOCK_STRATEGY)
	// return
	Memo.LockStrategyStatusMap.RLock()
	b, err := json.Marshal(Memo.StrategyStatusMap)
	Memo.LockStrategyStatusMap.RUnlock()

	if err != nil {
		logger.Error(err)
	}
	this.Ctx.ResponseWriter.Write(b)
}

// func (this *MainController) PORTIFOLIO() {
// 	this.Ctx.WriteString(models.MOCK_PORTIFOLIO)
// }
