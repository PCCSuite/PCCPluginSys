package cmd

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/PCCSuite/PCCPluginSys/lib/host/data"
)

const ASK = "ASK"

type AskCmd struct {
	Package *data.Package
	plugin  *data.Plugin
	param   []string
	ctx     context.Context
}

func NewAskCmd(Package *data.Package, plugin *data.Plugin, param []string, ctx context.Context) *AskCmd {
	return &AskCmd{
		Package: Package,
		plugin:  plugin,
		param:   param,
		ctx:     ctx,
	}
}

// ErrActionNotFound,ErrStoped throwable
func (c *AskCmd) Run() error {
	if len(c.param) < 1 {
		return ErrTooFewArgs
	}
	select {
	case <-c.ctx.Done():
		return ErrStopped
	default:
	}
	req, ch := newAskRequest(c.Package.RunningAction.PackageIdentifier, c.plugin.Name, c.param[0], strings.Join(c.param[1:], " "))
	defer askUnlisten(req.ID)
	c.Package.RunningAction.SetActionStatusOnly(data.ActionStatusWaitAsk)
	select {
	case res := <-ch:
		c.Package.RunningAction.SetActionStatusOnly(data.ActionStatusRunning)
		file, err := os.Create(filepath.Join(c.plugin.GetTempDir(), "plugin_ask.txt"))
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = file.WriteString(res)
		return err
	case <-c.ctx.Done():
		return ErrStopped
	}
}

type AskData struct {
	ID      int           `json:"id"`
	Package string        `json:"package"`
	Plugin  string        `json:"plugin"`
	Type    string        `json:"type"`
	Message string        `json:"message"`
	Ch      chan<- string `json:"-"`
}

var Asking map[int]*AskData = make(map[int]*AskData)

var lastAskNum int

var AskMutex sync.RWMutex

func newAskRequest(Package string, plugin string, Type string, msg string) (AskData, <-chan string) {
	ch := make(chan string)
	AskMutex.Lock()
	defer AskMutex.Unlock()
	lastAskNum++
	data := AskData{
		ID:      lastAskNum,
		Package: Package,
		Plugin:  plugin,
		Type:    Type,
		Message: msg,
		Ch:      ch,
	}
	Asking[lastAskNum] = &data
	return data, ch
}

func askUnlisten(requestId int) {
	AskMutex.Lock()
	defer AskMutex.Unlock()
	ch := Asking[requestId].Ch
	delete(Asking, requestId)
	close(ch)
}
