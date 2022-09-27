package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/PCCSuite/PCCPluginSys/lib/host/plugin"
)

const CALL = "CALL"

type CallCmd struct {
	plugin  *plugin.Plugin
	param   []string
	ctx     context.Context
	cancel  context.CancelFunc
	running Cmd
}

func NewCallCmd(plugin *plugin.Plugin, param []string, ctx context.Context) *CallCmd {
	ctx, cancel := context.WithCancel(ctx)
	return &CallCmd{
		plugin: plugin,
		param:  param,
		ctx:    ctx,
		cancel: cancel,
	}
}

var ErrCommandNotFound = errors.New("command not found")

// ErrActionNotFound,ErrStoped throwable
func (c *CallCmd) Run() error {
	log.Printf("CALL: plugin: %s, param: %s", c.plugin.General.Name, strings.Join(c.param, " , "))
	if len(c.param) < 1 {
		return ErrTooFewArgs
	}
	splitParam := strings.SplitN(c.param[0], ":", 2)
	var actionPlugin *plugin.Plugin
	var actionName string
	if len(splitParam) == 1 {
		actionPlugin = c.plugin
		actionName = splitParam[0]
	} else {
		var err error
		actionPlugin, err = plugin.SearchPlugin(splitParam[0])
		if err != nil {
			return err
		}
		actionName = splitParam[1]
	}
	rawAction := actionPlugin.GetAction(actionName)
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
		param = replaceParams(param, c.plugin, actionPlugin, c.param[1:])
		switch strings.ToUpper(split[0]) {
		case CALL:
			c.running = NewCallCmd(c.plugin, param, c.ctx)
		case EXEC:
			c.running = NewExecCmd(c.plugin, actionPlugin, param, c.ctx)
		default:
			return fmt.Errorf("action %s:%s:%d: %w", actionPlugin.General.Name, actionName, i, ErrCommandNotFound)
		}
		err := c.running.Run()
		if err != nil {
			return fmt.Errorf("action %s:%s:%d: %w", actionPlugin.General.Name, actionName, i, err)
		}
	}
	return nil
}

func (c *CallCmd) Stop() {
	c.cancel()
	if c.running != nil {
		c.running.Stop()
	}
}
