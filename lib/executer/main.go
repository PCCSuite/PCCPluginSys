package executer

import (
	"encoding/json"
	"log"
	"net"

	"github.com/PCCSuite/PCCPluginSys/lib/data"
)

var Mode data.ClientType

var Conn *net.TCPConn

func ExecuterMain(isAdmin bool) {
	if isAdmin {
		Mode = data.ExecuterAdmin
	} else {
		Mode = data.ExecuterUser
	}
	connect()
	log.Print("Negotiate complete, listening...")
	listen()
}

func connect() {
	var err error
	Conn, err = net.DialTCP("tcp", nil, data.Addr)
	if err != nil {
		log.Fatal("Failed to connect host: ", err)
	}
	negotiate := data.NewNegotiateData(Mode)
	raw, err := json.Marshal(negotiate)
	if err != nil {
		log.Fatal("Failed to marshal negotiate data: ", err)
	}
	_, err = Conn.Write(raw)
	if err != nil {
		log.Fatal("Failed to send negotiate data: ", err)
	}
}
