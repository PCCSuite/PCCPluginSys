package common

import "net"

type ClientType string

const (
	ExecuterUser  ClientType = "executer-user"
	ExecuterAdmin ClientType = "executer-admin"
	PCCClient     ClientType = "pccclient"
	API           ClientType = "api"
)

type DataType string

const (
	DataTypeNegotiate DataType = "negotiate"
)

type CommonData struct {
	Data_type DataType `json:"data_type"`
}

type Negotiate struct {
	Data_type   DataType   `json:"data_type"`
	Client_type ClientType `json:"client_type"`
}

func NewNegotiateData(client_type ClientType) Negotiate {
	return Negotiate{
		Data_type:   DataTypeNegotiate,
		Client_type: client_type,
	}
}

var Addr = &net.TCPAddr{
	IP:   net.IPv4(127, 0, 0, 1),
	Port: 15000,
}
