package executer

import (
	"encoding/json"
	"log"

	"github.com/PCCSuite/PCCPluginSys/lib/common"
)

func listen() {

	decoder := json.NewDecoder(Conn)

	for {
		cmddata := common.ExecuterCommandData{}
		err := decoder.Decode(&cmddata)
		if err != nil {
			log.Print("Error in unmarshaling message from conn: ", err)
			continue
		}
		if cmddata.DataType != common.DataTypeExecuterCommand {
			log.Print("Unexpected data_type from conn: ", err)
			continue
		}
		switch cmddata.Command {
		case common.ExecuterCommandExec:
			Exec(cmddata)
		case common.ExecuterCommandEnv:
			go Env(cmddata)
		case common.ExecuterCommandStop:
			Stop(cmddata)
		default:
			log.Print("Unknown command: ", cmddata.Command)
		}
	}
}
