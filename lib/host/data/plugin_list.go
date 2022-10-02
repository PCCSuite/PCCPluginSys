package data

import (
	"encoding/json"
	"errors"
	"log"
	"os"

	"github.com/PCCSuite/PCCPluginSys/lib/host/config"
)

var PluginList pluginList = pluginList{
	Plugins: []pluginListed{
		pluginListed{
			Name:     "ExamplePluginName",
			Priority: 100,
			Enabled:  false,
		},
	},
}

type pluginList struct {
	Plugins []pluginListed `json:"plugins"`
}

type pluginListed struct { // plugins.jsonに指定される、リストア対象のplugin
	Name     string `json:"name"`
	Priority int    `json:"priority"`
	Enabled  bool   `json:"enabled"`
}

func ReadPluginList() error {
	raw, err := os.ReadFile(config.Config.PluginsList)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			file, err := os.Create(config.Config.PluginsList)
			if err != nil {
				log.Fatal("Failed to create plugins list: ", err)
			}
			defer file.Close()
			raw, err := json.MarshalIndent(PluginList, "", "\t")
			if err != nil {
				log.Fatal("Failed to marshal template plugins list: ", err)
			}
			_, err = file.Write(raw)
			if err != nil {
				log.Fatal("Failed to write template plugins list: ", err)
			}
		} else {
			log.Fatal("Failed to read plugins_list: ", err)
		}
	}
	PluginList = pluginList{}
	err = json.Unmarshal(raw, &PluginList)
	if err != nil {
		log.Fatal("Failed to unmarshal plugins_list: ", err)
	}
	return nil
}
