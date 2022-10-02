package pccclient

import (
	"github.com/PCCSuite/PCCPluginSys/lib/host/data"
	"github.com/PCCSuite/PCCPluginSys/lib/host/status"
)

type DataType string

const (
	DataTypeNotify DataType = "notify"

	DataTypeRestore DataType = "restore"
	DataTypeInstall DataType = "install"
	DataTypeAction  DataType = "action"
)

type ClientNotifyData struct {
	Data_type DataType         `json:"data_type"`
	Status    status.SysStatus `json:"status"`
	Plugins   []PluginData     `json:"plugins"`
}

func NewClientNotifyData(status status.SysStatus, plugins []PluginData) ClientNotifyData {
	return ClientNotifyData{
		Data_type: DataTypeNotify,
		Status:    status,
		Plugins:   plugins,
	}
}

type PluginData struct {
	Name       string            `json:"name"`
	Repository string            `json:"repository"`
	Installed  bool              `json:"installed"`
	Locking    bool              `json:"locking"`
	Status     data.ActionStatus `json:"status"`
	StatusText string            `json:"status_text"`
	Dependency []string          `json:"dependency"`
}

func NewPluginData(name, repository string, installed, locking bool, status data.ActionStatus, statusText string, dependency []string) PluginData {
	return PluginData{
		Name:       name,
		Repository: repository,
		Installed:  installed,
		Locking:    locking,
		Status:     status,
		StatusText: statusText,
		Dependency: dependency,
	}
}

type CommandData struct {
	Data_type DataType `json:"data_type"`
}

type InstallCommandData struct {
	Data_type DataType `json:"data_type"`
	Plugin    string   `json:"plugin"`
}

type ActionCommandData struct {
	Data_type DataType `json:"data_type"`
	Plugin    string   `json:"plugin"`
	Action    string   `json:"action"`
}
