package server

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/toolbox"
	"github.com/yushizhao/authenticator/boltwrapper"
	"github.com/yushizhao/hubble/config"
	"github.com/yushizhao/hubble/ding"
	"github.com/yushizhao/hubble/rediswrapper"
)

// var RedisClients map[string]*radix.Sentinel

var MarketDataSource *rediswrapper.Client
var TradingSource *rediswrapper.Client
var GalaxySource *rediswrapper.Client

var InvitationDing ding.Ding

func StartServer() {

	boltwrapper.InitDB()
	// RedisClients = redis.NewSentinels(models.MasterNames, config.Conf.Sentinels, config.Conf.SentinelPassword, config.Conf.ServerPassword)

	MarketDataSource = rediswrapper.NewClient(config.Conf.MarketData.Addr, config.Conf.MarketData.Pass, 3, 60)
	TradingSource = rediswrapper.NewClient(config.Conf.Trading.Addr, config.Conf.Trading.Pass, 3, 60)
	GalaxySource = rediswrapper.NewClient(config.Conf.Galaxy.Addr, config.Conf.Galaxy.Pass, 3, 60)

	go UpdateFromDepth()
	go SubscribeTrade()
	go UpdateAccount()
	go UpdateGalaxy()

	beego.BConfig.CopyRequestBody = true
	beego.BConfig.Listen.EnableHTTPS = true
	beego.BConfig.Listen.EnableHTTP = false
	beego.BConfig.Listen.HTTPSPort = config.Conf.Port
	beego.BConfig.Listen.HTTPSCertFile = "config/quant.crt"
	beego.BConfig.Listen.HTTPSKeyFile = "config/quant.key"
	beego.BConfig.WebConfig.Session.SessionOn = true

	beego.InsertFilter("*", beego.BeforeRouter, FilterCrossDomain)
	beego.InsertFilter("/trading/*", beego.BeforeRouter, FilterJWT)

	beego.Router("*", &MainController{}, "options:Options")
	beego.Router("/user/Invite", &MainController{}, "get:Invite")
	beego.Router("/user/SignUp", &MainController{}, "post:SignUp")
	beego.Router("/user/Login", &MainController{}, "post:Login")
	beego.Router("/user/List", &MainController{}, "get:List")
	beego.Router("/user/Retire", &MainController{}, "get:Retire")
	beego.Router("/marketData/STATUS", &MainController{}, "get:STATUS")
	beego.Router("/marketData/TRADEx", &MainController{}, "get:TRADEx")
	beego.Router("/marketData/KLINE", &MainController{}, "post:KLINE")
	beego.Router("/marketData/DEPTHx", &MainController{}, "post:DEPTHx")
	beego.Router("/trading/STATUS", &MainController{}, "get:TSTATUS")
	beego.Router("/trading/MYORDERS", &MainController{}, "post:MYORDERS")
	beego.Router("/trading/GALAXY", &MainController{}, "get:GALAXY")
	beego.Router("/trading/SINGULARITY", &MainController{}, "get:SINGULARITY")
	beego.Router("/trading/ACCOUNTNAME", &MainController{}, "get:ACCOUNTNAME")
	beego.Router("/trading/ACCOUNT", &MainController{}, "get:ACCOUNT")
	beego.Router("/galaxy/STATUS", &MainController{}, "get:GSTATUS")
	beego.Router("/galaxy/STRATEGY", &MainController{}, "get:STRATEGY")

	// beego.Router("/trading/PORTIFOLIO", &MainController{}, "get:PORTIFOLIO")

	TaskWriteReport()
	toolbox.StartTask()
	defer toolbox.StopTask()

	beego.Run()
}
