package data

import "context"

var InstallingPackages map[*Package]*InstallingPackage = map[*Package]*InstallingPackage{}

type InstallingPackage struct {
	Status *RunningAction

	Dependent []InstallingDependency
}

func (p *InstallingPackage) WaitIsSucsess(ctx context.Context) bool {
	if p.Status.Package == nil {
		return false
	}
	if p.Status.Package.Installed {
		return true
	}
	if p.Status.Package.RunningAction.Status == ActionStatusFailed {
		return false
	}
	ch := make(chan ActionStatus)
	defer close(ch)
	p.Status.Package.RunningAction.SubscribeStatus(ch)
	defer p.Status.Package.RunningAction.UnsubscribeStatus(ch)
	for {
		select {
		case status := <-ch:
			if p.Status.Package.Installed {
				return true
			}
			if status == ActionStatusFailed {
				return false
			}
		case <-ctx.Done():
			return false
		}

	}
}

func (p *InstallingPackage) IsEnded() bool {
	if p.Status.Package.Installed {
		return true
	}
	if p.Status.Package.RunningAction.Status == ActionStatusFailed {
		return true
	}
	return false
}

type InstallingDependency struct {
	Status *InstallingPackage
	Before bool
}
