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

type MOCK_AdminController struct {
	beego.Controller
}

func (this *MOCK_AdminController) Options() {
	// requestDump, err := httputil.DumpRequest(this.Ctx.Request, true)
	// if err != nil {
	// 	logger.Debug(err)
	// }
	// logger.Debug(string(requestDump))
	this.Ctx.WriteString(ISSUER)
}

func (this *MOCK_AdminController) Invite() {

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

func (this *MOCK_AdminController) List() {

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

func (this *MOCK_AdminController) Retire() {

	name := this.GetString("UserName")

	err := boltwrapper.UserDB.DelUser(name)

	if err != nil {
		this.Ctx.WriteString(err.Error())
		return
	}

	this.Ctx.WriteString(name)
}

func (this *MOCK_AdminController) SignUp() {

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

func (this *MOCK_AdminController) Login() {

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

type MOCK_MarketDataController struct {
	beego.Controller
}

func (this *MOCK_MarketDataController) STATUS() {
	this.Ctx.WriteString(models.MOCK_STATUS)
	return
}

func (this *MOCK_MarketDataController) TRADEx() {
	this.Ctx.WriteString(models.MOCK_TRADEx)
	return
}

func (this *MOCK_MarketDataController) KLINE() {
	this.Ctx.WriteString(models.MOCK_KLINE)
	return
}

func (this *MOCK_MarketDataController) DEPTHx() {
	this.Ctx.WriteString(models.MOCK_DEPTHx)
	return
}

type MOCK_TradingController struct {
	beego.Controller
}

func (this *MOCK_TradingController) STATUS() {
	this.Ctx.WriteString(models.MOCK_TSTATUS)
	return
}

func (this *MOCK_TradingController) MYORDERS() {
	this.Ctx.WriteString(models.MOCK_DEPTHx)
	return
}

func (this *MOCK_TradingController) ACCOUNTNAME() {
	this.Ctx.WriteString(models.MOCK_ACCOUNTNAME)
	return
}

func (this *MOCK_TradingController) ACCOUNT() {
	this.Ctx.WriteString(models.MOCK_ACCOUNT)
	return
}

type MOCK_GalaxyController struct {
	beego.Controller
}

func (this *MOCK_GalaxyController) GalaxyDetail() {
	this.Ctx.WriteString(models.MOCK_GSTATUS)
	return
}

func (this *MOCK_GalaxyController) StrategyList() {
	b, _ := json.Marshal(models.MOCK_StrategyList)
	this.Ctx.ResponseWriter.Write(b)
	return
}

func (this *MOCK_GalaxyController) StrategySummary() {
	ob := make(map[string]string)
	err := json.Unmarshal(this.Ctx.Input.RequestBody, &ob)
	if err != nil {
		logger.Debug(err)
	}

	set := models.MOCK_MakeStrategyMessageSet(ob["StrategyName"])
	b, err := json.Marshal(set.Summary)

	if err != nil {
		logger.Error(err)
	}
	this.Ctx.ResponseWriter.Write(b)
	return
}

func (this *MOCK_GalaxyController) StrategyMarket() {
	ob := make(map[string]string)
	err := json.Unmarshal(this.Ctx.Input.RequestBody, &ob)
	if err != nil {
		logger.Debug(err)
	}

	set := models.MOCK_MakeStrategyMessageSet(ob["StrategyName"])
	b, err := json.Marshal(set.Market)

	if err != nil {
		logger.Error(err)
	}
	this.Ctx.ResponseWriter.Write(b)
	return
}

func (this *MOCK_GalaxyController) StrategyUserDefine() {
	ob := make(map[string]string)
	err := json.Unmarshal(this.Ctx.Input.RequestBody, &ob)
	if err != nil {
		logger.Debug(err)
	}

	set := models.MOCK_MakeStrategyMessageSet(ob["StrategyName"])
	b, err := json.Marshal(set.UserDefine)

	if err != nil {
		logger.Error(err)
	}
	this.Ctx.ResponseWriter.Write(b)
	return
}

func (this *MOCK_GalaxyController) StrategyTrade() {
	ob := make(map[string]string)
	err := json.Unmarshal(this.Ctx.Input.RequestBody, &ob)
	if err != nil {
		logger.Debug(err)
	}

	set := models.MOCK_MakeStrategyMessageSet(ob["StrategyName"])
	b, err := json.Marshal(set.Trade)

	if err != nil {
		logger.Error(err)
	}
	this.Ctx.ResponseWriter.Write(b)
	return
}

func (this *MOCK_GalaxyController) StrategyOrder() {
	ob := make(map[string]string)
	err := json.Unmarshal(this.Ctx.Input.RequestBody, &ob)
	if err != nil {
		logger.Debug(err)
	}

	set := models.MOCK_MakeStrategyMessageSet(ob["StrategyName"])
	b, err := json.Marshal(set.Order)

	if err != nil {
		logger.Error(err)
	}
	this.Ctx.ResponseWriter.Write(b)
	return
}
