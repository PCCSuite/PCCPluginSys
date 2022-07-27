package data

type ClientType int

const (
	ExecuterUser ClientType = iota
	ExecuterAdmin
	API
)

type Negotiate struct {
	Data_type   string
	Client_type ClientType
}

func NewNegotiateData(client_type ClientType) Negotiate {
	return Negotiate{
		Data_type:   "negotiate",
		Client_type: client_type,
	}
}
