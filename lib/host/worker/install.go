package worker

import (
	"context"
	"errors"
	"os"
	"sort"
	"strings"

	"github.com/PCCSuite/PCCPluginSys/lib/host/cmd"
	"github.com/PCCSuite/PCCPluginSys/lib/host/plugin"
)

type InstallingPlugin struct {
	Name      string
	Plugin    *plugin.Plugin
	Priority  int
	Ctx       context.Context
	cancel    context.CancelFunc
	cmd       cmd.Cmd
	Dependent []Dependency
}

func (p *InstallingPlugin) WaitIsSucsess() bool {
	if p.Plugin.Installed {
		return true
	}
	if *p.Plugin.ActionStatus.Status == plugin.ActionStatusFailed {
		return false
	}
	ch := make(chan plugin.ActionStatusSet)
	defer close(ch)
	p.Plugin.SubscribeStatus(ch)
	defer p.Plugin.UnsubscribeStatus(ch)
	for {
		status := <-ch
		if p.IsEnded() {
			if p.Plugin.Installed {
				return true
			}
			if *status.Status == plugin.ActionStatusFailed {
				return false
			}
			return false
		}
	}
}

func (p *InstallingPlugin) IsEnded() bool {
	if p.Plugin.Installed {
		return true
	}
	if *p.Plugin.ActionStatus.Status == plugin.ActionStatusDone {
		return true
	}
	if *p.Plugin.ActionStatus.Status == plugin.ActionStatusFailed {
		return true
	}
	return false
}

type Dependency struct {
	status *InstallingPlugin
	before bool
}

var InstallingPlugins map[string]*InstallingPlugin

var starterRunning bool

var startQueue []*InstallingPlugin

// error
func InstallPlugin(pluginName string, priority int) (*InstallingPlugin, error) {
	ctx, cancel := context.WithCancel(context.Background())
	installing := &InstallingPlugin{
		Name:     pluginName,
		Priority: priority,
		Ctx:      ctx,
		cancel:   cancel,
	}
	InstallingPlugins[pluginName] = installing
	splitName := strings.SplitN(pluginName, ":", 2)
	if len(splitName) == 1 {
		installingPlugin, err := plugin.SearchPlugin(pluginName)
		if err != nil {
			if errors.Is(err, plugin.ErrPluginNotFound) {
				installing.Plugin.SetActionStatusBoth(plugin.ActionStatusFailed, "plugin not found")
			} else {
				installing.Plugin.SetActionStatusBoth(plugin.ActionStatusFailed, "load failed")
			}
			return installing, err
		}
		installing.Plugin = installingPlugin
		for _, v := range installingPlugin.Dependency.Dependent {
			depend, err := InstallPlugin(v.Name, priority)
			if err != nil {
				installing.Plugin.SetActionStatusBoth(plugin.ActionStatusFailed, "dependency failed")
				return installing, err
			}
			installing.Dependent = append(installing.Dependent, Dependency{status: depend, before: v.Before})
		}
	} else {
		depend, err := InstallPlugin(splitName[0], priority)
		if err != nil {
			installing.Plugin.SetActionStatusBoth(plugin.ActionStatusFailed, "dependency failed")
			return installing, err
		}
		installing.Dependent = append(installing.Dependent, Dependency{status: depend, before: true})
	}
	installing.Plugin.SetActionStatusOnly(plugin.ActionStatusWaitStart)
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

var installerChan map[string]chan string

func starter() {
	defer func() {
		starterRunning = false
	}()
	for {
		if len(startQueue) == 0 {
			return
		}
		sort.SliceStable(startQueue, func(i, j int) bool {
			return startQueue[i].Priority < startQueue[j].Priority
		})
		go startQueue[0].start()
	}
}

func (p *InstallingPlugin) start() {
	err := p.waitDepend(true)
	if err != nil {
		return
	}
	p.Plugin.SetActionStatusBoth(plugin.ActionStatusRunning, "Checking directories")
	var newInstall bool
	_, err = os.Stat(p.Plugin.GetDataDir())
	if err != nil {
		newInstall = true
		err = os.MkdirAll(p.Plugin.GetDataDir(), os.ModeDir)
		if err != nil {
			p.Plugin.SetActionStatusBoth(plugin.ActionStatusFailed, "Failed to make datadir")
			return
		}
	}
	err = os.MkdirAll(p.Plugin.GetTempDir(), os.ModeDir)
	if err != nil {
		p.Plugin.SetActionStatusBoth(plugin.ActionStatusFailed, "Failed to make tempdir")
		return
	}
	var call *cmd.CallCmd
	if newInstall {
		call = cmd.NewCallCmd(p.Plugin, []string{"install"}, p.Ctx)
	} else {
		call = cmd.NewCallCmd(p.Plugin, []string{"restore"}, p.Ctx)
	}
	err = call.Run()
	p.Plugin.SetActionStatusBoth(plugin.ActionStatusFailed, "Error: "+err.Error())
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
		if v.before || !before {
			installing, ok := InstallingPlugins[v.status.Name]
			if !ok {
				var err error
				installing, err = InstallPlugin(v.status.Name, p.Priority)
				if err != nil {
					p.Plugin.SetActionStatusBoth(plugin.ActionStatusFailed, "dependency failed")
					return ErrDependencyFailed
				}
			}
			if !installing.IsEnded() {
				p.Plugin.SetActionStatusBoth(plugin.ActionStatusWaitDepend, "Waiting for '"+v.status.Name+"'")
			}
			ok = installing.WaitIsSucsess()
			if !ok {
				p.Plugin.SetActionStatusBoth(plugin.ActionStatusFailed, "dependency failed")
				return ErrDependencyFailed
			}
			p.Plugin.SetActionStatusBoth(plugin.ActionStatusRunning, "Checking dependency")
		}
	}
	return nil
}
