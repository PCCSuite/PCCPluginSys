package executer

import (
	"encoding/json"
	"log"

	"github.com/PCCSuite/PCCPluginSys/lib/data"
)

func listen() {
	for {
		buf := make([]byte, 8192)
		n, err := Conn.Read(buf)
		if err != nil {
			log.Print("Error in reading message from conn", err)
			return
		}
		cmddata := data.ExecuterCommandData{}
		err = json.Unmarshal(buf[:n], &cmddata)
		if err != nil {
			log.Print("Error in unmarshaling message from conn", err)
			continue
		}
		if cmddata.DataType != data.DataTypeExecuterCommand {
			log.Print("Unexpected data_type from conn", err)
			continue
		}
		switch cmddata.Command {
		case data.ExecuterCommandExec:
			execdata := data.ExecuterExecData{}
			err = json.Unmarshal(buf[:n], &execdata)
			if err != nil {
				log.Print("Error in unmarshaling exec message from conn", err)
				continue
			}
			Exec(execdata)
		case data.ExecuterCommandStop:
			stopdata := data.ExecuterStopData{}
			err = json.Unmarshal(buf[:n], &stopdata)
			if err != nil {
				log.Print("Error in unmarshaling stop message from conn", err)
				continue
			}
			Stop(stopdata)
		default:
			log.Print("Unknown command", err)
		}
	}
}
