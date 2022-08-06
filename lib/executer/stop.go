package executer

import (
	"log"

	"github.com/PCCSuite/PCCPluginSys/lib/data"
)

func Stop(stopdata data.ExecuterStopData) {
	cmd, ok := cmds[stopdata.StopId]
	if !ok {
		log.Println("Failed to find cmd: ", stopdata.StopId)
	}
	go cmd.stop()
}
