package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

type ConfigFile struct {
	PluginsList  string   `json:"plugins_list"`
	Repositories []string `json:"repositories"`
	TempDir      string   `json:"temp_dir"`
	DataDir      string   `json:"data_dir"`
}

var Config ConfigFile

func ReadConfig() {
	raw, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal("Error reading config file: ", err)
	}
	Config = ConfigFile{}
	err = json.Unmarshal(raw, &Config)
	if err != nil {
		log.Fatal("Error unmarshaling config file: ", err)
	}
}
