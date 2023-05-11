package srv

import (
	"encoding/json"
	"log"
	"net"

	"github.com/PCCSuite/PCCPluginSys/lib/common"
	"github.com/PCCSuite/PCCPluginSys/lib/host/cmd"
)

var execUserConnected bool
var execAdminConnected bool
var execConnectChan chan<- struct{}

func WaitExecuter() {
	if execUserConnected && execAdminConnected {
		return
	}
	if execConnectChan != nil {
		log.Panic("many process waiting execConnect")
	}
	defer func() {
		execConnectChan = nil
	}()
	for {
		ch := make(chan struct{})
		execConnectChan = ch
		if execUserConnected && execAdminConnected {
			return
		}
		<-ch
	}
}

func listenExecuter(conn *net.TCPConn, admin bool) {
	if admin {
		if cmd.ExecuterAdminConn != nil {
			cmd.ExecuterAdminConn.Close()
		}
		cmd.ExecuterAdminConn = conn
		execAdminConnected = true
	} else {
		if cmd.ExecuterUserConn != nil {
			cmd.ExecuterUserConn.Close()
		}
		cmd.ExecuterUserConn = conn
		execUserConnected = true
	}
	if execConnectChan != nil {
		execConnectChan <- struct{}{}
	}

	decoder := json.NewDecoder(conn)

	for {
		data := common.ExecuterResultData{}
		err := decoder.Decode(&data)
		if err != nil {
			log.Print("Error in decoding message from executer: ", err)
			continue
		}
		cmd.ExecMutex.RLock()
		if cmd.ExecProcess[data.Request_id] != nil {
			cmd.ExecProcess[data.Request_id] <- &data
		}
		cmd.ExecMutex.RUnlock()
	}
}
