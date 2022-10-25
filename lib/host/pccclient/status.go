package pccclient

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/PCCSuite/PCCPluginSys/lib/host/cmd"
	"github.com/PCCSuite/PCCPluginSys/lib/host/data"
	"github.com/PCCSuite/PCCPluginSys/lib/host/lock"
	"github.com/PCCSuite/PCCPluginSys/lib/host/status"
)

var sendQueued bool

var senderMutex = sync.Mutex{}
var senderRunning bool

func SendUpdate() {
	log.Println("Change detected, queueing update...")
	senderMutex.Lock()
	defer senderMutex.Unlock()
	sendQueued = true
	if !senderRunning {
		senderRunning = true
		go updateSender()
	}
}

func updateSender() {
	senderMutex.Lock()
	for sendQueued {
		sendQueued = false
		senderMutex.Unlock()
		var plugins = make([]PluginData, 0)
		for _, v := range data.RunningActions {
			if v.Package != nil {
				dependency := make([]string, 0)
				if !v.Package.Installed {
					installing := data.InstallingPackages[v.Package]
					for _, d := range installing.Dependent {
						dependency = append(dependency, d.Status.Status.PackageIdentifier)
					}
				}
				plugins = append(plugins, NewPluginData(v.PackageIdentifier, v.Package.Repo.Name, v.Package.Installed, lock.IsLocking(lock.DefaultName, v), v.Status, v.StatusText, dependency))
			} else {
				plugins = append(plugins, NewPluginData(v.PackageIdentifier, "", false, false, v.Status, v.StatusText, []string{}))
			}
		}
		var asking = make([]*cmd.AskData, 0)
		cmd.AskMutex.RLock()
		for _, v := range cmd.Asking {
			asking = append(asking, v)
		}
		cmd.AskMutex.RUnlock()
		data := NewClientNotifyData(status.Status, plugins, asking)
		raw, err := json.Marshal(data)
		if err != nil {
			log.Print("Failed to marshal client notify: ", err)
		}
		_, err = Conn.Write(raw)
		if err != nil {
			log.Print("Failed to send client notify: ", err)
		}
		log.Println("Sent: ", string(raw))
		time.Sleep(500 * time.Millisecond)
		senderMutex.Lock()
	}
	senderRunning = false
	senderMutex.Unlock()
}
