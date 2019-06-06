package config

import (
	"encoding/xml"
	"io/ioutil"
	"os"
)

type SConfig struct {
	XMLName xml.Name `xml:"config"` // 指定最外层的标签为config

	MarketDataRedis SRedis `xml:"MarketDataRedis"`
	TradingRedis    SRedis `xml:"TradingRedis"`

	HTTPPort int `xml:"HTTPPort"`

	EnableHTTPS bool `xml:"EnableHTTPS"`
}

type SRedis struct {
	Url string `xml:"url"`
	Pwd string `xml:"pwd"`
}

var Conf SConfig

func ReadXml() error {
	file, err := os.Open("hubble.xml") // For read access.
	if err != nil {
		//fmt.Printf("error: %v", err)
		return err
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		//fmt.Printf("error: %v", err)
		return err
	}
	v := SConfig{}
	err = xml.Unmarshal(data, &v)
	if err != nil {
		//fmt.Printf("error: %v", err)
		return err
	}

	//fmt.Println(v.VecSymbols, v.ApiNames)
	Conf = v
	return nil
}
