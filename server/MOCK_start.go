package server

import (
	"github.com/astaxie/beego"
	"github.com/yushizhao/hubble/config"
)

func MOCK_StartServer() {

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

	beego.Router("*", &MOCK_AdminController{}, "options:Options")
	beego.Router("/root/Invite", &MOCK_AdminController{}, "get:Invite")
	beego.Router("/root/List", &MOCK_AdminController{}, "get:List")
	beego.Router("/root/Retire", &MOCK_AdminController{}, "get:Retire")
	beego.Router("/user/SignUp", &MOCK_AdminController{}, "post:SignUp")
	beego.Router("/user/Login", &MOCK_AdminController{}, "post:Login")

	beego.Router("/marketData/STATUS", &MOCK_MarketDataController{}, "get:STATUS")
	beego.Router("/marketData/TRADEx", &MOCK_MarketDataController{}, "get:TRADEx")
	beego.Router("/marketData/KLINE", &MOCK_MarketDataController{}, "post:KLINE")
	beego.Router("/marketData/DEPTHx", &MOCK_MarketDataController{}, "post:DEPTHx")

	beego.Router("/trading/STATUS", &MOCK_TradingController{}, "get:STATUS")
	beego.Router("/trading/MYORDERS", &MOCK_TradingController{}, "post:MYORDERS")
	beego.Router("/trading/ACCOUNTNAME", &MOCK_TradingController{}, "get:ACCOUNTNAME")
	beego.Router("/trading/ACCOUNT", &MOCK_TradingController{}, "get:ACCOUNT")
	// beego.Router("/trading/PORTIFOLIO", &MainController{}, "get:PORTIFOLIO")

	beego.Router("/galaxy/GalaxyDetail", &MOCK_GalaxyController{}, "get:GalaxyDetail")
	beego.Router("/galaxy/StrategyList", &MOCK_GalaxyController{}, "get:StrategyList")
	beego.Router("/galaxy/StrategySummary", &MOCK_GalaxyController{}, "post:StrategySummary")
	beego.Router("/galaxy/StrategyMarket", &MOCK_GalaxyController{}, "post:StrategyMarket")
	beego.Router("/galaxy/StrategyUserDefine", &MOCK_GalaxyController{}, "post:StrategyUserDefine")
	beego.Router("/galaxy/StrategyTrade", &MOCK_GalaxyController{}, "post:StrategyTrade")
	beego.Router("/galaxy/StrategyOrder", &MOCK_GalaxyController{}, "post:StrategyOrder")

	beego.Run()
}
