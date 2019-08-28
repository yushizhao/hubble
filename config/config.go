package config

import (
	"encoding/json"
	"io/ioutil"

	"github.com/wonderivan/logger"
)

type SConfig struct {
	MOCK bool

	Port int

	Trading    AddrPass
	MarketData AddrPass
	OMS        AddrPass
	Galaxy     AddrPass

	ReportSchedule string

	Ding string

	JWTSecret string
	JWTExpire int64

	InvitationExpire int64

	RootKey string

	StrategyExpire  int64
	UserDefineLimit int
	TradeLimit      int
	OrderLimit      int
}

var Server SConfig

type AddrPass struct {
	Addr string
	Pass string
}

func ReadConfig() error {

	data, err := ioutil.ReadFile("hubble.json")
	if err != nil {
		logger.Warn(err)
	}

	v := SConfig{}
	err = json.Unmarshal(data, &v)
	if err != nil {
		return err
	}

	Server = v
	return nil
}
