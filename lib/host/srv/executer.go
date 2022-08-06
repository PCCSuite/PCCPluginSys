package srv

import (
	"encoding/json"
	"log"
	"net"

	"github.com/PCCSuite/PCCPluginSys/lib/data"
	"github.com/PCCSuite/PCCPluginSys/lib/host/cmd"
)

func listenExecuter(conn *net.TCPConn, admin bool) {
	if admin {
		if cmd.ExecuterAdminConn != nil {
			cmd.ExecuterAdminConn.Close()
		}
		cmd.ExecuterAdminConn = conn
	} else {
		if cmd.ExecuterUserConn != nil {
			cmd.ExecuterUserConn.Close()
		}
		cmd.ExecuterUserConn = conn
	}
	for {
		buf := make([]byte, 8192)
		n, err := conn.Read(buf)
		if err != nil {
			log.Print("Error in reading message from executer", err)
			continue
		}
		data := data.ExecuterResultData{}
		err = json.Unmarshal(buf[:n], &data)
		if err != nil {
			log.Print("Error in unmarshaling message from executer", err)
			continue
		}
		if cmd.Process[data.Request_id] != nil {
			cmd.Process[data.Request_id] <- &data
		}
	}
}
