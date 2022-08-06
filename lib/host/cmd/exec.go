package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"os"
	"strings"

	"github.com/PCCSuite/PCCPluginSys/lib/data"
	"github.com/PCCSuite/PCCPluginSys/lib/host/plugin"
)

const EXEC = "EXEC"

var ExecuterUserConn *net.TCPConn
var ExecuterAdminConn *net.TCPConn

var Process map[int]chan<- *data.ExecuterResultData

var lastNum int

type ExecCmd struct {
	pluginStarter *plugin.Plugin
	pluginCaller  *plugin.Plugin
	param         []string
	ctx           context.Context
	cancel        context.CancelFunc
	child         Cmd
}

func NewExecCmd(plugin *plugin.Plugin, caller *plugin.Plugin, param []string, ctx context.Context) *ExecCmd {
	ctx, cancel := context.WithCancel(ctx)
	return &ExecCmd{
		pluginStarter: plugin,
		pluginCaller:  caller,
		param:         param,
		ctx:           ctx,
		cancel:        cancel,
	}
}

var ErrNonZeroCode = errors.New("exec return non zero code")

// ErrActionNotFound,ErrStoped throwable
func (c *ExecCmd) Run() error {
	if len(c.param) < 1 {
		return ErrTooFewArgs
	}
	param := c.param
	admin := false
	nofail := false
	dir := ""
paramcheck:
	for {
		switch c.param[0] {
		case "/ADMIN":
			param = param[1:]
			admin = true
		case "/DATADIR":
			dir = c.pluginCaller.GetDataDir()
			param = param[1:]
		case "/REPODIR":
			dir = c.pluginCaller.GetRepoDir()
			param = param[1:]
		case "/TEMPDIR":
			dir = c.pluginCaller.GetTempDir()
			param = param[1:]
		case "/NOFAIL":
			param = param[1:]
			nofail = true
		default:
			break paramcheck
		}
	}
	if admin {
		var network bool
		if strings.HasPrefix(param[0], "\\\\") {
			// is absolute network path
			network = true
		} else if param[0][1] == ':' {
			// is absolute drive path
			if param[0][0] == 'A' || param[0][0] == 'B' {
				// is pcc_homes or groups
				network = true
			} else {
				network = false
			}
		} else if strings.Contains(param[0], string(os.PathSeparator)) {
			// is relative path
		}
	}
	lastNum++
	reqId := lastNum
	env := []string{
		"PLUGIN_STARTER=" + c.pluginStarter.General.Name,
		"PLUGIN_NAME=" + c.pluginCaller.General.Name,
		"PLUGIN_REPODIR=" + c.pluginCaller.GetRepoDir(),
		"PLUGIN_DATADIR=" + c.pluginCaller.GetDataDir(),
		"PLUGIN_TEMPDIR=" + c.pluginCaller.GetTempDir(),
	}
	reqData := data.NewExecuterExec(param, dir, env, reqId)
	raw, err := json.Marshal(reqData)
	if err != nil {
		return err
	}
	if admin {
		_, err = ExecuterAdminConn.Write(raw)
	} else {
		_, err = ExecuterUserConn.Write(raw)
	}
	if err != nil {
		return err
	}
	ch := make(chan *data.ExecuterResultData)
	defer close(ch)
	Process[reqId] = ch
	defer func() {
		Process[reqId] = nil
	}()
	select {
	case result := <-ch:
		if !nofail && result.Code != 0 {
			return ErrNonZeroCode
		}
		return nil
	case <-c.ctx.Done():
		stopData := data.NewExecuterStop(reqId)
		raw, err := json.Marshal(stopData)
		if err != nil {
			return err
		}
		if admin {
			_, err = ExecuterAdminConn.Write(raw)
		} else {
			_, err = ExecuterUserConn.Write(raw)
		}
		if err != nil {
			return err
		}
		return ErrStopped
	}
}

func (c *ExecCmd) Stop() {
	c.cancel()
}
