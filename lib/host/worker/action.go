package worker

import (
	"context"
	"errors"
	"strings"

	"github.com/PCCSuite/PCCPluginSys/lib/host/cmd"
	"github.com/PCCSuite/PCCPluginSys/lib/host/data"
	"github.com/PCCSuite/PCCPluginSys/lib/host/lock"
)

var ErrNotInstalled = errors.New("data not installed")
var ErrDirectOtherAction = errors.New("calling other action direct is not allowed")

func RunAction(pluginName string, action string, priority int, ctx context.Context) error {
	pl := data.GetPlugin(pluginName)
	if pl == nil {
		return data.ErrPluginNotFound
	}
	if !pl.Installed {
		return ErrNotInstalled
	}
	if strings.ContainsRune(action, ':') {
		return ErrDirectOtherAction
	}
	ctx, cancel := context.WithCancel(ctx)
	pl.RunningAction.Cancel = cancel
	pl.RunningAction.Priority = priority
	pl.RunningAction.SetActionStatusBoth(data.ActionStatusRunning, "Running action: "+action)
	call := cmd.NewCallCmd(pl.Package, []string{action}, ctx)
	err := call.Run()
	lock.UnlockAll(pl.RunningAction)
	if err != nil {
		pl.RunningAction.SetActionStatusBoth(data.ActionStatusFailed, "Error: "+err.Error())
	} else {
		pl.RunningAction.SetActionStatusBoth(data.ActionStatusDone, "")
	}
	return err
}
