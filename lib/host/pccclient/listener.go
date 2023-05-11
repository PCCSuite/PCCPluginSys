package pccclient

import (
	"context"
	"encoding/json"
	"log"
	"net"

	"github.com/PCCSuite/PCCPluginSys/lib/host/cmd"
	"github.com/PCCSuite/PCCPluginSys/lib/host/data"
	"github.com/PCCSuite/PCCPluginSys/lib/host/status"
	"github.com/PCCSuite/PCCPluginSys/lib/host/worker"
)

var Conn *net.TCPConn

func PCCCliListener(conn *net.TCPConn) {
	if Conn != nil {
		Conn.Close()
	}
	Conn = conn
	defer conn.Close()
	defer func() {
		Conn = nil
	}()
	go subscriber()
	go SendUpdate()
	decoder := json.NewDecoder(conn)
	for {
		cmdData := CommandData{}
		err := decoder.Decode(&cmdData)
		if err != nil {
			log.Print("Error decoding PCCClient message: ", err)
			break
		}
		go func() {
			switch cmdData.Data_type {
			case DataTypeRestore:
				worker.Restore()
			case DataTypeInstall:
				_, err := worker.InstallPackage(cmdData.Package, 0)
				if err != nil {
					log.Print("Error installing package requested from PCCClient: ", err)
				}
			case DataTypeAction:
				err = worker.RunAction(cmdData.Plugin, cmdData.Action, 0, context.Background())
				if err != nil {
					log.Print("Error running action requested from PCCClient: ", err)
				}
			case DataTypeCancel:
				running, ok := data.RunningActions[cmdData.Package]
				if !ok {
					log.Print("Cancelling action not found: ", cmdData.Package)
					return
				}
				running.Cancel()
			case DataTypeAnswer:
				askData, ok := cmd.Asking[cmdData.ID]
				if ok {
					askData.Ch <- cmdData.Value
				}
			default:
				log.Print("Unknown datatype from pccclient: ", cmdData.Data_type)
			}
		}()
	}
}

var pluginCh chan struct{}
var statusCh chan status.SysStatus

func subscriber() {
	if pluginCh != nil {
		return
	}
	pluginCh = make(chan struct{})
	defer close(pluginCh)
	data.SubscribeGlobalStatus(pluginCh)
	defer data.UnsubscribeGlobalStatus(pluginCh)
	statusCh = make(chan status.SysStatus)
	defer close(statusCh)
	status.Listener = statusCh
	defer func() {
		status.Listener = nil
	}()
	for {
		select {
		case <-pluginCh:
		case <-statusCh:
		}
		go SendUpdate()
	}
}
