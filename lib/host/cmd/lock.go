package cmd

import (
	"context"

	"github.com/PCCSuite/PCCPluginSys/lib/host/data"
	"github.com/PCCSuite/PCCPluginSys/lib/host/lock"
)

const LOCK = "LOCK"

type LockCmd struct {
	Package *data.Package
	param   []string
	ctx     context.Context
}

func NewLockCmd(Package *data.Package, param []string, ctx context.Context) *LockCmd {
	return &LockCmd{
		Package: Package,
		param:   param,
		ctx:     ctx,
	}
}

func (c *LockCmd) Run() error {
	param := c.param
	name := lock.DefaultName
	locking := true
paramcheck:
	for {
		if len(param) == 0 {
			break
		}
		switch param[0] {
		case "/LOCK":
			param = param[1:]
			locking = true
		case "/UNLOCK":
			param = param[1:]
			locking = false
		default:
			name = param[0]
			param = param[1:]
			break paramcheck
		}
	}
	if len(param) > 0 {
		return ErrTooMuchArgs
	}
	if locking {
		req := lock.RequestLock(name, c.Package.RunningAction)
		select {
		case <-req.Ch:
			return req.Err
		case <-c.ctx.Done():
			return ErrStopped
		}
	} else {
		lock.Unlock(name, c.Package.RunningAction)
		return nil
	}
}
