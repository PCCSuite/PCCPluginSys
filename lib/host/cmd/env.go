package cmd

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/PCCSuite/PCCPluginSys/lib/common"
	"github.com/PCCSuite/PCCPluginSys/lib/host/data"
)

const ENV = "ENV"

type EnvCmd struct {
	Package *data.Package
	plugin  *data.Plugin
	param   []string
	ctx     context.Context
}

func NewEnvCmd(Package *data.Package, plugin *data.Plugin, param []string, ctx context.Context) *EnvCmd {
	return &EnvCmd{
		Package: Package,
		plugin:  plugin,
		param:   param,
		ctx:     ctx,
	}
}

// ErrActionNotFound,ErrStoped throwable
func (c *EnvCmd) Run() error {
	if len(c.param) < 1 {
		return ErrTooFewArgs
	}
	param := c.param

	modeParam := false

	target := common.ExecuterEnvTargetMachine
	mode := common.ExecuterEnvModeSet
paramcheck:
	for {
		switch param[0] {
		case "/MACHINE":
			target = common.ExecuterEnvTargetMachine
			param = param[1:]
		case "/USER":
			target = common.ExecuterEnvTargetUser
			param = param[1:]
		case "/SET":
			modeParam = true
			mode = common.ExecuterEnvModeSet
			param = param[1:]
		case "/ADD":
			modeParam = true
			mode = common.ExecuterEnvModeAdd
			param = param[1:]
		default:
			break paramcheck
		}
		if len(param) < 2 {
			return ErrTooFewArgs
		}
	}

	key := param[0]
	value := strings.Join(param[1:], " ")

	if !modeParam {
		if strings.EqualFold(key, "Path") {
			mode = common.ExecuterEnvModeAdd
		}
	}

	// check stopped
	select {
	case <-c.ctx.Done():
		return ErrStopped
	default:
	}

	reqId, ch := newRequest()
	defer unlisten(reqId)

	reqData := common.NewExecuterEnv(target, mode, key, value, reqId)
	raw, err := json.Marshal(reqData)
	if err != nil {
		return err
	}
	if target == common.ExecuterEnvTargetMachine {
		_, err = ExecuterAdminConn.Write(raw)
	} else {
		_, err = ExecuterUserConn.Write(raw)
	}
	if err != nil {
		return err
	}

	select {
	case result := <-ch:
		if result.Code != 0 {
			return ErrNonZeroCode
		}
		return nil
	case <-c.ctx.Done():
		return ErrStopped
	}
}
