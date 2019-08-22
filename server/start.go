package server

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/toolbox"
	"github.com/yushizhao/authenticator/boltwrapper"
	"github.com/yushizhao/hubble/config"
	"github.com/yushizhao/hubble/ding"
	"github.com/yushizhao/hubble/rediswrapper"
)

var MarketDataSource *rediswrapper.Client
var TradingSource *rediswrapper.Client
var GalaxySource *rediswrapper.Client

var InvitationDing ding.Ding

func StartServer() {

	boltwrapper.InitDB()

	MarketDataSource = rediswrapper.NewClient(config.Server.MarketData.Addr, config.Server.MarketData.Pass, 3, 60)
	TradingSource = rediswrapper.NewClient(config.Server.Trading.Addr, config.Server.Trading.Pass, 3, 60)
	GalaxySource = rediswrapper.NewClient(config.Server.Galaxy.Addr, config.Server.Galaxy.Pass, 3, 60)

	go UpdateFromDepth()
	go SubscribeTrade()
	go UpdateAccount()
	go UpdateGalaxy()

	beego.BConfig.CopyRequestBody = true
	beego.BConfig.Listen.EnableHTTPS = true
	beego.BConfig.Listen.EnableHTTP = false
	beego.BConfig.Listen.HTTPSPort = config.Server.Port
	beego.BConfig.Listen.HTTPSCertFile = "config/quant.crt"
	beego.BConfig.Listen.HTTPSKeyFile = "config/quant.key"
	beego.BConfig.WebConfig.Session.SessionOn = true

	beego.InsertFilter("*", beego.BeforeRouter, FilterCrossDomain)
	// beego.InsertFilter("/trading/*", beego.BeforeRouter, FilterJWT)
	beego.InsertFilter("/root/*", beego.BeforeRouter, FilterRootTOTP)

	beego.Router("*", &MainController{}, "options:Options")
	beego.Router("/root/Invite", &MainController{}, "get:Invite")
	beego.Router("/root/List", &MainController{}, "get:List")
	beego.Router("/root/Retire", &MainController{}, "get:Retire")
	beego.Router("/user/SignUp", &MainController{}, "post:SignUp")
	beego.Router("/user/Login", &MainController{}, "post:Login")
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
