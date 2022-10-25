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
		err = json.Unmarshal(raw, &cmdData)
		if err != nil {
			log.Print("Error unmarshaling PCCClient message: ", err)
			break
		}
		switch cmdData.Data_type {
		case DataTypeRestore:
			worker.Restore()
		case DataTypeInstall:
			cmdData := InstallCommandData{}
			err = json.Unmarshal(raw, &cmdData)
			if err != nil {
				log.Print("Error unmarshaling PCCClient install message: ", err)
				continue
			}
			worker.InstallPackage(cmdData.Package, 0)
		case DataTypeAction:
			cmdData := ActionCommandData{}
			err = json.Unmarshal(raw, &cmdData)
			if err != nil {
				log.Print("Error unmarshaling PCCClient action message: ", err)
				continue
			}
			worker.RunAction(cmdData.Plugin, cmdData.Action, 0, context.Background())
		case DataTypeCancel:
			cmdData := CancelCommandData{}
			err = json.Unmarshal(raw, &cmdData)
			if err != nil {
				log.Print("Error unmarshaling PCCClient cancel message: ", err)
				continue
			}
			running, ok := data.RunningActions[cmdData.Package]
			if !ok {
				log.Print("Cancelling action not found: ", cmdData.Package)
				continue
			}
			running.Cancel()
		case DataTypeAnswer:
			cmdData := AnswerCommandData{}
			err = json.Unmarshal(raw, &cmdData)
			if err != nil {
				log.Print("Error unmarshaling PCCClient cancel message: ", err)
				continue
			}
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
