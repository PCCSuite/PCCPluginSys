package plugin

import (
	"log"
	"sync"
)

type ActionStatus string

const (
	ActionStatusLoaded     ActionStatus = "loaded"
	ActionStatusWaitStart  ActionStatus = "wait_start"
	ActionStatusRunning    ActionStatus = "running"
	ActionStatusWaitDepend ActionStatus = "wait_depend"
	ActionStatusWaitLock   ActionStatus = "wait_lock"
	ActionStatusDone       ActionStatus = "done"
	ActionStatusFailed     ActionStatus = "failed"
)

var Actions map[string]*ActionData = make(map[string]*ActionData)

type ActionData struct {
	Name       string
	Plugin     *Plugin
	Status     ActionStatus
	StatusText string
	Priority   int
	Notify     []chan<- ActionStatus
	Stop       func()
}

func NewActionData(name string, status ActionStatus, statusText string, priority int, stop func()) *ActionData {
	data := ActionData{
		Name:       name,
		Status:     status,
		StatusText: statusText,
		Priority:   priority,
		Stop:       stop,
	}
	Actions[name] = &data
	return &data
}

func (d *ActionData) SetActionStatusBoth(status ActionStatus, text string) {
	d.Status = status
	d.StatusText = text
	d.notifyStatus()
}

func (d *ActionData) SetActionStatusOnly(status ActionStatus) {
	d.Status = status
	d.notifyStatus()
}

func (d *ActionData) SetActionStatusText(text string) {
	d.StatusText = text
	d.notifyStatus()
}

func (d *ActionData) notifyStatus() {
	log.Printf("Plugin: %s, Status: %s, StatusText: %s", d.Name, d.Status, d.StatusText)
	for _, v := range d.Notify {
		v <- d.Status
	}
	for _, v := range globalNotify {
		v <- struct{}{}
	}
}

var subscribeMuteX sync.Mutex

func (d *ActionData) SubscribeStatus(ch chan<- ActionStatus) {
	subscribeMuteX.Lock()
	defer subscribeMuteX.Unlock()
	d.Notify = append(d.Notify, ch)
}

func (d *ActionData) UnsubscribeStatus(ch chan<- ActionStatus) {
	subscribeMuteX.Lock()
	defer subscribeMuteX.Unlock()
	new := make([]chan<- ActionStatus, len(d.Notify)-1)
	for _, v := range d.Notify {
		if v != ch {
			new = append(new, v)
		}
	}
	d.Notify = new
}

var globalNotify []chan<- struct{}

func SubscribeGlobalStatus(ch chan<- struct{}) {
	subscribeMuteX.Lock()
	defer subscribeMuteX.Unlock()
	globalNotify = append(globalNotify, ch)
}

func UnsubscribeGlobalStatus(ch chan<- struct{}) {
	subscribeMuteX.Lock()
	defer subscribeMuteX.Unlock()
	new := make([]chan<- struct{}, len(globalNotify)-1)
	for _, v := range globalNotify {
		if v != ch {
			new = append(new, v)
		}
	}
	globalNotify = new
}
