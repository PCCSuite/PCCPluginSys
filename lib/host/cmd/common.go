package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/PCCSuite/PCCPluginSys/lib/host/plugin"
)

type Cmd interface {
	Run() error
	Stop()
}

var ErrStopped = errors.New("Stopped running cmd")

var ErrTooFewArgs = errors.New("too few arguments")

var ErrParseParam = errors.New("Failed to parse action parameter")

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

func replaceParams(p []string, pluginStarter *plugin.Plugin, pluginActioner *plugin.Plugin, callArgs []string) []string {
	result := make([]string, len(p))
	for _, v := range p {
		v = strings.ReplaceAll(v, "${plguin_starter}", pluginStarter.General.Name)
		v = strings.ReplaceAll(v, "${plguin_name}", pluginActioner.General.Name)
		v = strings.ReplaceAll(v, "${plguin_repodir}", pluginActioner.GetRepoDir())
		v = strings.ReplaceAll(v, "${plguin_datadir}", pluginActioner.GetDataDir())
		v = strings.ReplaceAll(v, "${plguin_tempdir}", pluginActioner.GetTempDir())
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
