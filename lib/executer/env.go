package executer

import (
	"log"

	"golang.org/x/sys/windows/registry"
)

func GetSystemEnv() []string {
	machine := getSystemEnvs(true)
	user := getSystemEnvs(false)
	for k := range user {
		if k == "Path" {
			machine[k] = machine[k] + ";" + user[k]
		} else {
			machine[k] = user[k]
		}
	}
	var res []string = make([]string, 0)
	for k, v := range machine {
		res = append(res, k+"="+v)
	}
	return res
}

func getSystemEnvs(machine bool) map[string]string {
	res := map[string]string{}
	var key registry.Key
	var err error
	if machine {
		key, err = registry.OpenKey(registry.LOCAL_MACHINE, "SYSTEM\\CurrentControlSet\\Control\\Session Manager\\Environment", registry.READ)
	} else {
		key, err = registry.OpenKey(registry.CURRENT_USER, "Environment", registry.READ)
	}
	if err != nil {
		log.Panicln("Failed to open registry key: ", err)
	}
	values, err := key.ReadValueNames(0)
	if err != nil {
		log.Panicln("Failed to read registry values list: ", err)
	}
	for _, v := range values {
		data, valType, err := key.GetStringValue(v)
		if err != nil {
			log.Panicln("Failed to get value: ", err)
		}
		if valType == registry.EXPAND_SZ {
			data, err = registry.ExpandString(data)
			if err != nil {
				log.Panicln("Failed to expand value: ", err)
			}
		}
		res[v] = data
	}
	return res
}
