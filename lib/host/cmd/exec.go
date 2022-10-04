package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/PCCSuite/PCCPluginSys/lib/common"
	"github.com/PCCSuite/PCCPluginSys/lib/host/data"
)

const EXEC = "EXEC"

var ExecuterUserConn *net.TCPConn
var ExecuterAdminConn *net.TCPConn

var Process map[int]chan<- *common.ExecuterResultData = make(map[int]chan<- *common.ExecuterResultData)

var lastNum int

type ExecCmd struct {
	Package *data.Package
	plugin  *data.Plugin
	param   []string
	ctx     context.Context
	cancel  context.CancelFunc
	child   Cmd
}

func NewExecCmd(Package *data.Package, plugin *data.Plugin, param []string, ctx context.Context) *ExecCmd {
	ctx, cancel := context.WithCancel(ctx)
	return &ExecCmd{
		Package: Package,
		plugin:  plugin,
		param:   param,
		ctx:     ctx,
		cancel:  cancel,
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
	noauto := false
	dir := ""
paramcheck:
	for {
		switch param[0] {
		case "/ADMIN":
			param = param[1:]
			admin = true
		case "/DATADIR":
			dir = c.plugin.GetDataDir()
			param = param[1:]
		case "/REPODIR":
			dir = c.plugin.GetRepoDir()
			param = param[1:]
		case "/TEMPDIR":
			dir = c.plugin.GetTempDir()
			param = param[1:]
		case "/NOFAIL":
			param = param[1:]
			nofail = true
		case "/RAW":
			param = param[1:]
			noauto = true
		default:
			break paramcheck
		}
		if len(param) < 1 {
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
			dir = c.plugin.GetRepoDir()
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
			c.child = NewExecCmd(c.Package, c.plugin, []string{"robocopy", abs, c.plugin.GetTempDir()}, c.ctx)
			err = c.child.Run()
			if err != nil {
				return err
			}
			param[0] = filepath.Join(c.plugin.GetTempDir(), filepath.Base(abs))
		}
	}

	if !noauto {
		if strings.HasSuffix(param[0], ".ps1") {
			param = append([]string{"powershell.exe", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Unrestricted", "-File"}, param...)
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
		"PLUGIN_STARTER=" + c.Package.Name,
		"PLUGIN_NAME=" + c.plugin.General.Name,
		"PLUGIN_REPODIR=" + c.plugin.GetRepoDir(),
		"PLUGIN_DATADIR=" + c.plugin.GetDataDir(),
		"PLUGIN_TEMPDIR=" + c.plugin.GetTempDir(),
	}
	logFile := filepath.Join(filepath.Join(c.plugin.GetTempDir(), "executer.log"))
	reqData := common.NewExecuterExec(param, dir, logFile, env, reqId)
	raw, err := json.Marshal(reqData)
	if err != nil {
		return err
	}
	ch := make(chan *common.ExecuterResultData)
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
		stopData := common.NewExecuterStop(reqId)
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
