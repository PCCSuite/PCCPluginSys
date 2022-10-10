package cmd

import (
	"net"
	"sync"

	"github.com/PCCSuite/PCCPluginSys/lib/common"
)

var ExecuterUserConn *net.TCPConn
var ExecuterAdminConn *net.TCPConn

var Process map[int]chan<- *common.ExecuterResultData = make(map[int]chan<- *common.ExecuterResultData)

var lastNum int

var mutex sync.Mutex

func newRequest() (int, <-chan *common.ExecuterResultData) {
	mutex.Lock()
	defer mutex.Unlock()
	lastNum++
	ch := make(chan *common.ExecuterResultData)
	Process[lastNum] = ch
	return lastNum, ch
}

func unlisten(requestId int) {
	mutex.Lock()
	defer mutex.Unlock()
	ch := Process[requestId]
	Process[requestId] = nil
	close(ch)
}
