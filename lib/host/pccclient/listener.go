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
	for {
		buf := make([]byte, 8192)
		i, err := Conn.Read(buf)
		raw := buf[:i]
		if err != nil {
			log.Print("Error reading PCCClient message: ", err)
			break
		}
		cmdData := CommandData{}
		json.Unmarshal(raw, &cmdData)
		switch cmdData.Data_type {
		case DataTypeRestore:
			worker.Restore()
		case DataTypeInstall:
			cmdData := InstallCommandData{}
			json.Unmarshal(raw, &cmdData)
			worker.InstallPackage(cmdData.Package, 0)
		case DataTypeAction:
			cmdData := ActionCommandData{}
			json.Unmarshal(raw, &cmdData)
			worker.RunAction(cmdData.Plugin, cmdData.Action, 0, context.Background())
		case DataTypeCancel:
			cmdData := CancelCommandData{}
			json.Unmarshal(raw, &cmdData)
			running, ok := data.RunningActions[cmdData.Package]
			if ok {
				running.Cancel()
			}
		case DataTypeAnswer:
			cmdData := AnswerCommandData{}
			json.Unmarshal(raw, &cmdData)
			askData, ok := cmd.Asking[cmdData.ID]
			if ok {
				askData.Ch <- cmdData.Value
			}
		default:
			log.Print("Unknown datatype from pccclient: ", cmdData.Data_type)
		}
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
