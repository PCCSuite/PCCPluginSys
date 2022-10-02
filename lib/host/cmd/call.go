package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/PCCSuite/PCCPluginSys/lib/host/data"
)

const CALL = "CALL"

type CallCmd struct {
	Package *data.Package
	param   []string
	ctx     context.Context
	running Cmd
}

func NewCallCmd(Package *data.Package, param []string, ctx context.Context) *CallCmd {
	return &CallCmd{
		Package: Package,
		param:   param,
		ctx:     ctx,
	}
}

var ErrCommandNotFound = errors.New("command not found")

// ErrActionNotFound,ErrStoped throwable
func (c *CallCmd) Run() error {
	log.Printf("CALL: plugin: %s, param: %s", c.Package.Name, strings.Join(c.param, " , "))
	if len(c.param) < 1 {
		return ErrTooFewArgs
	}
	splitParam := strings.SplitN(c.param[0], ":", 2)
	var plugin *data.Plugin
	var actionName string
	if len(splitParam) == 1 {
		if c.Package.Type != data.PackageTypeInternal {
			log.Panic("no package CALL by not Internal Package")
		}
		plugin = c.Package.Plugin
		actionName = splitParam[0]
	} else {
		plugin = data.GetPlugin(splitParam[0])
		if plugin == nil {
			return data.ErrPluginNotFound
		}
		actionName = splitParam[1]
	}
	rawAction := plugin.GetAction(actionName)
	splitAction := strings.Split(rawAction, "\n")
	log.Printf("CALL: splitAction: %s", strings.Join(splitAction, " , "))
	for i, v := range splitAction {

		// check stop
		select {
		case <-c.ctx.Done():
			return ErrStopped
		default:
			// continue process
		}

		trimed := strings.TrimSpace(v)
		if trimed == "" {
			continue
		}
		split := strings.SplitN(trimed, " ", 2)
		var param []string
		if len(split) == 2 {
			var err error
			param, err = parseParam(split[1])
			if err != nil {
				return err
			}
		}
		param = replaceParams(param, c.Package, plugin, c.param[1:])
		switch strings.ToUpper(split[0]) {
		case CALL:
			c.running = NewCallCmd(c.Package, param, c.ctx)
		case EXEC:
			c.running = NewExecCmd(c.Package, plugin, param, c.ctx)
		case LOCK:
			c.running = NewLockCmd(c.Package, param, c.ctx)
		default:
			return fmt.Errorf("action %s:%s:%d: %w", plugin.Name, actionName, i, ErrCommandNotFound)
		}
		err := c.running.Run()
		if err != nil {
			return fmt.Errorf("action %s:%s:%d: %w", plugin.General.Name, actionName, i, err)
		}
	}
	return nil
}
