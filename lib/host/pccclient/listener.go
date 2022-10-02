package pccclient

import (
	"context"
	"encoding/json"
	"log"
	"net"

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
		data := CommandData{}
		json.Unmarshal(raw, &data)
		switch data.Data_type {
		case DataTypeRestore:
			worker.Restore()
		case DataTypeInstall:
			data := InstallCommandData{}
			json.Unmarshal(raw, &data)
			worker.InstallPackage(data.Plugin, 0)
		case DataTypeAction:
			data := ActionCommandData{}
			json.Unmarshal(raw, &data)
			worker.RunAction(data.Plugin, data.Action, 0, context.Background())
		default:
			log.Print("Unknown datatype from pccclient: ", data.Data_type)
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
