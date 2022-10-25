package worker

import (
	"log"

	"github.com/PCCSuite/PCCPluginSys/lib/host/data"
	"github.com/PCCSuite/PCCPluginSys/lib/host/status"
)

func Restore() {
	status.SetStatus(status.SysStatusRunning)
	data.ReadPluginList()
	for _, v := range data.PluginList.Plugins {
		if v.Enabled {
			_, err := InstallPackage(v.Identifier, v.Priority)
			if err != nil {
				log.Printf("Failed to install %s: %v", v.Identifier, err)
			}
		}
	}
}
