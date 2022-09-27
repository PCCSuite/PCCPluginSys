package worker

import (
	"context"
	"errors"

	"github.com/PCCSuite/PCCPluginSys/lib/host/cmd"
	"github.com/PCCSuite/PCCPluginSys/lib/host/plugin"
)

var ErrNotInstalled = errors.New("plugin not installed")

func RunAction(pluginName string, action string, priority int, ctx context.Context) error {
	pl, err := plugin.SearchPlugin(pluginName)
	if err != nil {
		return err
	}
	if !pl.Installed {
		return ErrNotInstalled
	}
	ctx, cancel := context.WithCancel(ctx)
	pl.ActionData.Stop = cancel
	pl.ActionData.Priority = priority
	pl.ActionData.SetActionStatusBoth(plugin.ActionStatusRunning, "Running action: "+plugin.ActionNewInstall)
	call := cmd.NewCallCmd(pl, []string{plugin.ActionNewInstall}, ctx)
	err = call.Run()
	if err != nil {
		pl.ActionData.SetActionStatusBoth(plugin.ActionStatusFailed, "Error: "+err.Error())
	} else {
		pl.ActionData.SetActionStatusBoth(plugin.ActionStatusDone, "")
	}
	return err
}
