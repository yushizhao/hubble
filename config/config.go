package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type SConfig struct {
	Port int

	Sentinels        []string
	SentinelPassword string
	ServerPassword   string
}

var Conf SConfig

func ReadConfig() error {
	file, err := os.Open("hubble.json")
	if err != nil {
		return err
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	v := SConfig{}
	err = json.Unmarshal(data, &v)
	if err != nil {
		return err
	}

	Conf = v
	return nil
}
