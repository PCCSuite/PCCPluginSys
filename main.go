package main

import (
	"os"

	"github.com/PCCSuite/PCCPluginSys/lib/api"
	"github.com/PCCSuite/PCCPluginSys/lib/executer"
	"github.com/PCCSuite/PCCPluginSys/lib/host"
)

func main() {
	mode := os.Args[1]
	switch mode {
	case "host":
		host.HostMain()
	case "executer-user":
		executer.ExecuterMain(false)
	case "executer-admin":
		executer.ExecuterMain(true)
	default:
		api.ApiMain()
	}
}
