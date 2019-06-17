package server

import (
	"encoding/json"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/toolbox"
	"github.com/wonderivan/logger"
	"github.com/yushizhao/hubble/config"
	"github.com/yushizhao/hubble/models"
	"github.com/yushizhao/hubble/redis"
)

type MainController struct {
	beego.Controller
}

// @router /marketData/STATUS [get]
func (this *MainController) STATUS() {
	// this.Ctx.WriteString(models.MOCK_STATUS)

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

	// requestDump, err := httputil.DumpRequest(this.Ctx.Request, true)
	// if err != nil {
	// 	logger.Debug(err)
	// }
	// logger.Debug(string(requestDump))

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
}

func (this *MainController) MYORDERS() {
	this.Ctx.WriteString(models.MOCK_DEPTHx)
}

func (this *MainController) GALAXY() {
	this.Ctx.WriteString(models.MOCK_GALAXY)
}

func (this *MainController) SINGULARITY() {
	this.Ctx.WriteString(models.MOCK_SINGULARITY)
}

func (this *MainController) ACCOUNT() {
	this.Ctx.WriteString(models.MOCK_ACCOUNT)
}

// func (this *MainController) PORTIFOLIO() {
// 	this.Ctx.WriteString(models.MOCK_PORTIFOLIO)
// }

// var RedisClients map[string]*radix.Sentinel

var MarketDataSource *redis.Client
var TradingSource *redis.Client

func StartServer() {

	// RedisClients = redis.NewSentinels(models.MasterNames, config.Conf.Sentinels, config.Conf.SentinelPassword, config.Conf.ServerPassword)

	MarketDataSource = redis.NewClient(config.Conf.MarketData.Addr, config.Conf.MarketData.Pass, 3, 60)
	TradingSource = redis.NewClient(config.Conf.Trading.Pass, config.Conf.Trading.Pass, 3, 60)

	go UpdateFromDepth()
	go SubscribeTrade()
	go UpdateAccount()

	beego.BConfig.CopyRequestBody = true
	beego.BConfig.Listen.EnableHTTPS = true
	beego.BConfig.Listen.EnableHTTP = false
	beego.BConfig.Listen.HTTPSPort = config.Conf.Port
	beego.BConfig.Listen.HTTPSCertFile = "config/quant.crt"
	beego.BConfig.Listen.HTTPSKeyFile = "config/quant.key"

	beego.Router("/marketData/STATUS", &MainController{}, "get:STATUS")
	beego.Router("/marketData/TRADEx", &MainController{}, "get:TRADEx")
	beego.Router("/marketData/KLINE", &MainController{}, "post:KLINE")
	beego.Router("/marketData/DEPTHx", &MainController{}, "post:DEPTHx")
	beego.Router("/trading/STATUS", &MainController{}, "get:TSTATUS")
	beego.Router("/trading/MYORDERS", &MainController{}, "post:MYORDERS")
	beego.Router("/trading/GALAXY", &MainController{}, "get:GALAXY")
	beego.Router("/trading/SINGULARITY", &MainController{}, "get:SINGULARITY")
	beego.Router("/trading/ACCOUNT", &MainController{}, "get:ACCOUNT")
	// beego.Router("/trading/PORTIFOLIO", &MainController{}, "get:PORTIFOLIO")

	toolbox.StartTask()
	defer toolbox.StopTask()

	beego.Run()
}
