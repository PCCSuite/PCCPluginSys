package worker

import (
	"context"
	"errors"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/PCCSuite/PCCPluginSys/lib/host/cmd"
	"github.com/PCCSuite/PCCPluginSys/lib/host/data"
)

var starterRunning bool

var startQueue []*data.InstallingPackage

var ErrRepoNotFound = errors.New("repository not found")
var ErrInstalledOtherRepo = errors.New("plugin already installed from other repository")
var ErrRepoAlreadyExist = errors.New("same name repository already exists")

func InstallPackage(packageIdentifier string, priority int) (*data.InstallingPackage, error) {
	var Package *data.Package
	var packageName string
	var repoName string
	var repo *data.Repository

	//
	// Check package already installing
	//

	status, ok := data.RunningActions[packageIdentifier]

	if ok {
		Package = status.Package
		packageName = status.Package.Name
		repoName = status.Package.Repo.Name
		repo = status.Package.Repo
	} else {
		if splitName := strings.SplitN(packageIdentifier, ":", 2); len(splitName) == 1 {
			packageName = packageIdentifier
		} else {
			repoName = splitName[0]
			packageName = splitName[1]
		}
		if repoName != "" {
			repo = data.Repositories[repoName]
			if repo == nil || repo.Type == data.RepositoryTypeExternal {
				// external repo
				Package = data.GetExternalPackage(repoName, packageName)
			} else {
				// internal repo
				plugin := data.GetPlugin(packageName)
				if plugin != nil {
					Package = plugin.Package
					if (Package.Installed || !Package.RunningAction.IsEnded()) && Package.Repo.Name != repoName {
						status.SetActionStatusBoth(data.ActionStatusFailed, "already instaled from other repository")
						return nil, ErrInstalledOtherRepo
					}
				}
			}
		} else {
			// repo is not specify
			plugin := data.GetPlugin(packageName)
			if plugin != nil {
				Package = plugin.Package
			}
		}
		ctx, cancel := context.WithCancel(context.Background())
		status = data.NewRunningAction(packageIdentifier, data.ActionStatusRunning, "", priority, ctx, cancel)
	}

	// if package already loaded
	if Package != nil {
		installing, ok := data.InstallingPackages[Package]
		if ok && !installing.IsEnded() {
			return installing, nil
		}
	}

	//
	// Start install
	//

	installing := &data.InstallingPackage{
		Status:    status,
		Dependent: make([]data.InstallingDependency, 0),
	}
	if repoName != "" && (repo == nil || repo.Type == data.RepositoryTypeExternal) {
		// if external
		depend, err := InstallPackage(repoName, priority)
		if err != nil {
			status.SetActionStatusBoth(data.ActionStatusFailed, "dependency failed")
			return nil, err
		}
		Package = data.NewExternalPackage(packageName, data.Repositories[repoName])
		installing.Dependent = append(installing.Dependent, data.InstallingDependency{Status: depend, Before: true})
	} else {
		// if internal
		var plugin *data.Plugin
		var err error
		if repo != nil {
			plugin, err = data.LoadPlugin(repo, packageName)
		} else {
			plugin, err = data.SearchPlugin(packageName)
		}
		if err != nil {
			if errors.Is(err, data.ErrPluginNotFound) {
				status.SetActionStatusBoth(data.ActionStatusFailed, "plugin not found")
			} else {
				status.SetActionStatusBoth(data.ActionStatusFailed, "load failed")
			}
			return installing, err
		}
		Package = plugin.Package
		for _, v := range plugin.Dependency.Dependent {
			depend, err := InstallPackage(v.Name, priority)
			if err != nil {
				status.SetActionStatusBoth(data.ActionStatusFailed, "dependency failed")
				return installing, err
			}
			installing.Dependent = append(installing.Dependent, data.InstallingDependency{Status: depend, Before: v.Before})
		}
		if plugin.GetAction(data.ActionExternal) != "" {
			thisRepo, ok := data.Repositories[plugin.Name]
			if ok {
				if thisRepo.Source != plugin {
					status.SetActionStatusBoth(data.ActionStatusFailed, ErrRepoAlreadyExist.Error())
					return installing, ErrRepoAlreadyExist
				}
			} else {
				data.NewExternalRepository(plugin)
			}
		}
	}
	Package.RunningAction = status
	status.Package = Package
	data.InstallingPackages[Package] = installing
	status.SetActionStatusOnly(data.ActionStatusWaitStart)
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
			return startQueue[i].Status.Priority < startQueue[j].Status.Priority
		})
		go start(startQueue[0])
		startQueue = startQueue[1:]
		time.Sleep(1 * time.Second)
	}
}

func start(p *data.InstallingPackage) {
	err := waitDepend(p, true)
	if err != nil {
		return
	}
	var newInstall bool
	if p.Status.Package.Type != data.PackageTypeExternal {
		p.Status.SetActionStatusBoth(data.ActionStatusRunning, "Checking directories")
		_, err = os.Stat(p.Status.Package.Plugin.GetDataDir())
		if err != nil {
			newInstall = true
			err = os.MkdirAll(p.Status.Package.Plugin.GetDataDir(), os.ModeDir)
			if err != nil {
				p.Status.SetActionStatusBoth(data.ActionStatusFailed, "Failed to make datadir")
				return
			}
		}
		err = os.MkdirAll(p.Status.Package.Plugin.GetTempDir(), os.ModeDir)
		if err != nil {
			p.Status.SetActionStatusBoth(data.ActionStatusFailed, "Failed to make tempdir")
			return
		}
	}
	var action string
	if p.Status.Package.Type == data.PackageTypeExternal {
		action = data.ActionExternal
	} else if newInstall {
		action = data.ActionNewInstall
	} else {
		action = data.ActionRestore
	}
	p.Status.SetActionStatusBoth(data.ActionStatusRunning, "Running action: "+action)
	call := cmd.NewCallCmd(p.Status.Package, []string{action}, p.Status.Ctx)
	err = call.Run()
	if err != nil {
		p.Status.SetActionStatusBoth(data.ActionStatusFailed, "Error: "+err.Error())
		return
	}
	err = waitDepend(p, false)
	if err != nil {
		return
	}
	p.Status.Package.Installed = true
	p.Status.SetActionStatusBoth(data.ActionStatusDone, "")
}

func Stop(p *data.InstallingPackage) {
	p.Status.Cancel()
}

var ErrDependencyFailed = errors.New("dependency failed")

func waitDepend(p *data.InstallingPackage, before bool) error {
	for _, v := range p.Dependent {
		p.Status.SetActionStatusBoth(data.ActionStatusRunning, "Checking dependency: "+v.Status.Status.PackageIdentifier)
		if v.Before || !before {
			installing, ok := data.InstallingPackages[v.Status.Status.Package]
			if !ok {
				log.Panic("Waiting for not installing package")
			}
			if !installing.IsEnded() {
				p.Status.SetActionStatusBoth(data.ActionStatusWaitDepend, "Waiting for '"+v.Status.Status.PackageIdentifier+"'")
			}
			ok = installing.WaitIsSucsess(p.Status.Ctx)
			select {
			case <-p.Status.Ctx.Done():
				p.Status.SetActionStatusBoth(data.ActionStatusFailed, "Stopped")
				return cmd.ErrStopped
			default:
			}
			if !ok {
				p.Status.SetActionStatusBoth(data.ActionStatusFailed, "dependency failed")
				return ErrDependencyFailed
			}
		}
	}
	return nil
}
