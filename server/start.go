package server

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/toolbox"
	"github.com/yushizhao/hubble/config"
	"github.com/yushizhao/hubble/redis"
)

// var RedisClients map[string]*radix.Sentinel

var MarketDataSource *redis.Client
var TradingSource *redis.Client

func StartServer() {

	// RedisClients = redis.NewSentinels(models.MasterNames, config.Conf.Sentinels, config.Conf.SentinelPassword, config.Conf.ServerPassword)

	MarketDataSource = redis.NewClient(config.Conf.MarketData.Addr, config.Conf.MarketData.Pass, 3, 60)
	TradingSource = redis.NewClient(config.Conf.Trading.Addr, config.Conf.Trading.Pass, 3, 60)

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
