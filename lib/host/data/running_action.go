package data

import (
	"context"
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

var RunningActions map[string]*RunningAction = make(map[string]*RunningAction)

type RunningAction struct {
	PackageIdentifier string
	// This can be nil if not found
	Package    *Package
	Status     ActionStatus
	StatusText string
	Priority   int
	Notify     []chan<- ActionStatus
	Ctx        context.Context
	Cancel     context.CancelFunc
}

func NewRunningAction(packageIdentifier string, status ActionStatus, statusText string, priority int, ctx context.Context, cancel context.CancelFunc) *RunningAction {
	data := RunningAction{
		PackageIdentifier: packageIdentifier,
		Status:            status,
		StatusText:        statusText,
		Priority:          priority,
		Notify:            make([]chan<- ActionStatus, 0),
		Ctx:               ctx,
		Cancel:            cancel,
	}
	RunningActions[packageIdentifier] = &data
	return &data
}

func (r *RunningAction) IsEnded() bool {
	if r.Status == ActionStatusDone {
		return true
	}
	if r.Status == ActionStatusFailed {
		return true
	}
	return false
}

func (r *RunningAction) SetActionStatusBoth(status ActionStatus, text string) {
	r.Status = status
	r.StatusText = text
	r.notifyStatus()
}

func (r *RunningAction) SetActionStatusOnly(status ActionStatus) {
	r.Status = status
	r.notifyStatus()
}

func (r *RunningAction) SetActionStatusText(text string) {
	r.StatusText = text
	r.notifyStatus()
}

func (r *RunningAction) notifyStatus() {
	for _, v := range r.Notify {
		v <- r.Status
	}
	for _, v := range globalNotify {
		v <- struct{}{}
	}
}

var subscribeMuteX sync.Mutex

func (r *RunningAction) SubscribeStatus(ch chan<- ActionStatus) {
	subscribeMuteX.Lock()
	defer subscribeMuteX.Unlock()
	r.Notify = append(r.Notify, ch)
}

func (r *RunningAction) UnsubscribeStatus(ch chan<- ActionStatus) {
	subscribeMuteX.Lock()
	defer subscribeMuteX.Unlock()
	new := make([]chan<- ActionStatus, len(r.Notify)-1)
	for _, v := range r.Notify {
		if v != ch {
			new = append(new, v)
		}
	}
	r.Notify = new
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
