package executer

import (
	"encoding/json"
	"log"
	"net"

	"github.com/PCCSuite/PCCPluginSys/lib/common"
)

var Mode common.ClientType

var Conn *net.TCPConn

func ExecuterMain(isAdmin bool) {
	if isAdmin {
		Mode = common.ExecuterAdmin
	} else {
		Mode = common.ExecuterUser
	}
	connect()
	log.Print("Negotiate complete, listening...")
	listen()
}

func connect() {
	var err error
	Conn, err = net.DialTCP("tcp", nil, common.Addr)
	if err != nil {
		log.Fatal("Failed to connect host: ", err)
	}
	negotiate := common.NewNegotiateData(Mode)
	raw, err := json.Marshal(negotiate)
	if err != nil {
		log.Fatal("Failed to marshal negotiate data: ", err)
	}
	_, err = Conn.Write(raw)
	if err != nil {
		log.Fatal("Failed to send negotiate data: ", err)
	}
}
