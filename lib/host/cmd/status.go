package cmd

import (
	"strings"

	"github.com/PCCSuite/PCCPluginSys/lib/host/plugin"
)

const STATUS = "STATUS"

type StatusCmd struct {
	plugin *plugin.Plugin
	param  []string
}

func NewStatusCmd(plugin *plugin.Plugin, param []string) *StatusCmd {
	return &StatusCmd{
		plugin: plugin,
		param:  param,
	}
}

func (c *StatusCmd) Run() error {
	text := strings.Join(c.param, " ")
	c.plugin.ActionData.SetActionStatusText(text)
	return nil
}

func (c *StatusCmd) Stop() {
}
