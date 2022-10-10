package srv

import (
	"encoding/json"
	"log"
	"net"
	"time"

	"github.com/PCCSuite/PCCPluginSys/lib/common"
	"github.com/PCCSuite/PCCPluginSys/lib/host/pccclient"
)

func StartServer() {
	listener, err := net.ListenTCP("tcp", common.Addr)
	if err != nil {
		log.Fatal(err)
	}
	go accept(listener)
}

func accept(listener *net.TCPListener) {
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Print("Error in accepting connection: ", err)
			continue
		}
		conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		buf := make([]byte, 8192)
		i, err := conn.Read(buf)
		raw := buf[:i]
		if err != nil {
			log.Print("Error in reading message from new conn: ", err)
			conn.Close()
			continue
		}
		msg := common.Negotiate{}
		err = json.Unmarshal(raw, &msg)
		if err != nil {
			log.Print("Error in unmarshaling message from new conn: ", err)
			conn.Close()
			continue
		}
		if msg.Data_type != common.DataTypeNegotiate {
			log.Print("Invalid data_type from new conn: ", err)
			conn.Close()
			continue
		}
		conn.SetReadDeadline(time.Now().Add(1160000 * time.Hour))
		newConn(msg.Client_type, conn)
	}
}

func newConn(clientType common.ClientType, conn *net.TCPConn) {
	log.Print("Connected client: ", clientType)
	switch clientType {
	case common.ExecuterUser:
		go listenExecuter(conn, false)
	case common.ExecuterAdmin:
		go listenExecuter(conn, true)
	case common.PCCClient:
		go pccclient.PCCCliListener(conn)
	case common.API:
		go listenApi(conn)
	}
}
