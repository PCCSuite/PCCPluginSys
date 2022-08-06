package cmd

import (
	"context"

	"github.com/PCCSuite/PCCPluginSys/lib/host/plugin"
)

const STATUS = "STATUS"

type StatusCmd struct {
	plugin *plugin.Plugin
	param  []string
	ctx    context.Context
}

func NewStatusCmd(plugin *plugin.Plugin, param []string, ctx context.Context) *CallCmd {
	return &CallCmd{
		plugin: plugin,
		param:  param,
		ctx:    ctx,
	}
}

// func (c *StatusCmd) run() error {

// }
