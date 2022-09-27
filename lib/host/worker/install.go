package worker

import (
	"context"
	"errors"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/PCCSuite/PCCPluginSys/lib/host/cmd"
	"github.com/PCCSuite/PCCPluginSys/lib/host/plugin"
)

type InstallingPlugin struct {
	Name       string
	ActionData *plugin.ActionData
	Plugin     *plugin.Plugin
	Ctx        context.Context
	cancel     context.CancelFunc
	cmd        cmd.Cmd
	Dependent  []Dependency
}

func (p *InstallingPlugin) WaitIsSucsess(ctx context.Context) bool {
	if p.Plugin.Installed {
		return true
	}
	if p.Plugin.ActionData.Status == plugin.ActionStatusFailed {
		return false
	}
	ch := make(chan plugin.ActionStatus)
	defer close(ch)
	p.ActionData.SubscribeStatus(ch)
	defer p.ActionData.UnsubscribeStatus(ch)
	for {
		select {
		case status := <-ch:
			if p.IsEnded() {
				if p.Plugin.Installed {
					return true
				}
				if status == plugin.ActionStatusFailed {
					return false
				}
				return false
			}
		case <-ctx.Done():
			return false
		}

	}
}

func (p *InstallingPlugin) IsEnded() bool {
	if p.Plugin.Installed {
		return true
	}
	if p.ActionData.Status == plugin.ActionStatusFailed {
		return true
	}
	return false
}

type Dependency struct {
	status *InstallingPlugin
	before bool
}

var InstallingPlugins map[string]*InstallingPlugin = make(map[string]*InstallingPlugin)

var starterRunning bool

var startQueue []*InstallingPlugin

func InstallPlugin(pluginName string, priority int) (*InstallingPlugin, error) {
	if installing, ok := InstallingPlugins[pluginName]; ok {
		if installing.IsEnded() {
			return installing, nil
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	installing := &InstallingPlugin{
		Name:       pluginName,
		ActionData: plugin.NewActionData(pluginName, plugin.ActionStatusRunning, "", priority, cancel),
		Ctx:        ctx,
		cancel:     cancel,
	}
	InstallingPlugins[pluginName] = installing
	splitName := strings.SplitN(pluginName, ":", 2)
	if len(splitName) == 1 {
		var err error
		installing.Plugin, err = plugin.SearchPlugin(pluginName)
		if err != nil {
			if errors.Is(err, plugin.ErrPluginNotFound) {
				installing.ActionData.SetActionStatusBoth(plugin.ActionStatusFailed, "plugin not found")
			} else {
				installing.ActionData.SetActionStatusBoth(plugin.ActionStatusFailed, "load failed")
			}
			return installing, err
		}
		installing.Plugin.ActionData = installing.ActionData
		installing.ActionData.Plugin = installing.Plugin
		for _, v := range installing.Plugin.Dependency.Dependent {
			depend, err := InstallPlugin(v.Name, priority)
			if err != nil {
				installing.ActionData.SetActionStatusBoth(plugin.ActionStatusFailed, "dependency failed")
				return installing, err
			}
			installing.Dependent = append(installing.Dependent, Dependency{status: depend, before: v.Before})
		}
	} else {
		depend, err := InstallPlugin(splitName[0], priority)
		if err != nil {
			installing.ActionData.SetActionStatusBoth(plugin.ActionStatusFailed, "dependency failed")
			return installing, err
		}
		installing.Dependent = append(installing.Dependent, Dependency{status: depend, before: true})
	}
	installing.ActionData.SetActionStatusOnly(plugin.ActionStatusWaitStart)
	startQueue = append(startQueue, installing)
	needStarter()
	return installing, nil
}

func needStarter() {
	if !starterRunning {
		starterRunning = true
		go starter()
	}
}

func starter() {
	defer func() {
		starterRunning = false
	}()
	for {
		if len(startQueue) == 0 {
			return
		}
		sort.SliceStable(startQueue, func(i, j int) bool {
			return startQueue[i].ActionData.Priority < startQueue[j].ActionData.Priority
		})
		go startQueue[0].start()
		startQueue = startQueue[1:]
		time.Sleep(1 * time.Second)
	}
}

func (p *InstallingPlugin) start() {
	err := p.waitDepend(true)
	if err != nil {
		return
	}
	p.ActionData.SetActionStatusBoth(plugin.ActionStatusRunning, "Checking directories")
	var newInstall bool
	_, err = os.Stat(p.Plugin.GetDataDir())
	if err != nil {
		newInstall = true
		err = os.MkdirAll(p.Plugin.GetDataDir(), os.ModeDir)
		if err != nil {
			p.ActionData.SetActionStatusBoth(plugin.ActionStatusFailed, "Failed to make datadir")
			return
		}
	}
	err = os.MkdirAll(p.Plugin.GetTempDir(), os.ModeDir)
	if err != nil {
		p.ActionData.SetActionStatusBoth(plugin.ActionStatusFailed, "Failed to make tempdir")
		return
	}
	var call *cmd.CallCmd
	if newInstall {
		p.ActionData.SetActionStatusBoth(plugin.ActionStatusRunning, "Running action: "+plugin.ActionNewInstall)
		call = cmd.NewCallCmd(p.Plugin, []string{plugin.ActionNewInstall}, p.Ctx)
	} else {
		p.ActionData.SetActionStatusBoth(plugin.ActionStatusRunning, "Running action: "+plugin.ActionRestore)
		call = cmd.NewCallCmd(p.Plugin, []string{plugin.ActionRestore}, p.Ctx)
	}
	err = call.Run()
	if err != nil {
		p.ActionData.SetActionStatusBoth(plugin.ActionStatusFailed, "Error: "+err.Error())
		return
	}
	err = p.waitDepend(false)
	if err != nil {
		return
	}
	p.Plugin.Installed = true
	p.ActionData.SetActionStatusBoth(plugin.ActionStatusDone, "")
}

func (p *InstallingPlugin) Stop() {
	p.cancel()
	if p.cmd != nil {
		p.cmd.Stop()
	}
}

var ErrDependencyFailed = errors.New("dependency failed")

func (p *InstallingPlugin) waitDepend(before bool) error {
	for _, v := range p.Dependent {
		p.ActionData.SetActionStatusBoth(plugin.ActionStatusRunning, "Checking dependency: "+v.status.Name)
		if v.before || !before {
			installing, ok := InstallingPlugins[v.status.Name]
			if !ok {
				var err error
				installing, err = InstallPlugin(v.status.Name, p.ActionData.Priority)
				if err != nil {
					p.ActionData.SetActionStatusBoth(plugin.ActionStatusFailed, "dependency failed")
					return ErrDependencyFailed
				}
			}
			if !installing.IsEnded() {
				p.ActionData.SetActionStatusBoth(plugin.ActionStatusWaitDepend, "Waiting for '"+v.status.Name+"'")
			}
			ok = installing.WaitIsSucsess(p.Ctx)
			select {
			case <-p.Ctx.Done():
				p.ActionData.SetActionStatusBoth(plugin.ActionStatusFailed, "Stopped")
				return cmd.ErrStopped
			default:
			}
			if !ok {
				p.ActionData.SetActionStatusBoth(plugin.ActionStatusFailed, "dependency failed")
				return ErrDependencyFailed
			}
		}
	}
	return nil
}
