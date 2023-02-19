package executer

import (
	"log"
	"strings"
	"syscall"
	"unsafe"

	"github.com/PCCSuite/PCCPluginSys/lib/common"
	"golang.org/x/sys/windows/registry"
)

const (
	hwnd_BROADCAST   = uintptr(0xffff)
	wm_SETTINGCHANGE = uintptr(0x001A)
)

func Env(cmd common.ExecuterEnvData) {
	var regKey registry.Key
	var err error
	if cmd.Target == common.ExecuterEnvTargetMachine {
		regKey, err = registry.OpenKey(registry.LOCAL_MACHINE, "SYSTEM\\CurrentControlSet\\Control\\Session Manager\\Environment", registry.READ|registry.SET_VALUE)
	} else {
		regKey, err = registry.OpenKey(registry.CURRENT_USER, "Environment", registry.READ|registry.SET_VALUE)
	}
	if err != nil {
		log.Panicln("Failed to open registry key: ", err)
	}
	values, err := regKey.ReadValueNames(0)
	if err != nil {
		log.Panicln("Failed to read registry values list: ", err)
	}
	expand := false
	valueName := cmd.Key
	valueData := ""
	for _, v := range values {
		if strings.EqualFold(valueName, v) {
			valueName = v
			data, valType, err := regKey.GetStringValue(v)
			if err != nil {
				log.Println("Failed to get value: ", err)
				send(common.NewExecuterResult(1, cmd.RequestId))
				return
			}
			if valType == registry.EXPAND_SZ {
				expand = true
			}
			valueData = data
			break
		}
	}
	switch cmd.Mode {
	case common.ExecuterEnvModeSet:
		valueData = cmd.Value
	case common.ExecuterEnvModeAdd:
		if valueData != "" {
			valueData = valueData + ";" + cmd.Value
		} else {
			valueData = cmd.Value
		}
	case common.ExecuterEnvModeAddPrefix:
		if valueData != "" {
			valueData = cmd.Value + ";" + valueData
		} else {
			valueData = cmd.Value
		}
	}
	if expand {
		err = regKey.SetExpandStringValue(valueName, valueData)
	} else {
		err = regKey.SetStringValue(valueName, valueData)
	}
	if err != nil {
		log.Println("Failed to set value: ", err)
		send(common.NewExecuterResult(1, cmd.RequestId))
		return
	}
	// notify env change
	// https://gist.github.com/microo8/c1b9525efab9bb462adf9d123e855c52
	envPtr, err := syscall.UTF16PtrFromString("ENVIRONMENT")
	if err != nil {
		log.Println("Failed to get ptr from ENVIRONMENT string: ", err)
		send(common.NewExecuterResult(1, cmd.RequestId))
		return
	}

	syscall.NewLazyDLL("user32.dll").NewProc("SendMessageW").Call(hwnd_BROADCAST, wm_SETTINGCHANGE, 0, uintptr(unsafe.Pointer(envPtr)))
	// ignore error because they always return error "The operation completed successfully."

	send(common.NewExecuterResult(0, cmd.RequestId))
}
