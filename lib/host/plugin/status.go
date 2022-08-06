package plugin

import "sync"

type ActionStatus string

const (
	ActionStatusWaitStart  ActionStatus = "wait_start"
	ActionStatusWaitLock   ActionStatus = "wait_lock"
	ActionStatusWaitDepend ActionStatus = "wait_depend"
	ActionStatusRunning    ActionStatus = "running"
	ActionStatusFailed     ActionStatus = "failed"
	ActionStatusDone       ActionStatus = "done"
)

type ActionStatusSet struct {
	Status *ActionStatus
	Text   *string
}

func (p *Plugin) SetActionStatusSet(status ActionStatusSet) {
	if status.Status != nil {
		p.ActionStatus.Status = status.Status
	}
	if status.Text != nil {
		p.ActionStatus.Text = status.Text
	}
	for _, v := range p.StatusNotify {
		v <- status
	}
}

func (p *Plugin) SetActionStatusBoth(status ActionStatus, text string) {
	p.SetActionStatusSet(ActionStatusSet{
		Status: &status,
		Text:   &text,
	})
}

func (p *Plugin) SetActionStatusOnly(status ActionStatus) {
	p.SetActionStatusSet(ActionStatusSet{
		Status: &status,
	})
}

func (p *Plugin) SetActionStatusText(text string) {
	p.SetActionStatusSet(ActionStatusSet{
		Text: &text,
	})
}

var subscribeMuteX sync.Mutex

func (p *Plugin) SubscribeStatus(ch chan<- ActionStatusSet) {
	subscribeMuteX.Lock()
	defer subscribeMuteX.Unlock()
	p.StatusNotify = append(p.StatusNotify, ch)
}

func (p *Plugin) UnsubscribeStatus(ch chan<- ActionStatusSet) {
	subscribeMuteX.Lock()
	defer subscribeMuteX.Unlock()
	new := make([]chan<- ActionStatusSet, len(p.StatusNotify)-1)
	for _, v := range p.StatusNotify {
		if v != ch {
			new = append(new, v)
		}
	}
	p.StatusNotify = new
}
