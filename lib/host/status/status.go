package status

import "log"

type SysStatus int

const (
	SysStatusReadying SysStatus = 1
	SysStatusReady    SysStatus = 2
	SysStatusRunning  SysStatus = 3
)

var Status SysStatus = SysStatusReadying

var Listener chan<- SysStatus

func SetStatus(newStatus SysStatus) {
	Status = newStatus
	log.Print("System status changed: ", Status)
	Listener <- Status
}
