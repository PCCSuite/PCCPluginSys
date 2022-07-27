package srv

import (
	"encoding/json"
	"log"
	"net"
	"time"

	"github.com/PCCSuite/PCCPluginSys/lib/data"
)

func StartServer() {
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: 15000,
	})
	if err != nil {
		log.Fatal(err)
	}
	go listen(listener)
}

func listen(listener *net.TCPListener) {
	conn, err := listener.AcceptTCP()
	if err != nil {
		log.Print("Error in accepting connection", err)
	}
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		log.Print("Error in reading message from new conn", err)
		conn.Close()
	}
	msg := data.Negotiate{}
	err = json.Unmarshal(buf[:n], &msg)
	if err != nil {
		log.Print("Error in unmarshaling message from new conn", err)
		conn.Close()
	}
}
