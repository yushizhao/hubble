package server

import (
	"bytes"
	"crypto/rand"
	"encoding/base32"
	"encoding/gob"
	"encoding/json"
	"time"

	"github.com/yushizhao/authenticator/boltwrapper"
	"github.com/yushizhao/authenticator/jwtwrapper"

	"github.com/yushizhao/authenticator/gawrapper"

	"github.com/astaxie/beego"
	"github.com/wonderivan/logger"
	"github.com/yushizhao/hubble/config"
	"github.com/yushizhao/hubble/models"
)

type AdminController struct {
	beego.Controller
}

func (this *AdminController) Options() {
	// requestDump, err := httputil.DumpRequest(this.Ctx.Request, true)
	// if err != nil {
	// 	logger.Debug(err)
	// }
	// logger.Debug(string(requestDump))
	this.Ctx.WriteString(ISSUER)
}

func (this *AdminController) Invite() {

	b := make([]byte, 10)
	rand.Read(b)

	invitationCode := base32.StdEncoding.EncodeToString(b)[:6]
	exp := time.Now().Unix() + config.Server.InvitationExpire

	Memo.LockInvitationCode.Lock()
	Memo.InvitationCode[invitationCode] = exp
	Memo.LockInvitationCode.Unlock()

	this.Ctx.WriteString(invitationCode)
	return
}

func (this *AdminController) List() {

	userMapRaw := boltwrapper.UserDB.ListUser()

	buf := new(bytes.Buffer)
	decoder := gob.NewDecoder(buf)

	userMap := make(map[string]models.User)
	for k, v := range userMapRaw {
		var tmp models.User
		buf.Reset()
		buf.Write(v)
		err := decoder.Decode(&tmp)
		if err != nil {
			logger.Warn("Wrong User Data: %s", k)
		}
		userMap[k] = tmp
	}

	jsonBytes, err := json.Marshal(userMap)

	if err != nil {
		logger.Error(err)
	}

	this.Ctx.ResponseWriter.Write(jsonBytes)
}

func (this *AdminController) Retire() {

	name := this.GetString("UserName")

	err := boltwrapper.UserDB.DelUser(name)

	if err != nil {
		this.Ctx.WriteString(err.Error())
		return
	}

	this.Ctx.WriteString(name)
}

func (this *AdminController) SignUp() {

	Memo.LockInvitationCode.Lock()
	defer Memo.LockInvitationCode.Unlock()

	// clean outdated code
	now := time.Now().Unix()
	for k, v := range Memo.InvitationCode {
		if v < now {
			delete(Memo.InvitationCode, k)
		}
	}

	// read post body
	ob := make(map[string]string)
	err := json.Unmarshal(this.Ctx.Input.RequestBody, &ob)
	if err != nil {
		logger.Debug(err)
	}

	yourCode, ok := ob["InvitationCode"]
	if !ok {
		this.Data["json"] = map[string]interface{}{"status": 400, "message": "Missing InvitationCode"}
		this.ServeJSON()
		return
	}

	name, ok := ob["UserName"]

	if !ok {
		this.Data["json"] = map[string]interface{}{"status": 400, "message": "Missing UserName"}
		this.ServeJSON()
		return
	}

	// check code
	if _, ok := Memo.InvitationCode[yourCode]; !ok {
		this.Data["json"] = map[string]interface{}{"status": 400, "message": "Invalid Invitation Code"}
		this.ServeJSON()
		return
	}

	userBytes := boltwrapper.UserDB.GetUser(name)
	if userBytes != nil {
		this.Data["json"] = map[string]interface{}{"status": 400, "message": "UserName Exists"}
		this.ServeJSON()
		return
	}

	secret := gawrapper.GenerateSecret()

	var u models.User
	u.Secret = secret
	u.Role = ""

	buf := new(bytes.Buffer)
	encoder := gob.NewEncoder(buf)
	err = encoder.Encode(u)

	if err != nil {
		this.Data["json"] = map[string]interface{}{"status": 500, "message": "Internal Server Error"}
		this.ServeJSON()
		logger.Error(err)
		return
	}

	err = boltwrapper.UserDB.SetUser(name, buf.Bytes())
	if err != nil {
		this.Data["json"] = map[string]interface{}{"status": 500, "message": "Internal Server Error"}
		this.ServeJSON()
		logger.Error(err)
		return
	}

	delete(Memo.InvitationCode, yourCode)

	this.Data["json"] = map[string]interface{}{"status": 200, "message": gawrapper.NewOTPAuth(name, secret, ISSUER)}
	this.ServeJSON()
	return
}

