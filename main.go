package main

import (
	"log"
	"os"

	"github.com/PCCSuite/PCCPluginSys/lib/api"
	"github.com/PCCSuite/PCCPluginSys/lib/executer"
	"github.com/PCCSuite/PCCPluginSys/lib/host"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Must specify arguments")
	}
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
