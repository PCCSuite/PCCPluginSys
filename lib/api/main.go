package api

import (
	"encoding/json"
	"log"
	"net"
	"os"

	"github.com/PCCSuite/PCCPluginSys/lib/common"
)

var Conn *net.TCPConn

func ApiMain() {
	connect()
	data := common.NewApiRequestData(os.Getenv("plugin_starter"), os.Getenv("plugin_name"), os.Args[1:])
	raw, err := json.Marshal(data)
	if err != nil {
		log.Fatal("Failed to unmarshal data: ", err)
	}
	_, err = Conn.Write(raw)
	if err != nil {
		log.Fatal("Failed to send request: ", err)
	}
}

func connect() {
	var err error
	Conn, err = net.DialTCP("tcp", nil, common.Addr)
	if err != nil {
		log.Fatal("Failed to connect host: ", err)
	}
	negotiate := common.NewNegotiateData(common.API)
	raw, err := json.Marshal(negotiate)
	if err != nil {
		log.Fatal("Failed to marshal negotiate data: ", err)
	}
	_, err = Conn.Write(raw)
	if err != nil {
		log.Fatal("Failed to send negotiate data: ", err)
	}
}
