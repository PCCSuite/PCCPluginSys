package executer

import (
	"log"

	"github.com/PCCSuite/PCCPluginSys/lib/common"
)

func Stop(stopdata common.ExecuterCommandData) {
	cmd, ok := cmds[stopdata.StopId]
	if !ok {
		log.Println("Failed to find cmd: ", stopdata.StopId)
	}
	go cmd.stop()
}
