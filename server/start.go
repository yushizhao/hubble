package server

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/toolbox"
	"github.com/yushizhao/hubble/config"
	"github.com/yushizhao/hubble/rediswrapper"
)

func StartServer() {

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

	beego.Router("*", &AdminController{}, "options:Options")
	beego.Router("/root/Invite", &AdminController{}, "get:Invite")
	beego.Router("/root/List", &AdminController{}, "get:List")
	beego.Router("/root/Retire", &AdminController{}, "get:Retire")
	beego.Router("/user/SignUp", &AdminController{}, "post:SignUp")
	beego.Router("/user/Login", &AdminController{}, "post:Login")

	beego.Router("/marketData/STATUS", &MarketDataController{}, "get:STATUS")
	beego.Router("/marketData/TRADEx", &MarketDataController{}, "get:TRADEx")
	beego.Router("/marketData/KLINE", &MarketDataController{}, "post:KLINE")
	beego.Router("/marketData/DEPTHx", &MarketDataController{}, "post:DEPTHx")

	beego.Router("/trading/STATUS", &TradingController{}, "get:STATUS")
	beego.Router("/trading/MYORDERS", &TradingController{}, "post:MYORDERS")
	beego.Router("/trading/ACCOUNTNAME", &TradingController{}, "get:ACCOUNTNAME")
	beego.Router("/trading/ACCOUNT", &TradingController{}, "get:ACCOUNT")
	// beego.Router("/trading/PORTIFOLIO", &MainController{}, "get:PORTIFOLIO")

	beego.Router("/galaxy/GalaxyDetail", &GalaxyController{}, "get:GalaxyDetail")
	beego.Router("/galaxy/StrategySummary", &GalaxyController{}, "get:StrategySummary")
	beego.Router("/galaxy/StrategyMarket", &GalaxyController{}, "post:StrategyMarket")
	beego.Router("/galaxy/StrategyUserDefine", &GalaxyController{}, "post:StrategyUserDefine")
	beego.Router("/galaxy/StrategyTrade", &GalaxyController{}, "post:StrategyTrade")
	beego.Router("/galaxy/StrategyOrder", &GalaxyController{}, "post:StrategyOrder")

	TaskWriteReport()
	toolbox.StartTask()
	defer toolbox.StopTask()

	beego.Run()
}
