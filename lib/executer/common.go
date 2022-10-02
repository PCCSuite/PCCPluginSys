package executer

import (
	"encoding/json"
	"log"

	"github.com/PCCSuite/PCCPluginSys/lib/common"
)

type Cmd interface {
	run()
	stop()
}

var cmds map[int]Cmd = make(map[int]Cmd)

func send(result common.ExecuterResultData) {
	raw, err := json.Marshal(result)
	if err != nil {
		log.Println("failed to marshal result: ", err)
		return
	}
	_, err = Conn.Write(raw)
	if err != nil {
		log.Println("failed to send result: ", err)
		return
	}
}