func (this *AdminController) Login() {

	ob := make(map[string]string)
	err := json.Unmarshal(this.Ctx.Input.RequestBody, &ob)
	if err != nil {
		logger.Debug(err)
	}

	name, ok := ob["UserName"]
	if !ok {
		this.Data["json"] = map[string]interface{}{"status": 400, "message": "Missing UserName"}
		this.ServeJSON()
		return
	}

	yourCode, ok := ob["AuthenticationCode"]
	if !ok {
		this.Data["json"] = map[string]interface{}{"status": 400, "message": "Missing AuthenticationCode"}
		this.ServeJSON()
		return
	}

	userBytes := boltwrapper.UserDB.GetUser(name)
	if userBytes == nil {
		this.Data["json"] = map[string]interface{}{"status": 400, "message": "UserName Not Exists"}
		this.ServeJSON()
		return
	}

	var u models.User
	buf := new(bytes.Buffer)
	buf.Write(userBytes)
	decoder := gob.NewDecoder(buf)
	err = decoder.Decode(&u)
	if err != nil {
		this.Data["json"] = map[string]interface{}{"status": 500, "message": "Internal Server Error"}
		this.ServeJSON()
		logger.Error(err)
		return
	}

	verified, err := gawrapper.VerifyTOTP(u.Secret, yourCode)
	if err != nil {
		this.Data["json"] = map[string]interface{}{"status": 500, "message": "Internal Server Error"}
		this.ServeJSON()
		logger.Error(err)
		return
	}
	if !verified {
		this.Data["json"] = map[string]interface{}{"status": 400, "message": "Invalid AuthenticationCode"}
		this.ServeJSON()
		return
	}

	claims := map[string]interface{}{
		"Name": name,
		"Role": u.Role,
	}

	token, err := jwtwrapper.IssueTokenStrWithExp(claims, config.Server.JWTSecret, config.Server.JWTExpire)
	if err != nil {
		this.Data["json"] = map[string]interface{}{"status": 500, "message": "Internal Server Error"}
		this.ServeJSON()
		logger.Error(err)
		return
	}

	this.Data["json"] = map[string]interface{}{"status": 200, "message": token}
	this.ServeJSON()
	return
}

type MarketDataController struct {
	beego.Controller
}

// @router /marketData/STATUS [get]
func (this *MarketDataController) STATUS() {
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
func (this *MarketDataController) TRADEx() {
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

func (this *MarketDataController) KLINE() {
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

func (this *MarketDataController) DEPTHx() {
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

type TradingController struct {
	beego.Controller
}

func (this *TradingController) STATUS() {
	this.Ctx.WriteString(models.MOCK_TSTATUS)
	// return
}

func (this *TradingController) MYORDERS() {
	this.Ctx.WriteString(models.MOCK_DEPTHx)
	// return
}

func (this *TradingController) ACCOUNTNAME() {
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

func (this *TradingController) ACCOUNT() {
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

// func (this *TradingController) PORTIFOLIO() {
// 	this.Ctx.WriteString(models.MOCK_PORTIFOLIO)
// }

type GalaxyController struct {
	beego.Controller
}

func (this *GalaxyController) GalaxyDetail() {
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

func (this *GalaxyController) StrategyList() {
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

func (this *GalaxyController) StrategySummary() {
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

func (this *GalaxyController) StrategyUserDefine() {
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

func (this *GalaxyController) StrategyTrade() {
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

func (this *GalaxyController) StrategyOrder() {
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
