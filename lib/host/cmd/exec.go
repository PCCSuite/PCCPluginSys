package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/PCCSuite/PCCPluginSys/lib/data"
	"github.com/PCCSuite/PCCPluginSys/lib/host/plugin"
)

const EXEC = "EXEC"

var ExecuterUserConn *net.TCPConn
var ExecuterAdminConn *net.TCPConn

var Process map[int]chan<- *data.ExecuterResultData = make(map[int]chan<- *data.ExecuterResultData)

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
		if len(c.param) < 1 {
			return ErrTooFewArgs
		}
	}
	var err error
	var abs string // absolute path if param[0] is path
	var network bool
	if strings.HasPrefix(param[0], "\\\\") {
		// absolute network path
		abs, err = filepath.Abs(param[0])
		network = true
	} else if len(param[0]) > 1 && param[0][1] == ':' {
		// absolute drive path
		abs, err = filepath.Abs(param[0])
	} else if strings.Contains(param[0], string(os.PathSeparator)) {
		// relative path
		if dir == "" {
			// directory not specified
			dir = c.pluginCaller.GetRepoDir()
		}
		abs, err = filepath.Abs(filepath.Join(dir, param[0]))
	}
	if err != nil {
		return err
	}

	// if admin && network, copy exec file to tempdir
	if admin {
		if !network && abs != "" && (abs[0] == 'A' || abs[0] == 'B') {
			// is pcc_homes or groups
			network = true
		}
		if network {
			c.child = NewExecCmd(c.pluginStarter, c.pluginCaller, []string{"robocopy", abs, c.pluginCaller.GetTempDir()}, c.ctx)
			err = c.child.Run()
			if err != nil {
				return err
			}
			param[0] = filepath.Join(c.pluginCaller.GetTempDir(), filepath.Base(abs))
		}
	}

	// check stopped
	select {
	case <-c.ctx.Done():
		return ErrStopped
	default:
	}

	// Execution
	lastNum++
	reqId := lastNum
	env := []string{
		"PLUGIN_STARTER=" + c.pluginStarter.General.Name,
		"PLUGIN_NAME=" + c.pluginCaller.General.Name,
		"PLUGIN_REPODIR=" + c.pluginCaller.GetRepoDir(),
		"PLUGIN_DATADIR=" + c.pluginCaller.GetDataDir(),
		"PLUGIN_TEMPDIR=" + c.pluginCaller.GetTempDir(),
	}
	logFile := filepath.Join(filepath.Join(c.pluginCaller.GetTempDir(), "executer.log"))
	reqData := data.NewExecuterExec(param, dir, logFile, env, reqId)
	raw, err := json.Marshal(reqData)
	if err != nil {
		return err
	}
	ch := make(chan *data.ExecuterResultData)
	defer close(ch)
	Process[reqId] = ch
	defer func() {
		Process[reqId] = nil
	}()
	if admin {
		_, err = ExecuterAdminConn.Write(raw)
	} else {
		_, err = ExecuterUserConn.Write(raw)
	}
	if err != nil {
		return err
	}
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
	if c.child != nil {
		c.child.Stop()
	}
}
