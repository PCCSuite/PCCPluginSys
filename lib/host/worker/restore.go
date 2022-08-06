package worker

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/PCCSuite/PCCPluginSys/lib/host/config"
)

type PluginSubjects struct {
	Plugins []PluginSubject `json:"plugins"`
}

type PluginSubject struct { // plugins.jsonに指定される、リストア対象のplugin
	Name     string `json:"name"`
	Priority int    `json:"priority"`
}

var pluginSubjects PluginSubjects

func readPluginSubjects() error {
	raw, err := ioutil.ReadFile(config.Config.PluginsList)
	if err != nil {
		log.Fatal("Failed to read plugins.json:", err)
	}
	pluginSubjects = PluginSubjects{}
	err = json.Unmarshal(raw, &pluginSubjects)
	if err != nil {
		log.Fatal("Failed to unmarshal plugins.json:", err)
	}
	return nil
}
