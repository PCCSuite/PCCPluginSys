package cmd

import (
	"net"
	"sync"

	"github.com/PCCSuite/PCCPluginSys/lib/common"
)

var ExecuterUserConn *net.TCPConn
var ExecuterAdminConn *net.TCPConn

var ExecProcess map[int]chan<- *common.ExecuterResultData = make(map[int]chan<- *common.ExecuterResultData)

var lastExecNum int

var ExecMutex sync.RWMutex

func newExecRequest() (int, <-chan *common.ExecuterResultData) {
	ExecMutex.Lock()
	defer ExecMutex.Unlock()
	lastExecNum++
	ch := make(chan *common.ExecuterResultData)
	ExecProcess[lastExecNum] = ch
	return lastExecNum, ch
}

func execUnlisten(requestId int) {
	ExecMutex.Lock()
	defer ExecMutex.Unlock()
	ch := ExecProcess[requestId]
	ExecProcess[requestId] = nil
	close(ch)
}
