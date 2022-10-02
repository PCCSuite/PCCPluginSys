package cmd

import (
	"strings"

	"github.com/PCCSuite/PCCPluginSys/lib/host/data"
)

const STATUS = "STATUS"

type StatusCmd struct {
	Package *data.Package
	param   []string
}

func NewStatusCmd(Package *data.Package, param []string) *StatusCmd {
	return &StatusCmd{
		Package: Package,
		param:   param,
	}
}

func (c *StatusCmd) Run() error {
	text := strings.Join(c.param, " ")
	c.Package.RunningAction.SetActionStatusText(text)
	return nil
}
