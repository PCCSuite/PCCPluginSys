package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/PCCSuite/PCCPluginSys/lib/host/data"
)

type Cmd interface {
	Run() error
}

var ErrStopped = errors.New("stopped")

var ErrTooFewArgs = errors.New("too few arguments")
var ErrTooMuchArgs = errors.New("too much arguments")

var ErrParseParam = errors.New("failed to parse action parameter")

func parseParam(p string) ([]string, error) {
	split := strings.Split(p, " ")
	buf := ""
	quote := ""
	var result []string
	for _, v := range split {
		if quote != "" {
			if strings.HasSuffix(v, quote) {
				result = append(result, buf+strings.TrimSuffix(v, quote))
				buf = ""
				quote = ""
			} else {
				buf += v + " "
			}
		} else {
			if buf == "" {
				if strings.HasPrefix(v, "\"") {
					quote = "\""
				} else if strings.HasPrefix(v, "'") {
					quote = "'"
				}
			}
			if quote != "" {
				if strings.HasSuffix(v, quote) {
					result = append(result, strings.TrimPrefix(strings.TrimSuffix(v, quote), quote))
				} else {
					buf += strings.TrimPrefix(v, quote) + " "
				}
			} else {
				result = append(result, v)
			}
		}
	}
	if buf != "" {
		return nil, ErrParseParam
	}
	return result, nil
}

func replaceParams(p []string, Package *data.Package, plugin *data.Plugin, callArgs []string) []string {
	result := make([]string, 0)
	for _, v := range p {
		v = strings.ReplaceAll(v, "${plugin_starter}", Package.Name)
		v = strings.ReplaceAll(v, "${plugin_name}", plugin.General.Name)
		v = strings.ReplaceAll(v, "${plugin_repodir}", plugin.GetRepoDir())
		v = strings.ReplaceAll(v, "${plugin_datadir}", plugin.GetDataDir())
		v = strings.ReplaceAll(v, "${plugin_tempdir}", plugin.GetTempDir())
		v = strings.ReplaceAll(v, "${arg}", strings.Join(callArgs, " "))
		for i := 0; i < 10; i++ {
			if len(callArgs) > i {
				v = strings.ReplaceAll(v, "${arg"+fmt.Sprint(i)+"}", callArgs[i])
			} else {
				v = strings.ReplaceAll(v, "${arg"+fmt.Sprint(i)+"}", "")
			}
		}
		if strings.Contains(v, "${args}") {
			split := strings.SplitN(v, "${args}", 2)
			for i2, v2 := range callArgs {
				if i2 == 0 {
					v2 = split[0] + v2
				}
				if i2 == len(callArgs)-1 {
					v2 = v2 + split[1]
				}
				result = append(result, v2)
			}
		} else {
			result = append(result, v)
		}
	}
	return result
}

func ToCmd(Package *data.Package, plugin *data.Plugin, name string, param []string, ctx context.Context) (Cmd, error) {
	switch strings.ToUpper(name) {
	case CALL:
		return NewCallCmd(Package, param, ctx), nil
	case ENV:
		return NewEnvCmd(Package, plugin, param, ctx), nil
	case EXEC:
		return NewExecCmd(Package, plugin, param, ctx), nil
	case LOCK:
		return NewLockCmd(Package, param, ctx), nil
	case ASK:
		return NewAskCmd(Package, plugin, param, ctx), nil
	default:
		return nil, ErrCommandNotFound
	}
}
