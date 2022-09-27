package executer

import (
	"encoding/json"
	"log"

	"github.com/PCCSuite/PCCPluginSys/lib/data"
)

func listen() {
	for {
		buf := make([]byte, 8192)
		i, err := Conn.Read(buf)
		raw := buf[:i]
		if err != nil {
			log.Print("Error in reading message from conn: ", err)
			return
		}
		cmddata := data.ExecuterCommandData{}
		err = json.Unmarshal(raw, &cmddata)
		if err != nil {
			log.Print("Error in unmarshaling message from conn: ", err)
			continue
		}
		if cmddata.DataType != data.DataTypeExecuterCommand {
			log.Print("Unexpected data_type from conn: ", err)
			continue
		}
		switch cmddata.Command {
		case data.ExecuterCommandExec:
			execdata := data.ExecuterExecData{}
			err = json.Unmarshal(raw, &execdata)
			if err != nil {
				log.Print("Error in unmarshaling exec message from conn: ", err)
				continue
			}
			Exec(execdata)
		case data.ExecuterCommandStop:
			stopdata := data.ExecuterStopData{}
			err = json.Unmarshal(raw, &stopdata)
			if err != nil {
				log.Print("Error in unmarshaling stop message from conn: ", err)
				continue
			}
			Stop(stopdata)
		default:
			log.Print("Unknown command: ", cmddata.Command)
		}
	}
}
