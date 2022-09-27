package host

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/PCCSuite/PCCPluginSys/lib/host/config"
	"github.com/PCCSuite/PCCPluginSys/lib/host/srv"
	"github.com/PCCSuite/PCCPluginSys/lib/host/status"
	"github.com/PCCSuite/PCCPluginSys/lib/host/subp"
)

func HostMain() {
	config.ReadConfig()
	srv.StartServer()
	subp.StartExecuters()
	srv.WaitExecuter()

	status.SetStatus(status.SysStatusReady)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	s := <-sig

	fmt.Println("stopping... signal: ", s)
}
