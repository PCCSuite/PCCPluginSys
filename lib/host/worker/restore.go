package worker

import (
	"encoding/json"
	"log"
	"os"

	"github.com/PCCSuite/PCCPluginSys/lib/host/config"
	"github.com/PCCSuite/PCCPluginSys/lib/host/status"
)

type PluginSubjects struct {
	Plugins []PluginSubject `json:"plugins"`
}

type PluginSubject struct { // plugins.jsonに指定される、リストア対象のplugin
	Name     string `json:"name"`
	Priority int    `json:"priority"`
	Enabled  bool   `json:"enabled"`
}

var pluginSubjects PluginSubjects

func Restore() {
	status.SetStatus(status.SysStatusRunning)
	readPluginSubjects()
	for _, v := range pluginSubjects.Plugins {
		if v.Enabled {
			_, err := InstallPlugin(v.Name, v.Priority)
			if err != nil {
				log.Printf("Failed to install %s: %v", v.Name, err)
			}
		}
	}
}

func readPluginSubjects() error {
	raw, err := os.ReadFile(config.Config.PluginsList)
	if err != nil {
		log.Fatal("Failed to read plugins_list: ", err)
	}
	pluginSubjects = PluginSubjects{}
	err = json.Unmarshal(raw, &pluginSubjects)
	if err != nil {
		log.Fatal("Failed to unmarshal plugins_list: ", err)
	}
	return nil
}
