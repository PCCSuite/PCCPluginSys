package pccclient

import (
	"github.com/PCCSuite/PCCPluginSys/lib/host/cmd"
	"github.com/PCCSuite/PCCPluginSys/lib/host/data"
	"github.com/PCCSuite/PCCPluginSys/lib/host/status"
)

type DataType string

const (
	DataTypeNotify DataType = "notify"

	DataTypeRestore DataType = "restore"
	DataTypeInstall DataType = "install"
	DataTypeAction  DataType = "action"
	DataTypeCancel  DataType = "cancel"
	DataTypeAnswer  DataType = "answer"
)

type ClientNotifyData struct {
	Data_type DataType         `json:"data_type"`
	Status    status.SysStatus `json:"status"`
	Packages  []PackageData    `json:"packages"`
	Asking    []*cmd.AskData   `json:"asking"`
}

func NewClientNotifyData(status status.SysStatus, plugins []PackageData, asking []*cmd.AskData) ClientNotifyData {
	return ClientNotifyData{
		Data_type: DataTypeNotify,
		Status:    status,
		Packages:  plugins,
		Asking:    asking,
	}
}

type PackageData struct {
	Identifier string            `json:"identifier"`
	Repository string            `json:"repository"`
	Installed  bool              `json:"installed"`
	Locking    bool              `json:"locking"`
	Status     data.ActionStatus `json:"status"`
	StatusText string            `json:"status_text"`
	Priority   int               `json:"priority"`
	Dependency []string          `json:"dependency"`
}

func NewPackageData(name, repository string, installed, locking bool, status data.ActionStatus, statusText string, priority int, dependency []string) PackageData {
	return PackageData{
		Identifier: name,
		Repository: repository,
		Installed:  installed,
		Locking:    locking,
		Status:     status,
		StatusText: statusText,
		Priority:   priority,
		Dependency: dependency,
	}
}

type CommandData struct {
	Data_type DataType `json:"data_type"`

	// Install, Cancel
	Package string `json:"package"`

	// Action
	Plugin string `json:"plugin"`
	// Action
	Action string `json:"action"`

	// Answer
	ID    int    `json:"id"`
	Value string `json:"value"`
}
