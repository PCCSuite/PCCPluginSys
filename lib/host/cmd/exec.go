package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/PCCSuite/PCCPluginSys/lib/common"
	"github.com/PCCSuite/PCCPluginSys/lib/host/data"
)

const EXEC = "EXEC"

type ExecCmd struct {
	Package *data.Package
	plugin  *data.Plugin
	param   []string
	ctx     context.Context
}

func NewExecCmd(Package *data.Package, plugin *data.Plugin, param []string, ctx context.Context) *ExecCmd {
	return &ExecCmd{
		Package: Package,
		plugin:  plugin,
		param:   param,
		ctx:     ctx,
	}
}

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
			err = exec(c.plugin, c.Package, false, false, []string{"copy", abs, c.plugin.GetTempDir()}, dir, c.ctx)
			if err != nil {
				return err
			}
			param[0] = filepath.Join(c.plugin.GetTempDir(), filepath.Base(abs))
		}
	}

	if !noauto {
		if strings.HasSuffix(param[0], ".ps1") {
			param = append([]string{"powershell.exe", "-NonInteractive", "-ExecutionPolicy", "Unrestricted", "-File"}, param...)
		}
	}

	// check stopped
	select {
	case <-c.ctx.Done():
		return ErrStopped
	default:
	}

	// Execution
	return exec(c.plugin, c.Package, admin, nofail, param, dir, c.ctx)
}

var ErrNonZeroCode = errors.New("exec return non zero code")

func exec(plugin *data.Plugin, Package *data.Package, admin bool, nofail bool, param []string, dir string, ctx context.Context) error {
	reqId, ch := newRequest()
	defer unlisten(reqId)
	env := []string{
		"PLUGIN_STARTER=" + Package.Name,
		"PLUGIN_NAME=" + plugin.General.Name,
		"PLUGIN_REPODIR=" + plugin.GetRepoDir(),
		"PLUGIN_DATADIR=" + plugin.GetDataDir(),
		"PLUGIN_TEMPDIR=" + plugin.GetTempDir(),
	}
	logFile := filepath.Join(plugin.GetTempDir(), "executer.log")
	reqData := common.NewExecuterExec(param, dir, logFile, env, reqId)
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
	select {
	case result := <-ch:
		if !nofail && result.Code != 0 {
			return ErrNonZeroCode
		}
		return nil
	case <-ctx.Done():
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
