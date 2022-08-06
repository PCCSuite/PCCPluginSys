package srv

import (
	"encoding/json"
	"log"
	"net"
	"time"

	"github.com/PCCSuite/PCCPluginSys/lib/data"
)

func StartServer() {
	listener, err := net.ListenTCP("tcp", data.Addr)
	if err != nil {
		log.Fatal(err)
	}
	go accept(listener)
}

func accept(listener *net.TCPListener) {
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Print("Error in accepting connection", err)
			continue
		}
		conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			log.Print("Error in reading message from new conn", err)
			conn.Close()
			continue
		}
		msg := data.Negotiate{}
		err = json.Unmarshal(buf[:n], &msg)
		if err != nil {
			log.Print("Error in unmarshaling message from new conn", err)
			conn.Close()
			continue
		}
		if msg.Data_type != data.DataTypeNegotiate {
			log.Print("Invalid data_type from new conn", err)
			conn.Close()
			continue
		}
		newConn(msg.Client_type, conn)
	}
}

var pccClientConn *net.TCPConn

func newConn(clientType data.ClientType, conn *net.TCPConn) {
	switch clientType {
	case data.ExecuterUser:
		go listenExecuter(conn, false)
	case data.ExecuterAdmin:
		go listenExecuter(conn, true)
	case data.PCCClient:
		if pccClientConn != nil {
			pccClientConn.Close()
		}
		pccClientConn = conn
	}
}
