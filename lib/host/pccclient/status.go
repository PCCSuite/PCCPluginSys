package pccclient

import (
	"encoding/json"
	"log"

	"github.com/PCCSuite/PCCPluginSys/lib/host/plugin"
	"github.com/PCCSuite/PCCPluginSys/lib/host/status"
)

func SendUpdate() {
	log.Println("Change detected, sending update...")
	var plugins = make([]PluginData, 0)
	for _, v := range plugin.Actions {
		if v.Plugin != nil {
			dependency := make([]string, 0)
			for _, d := range v.Plugin.Dependency.Dependent {
				dependency = append(dependency, d.Name)
			}
			plugins = append(plugins, NewPluginData(v.Name, v.Plugin.GetRepoDir(), v.Plugin.Installed, false, v.Status, v.StatusText, dependency))
		} else {
			plugins = append(plugins, NewPluginData(v.Name, "", false, false, v.Status, v.StatusText, []string{}))
		}
	}
	data := NewClientNotifyData(status.Status, plugins)
	raw, err := json.Marshal(data)
	if err != nil {
		log.Print("Failed to marshal client notify: ", err)
	}
	_, err = Conn.Write(raw)
	if err != nil {
		log.Print("Failed to send client notify: ", err)
	}
	log.Println("Sent: ", raw)
}
