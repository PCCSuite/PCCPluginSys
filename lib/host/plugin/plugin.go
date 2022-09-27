package plugin

import (
	"encoding/xml"
	"errors"
	"os"
	"path/filepath"

	"github.com/PCCSuite/PCCPluginSys/lib/host/config"
)

var ErrPluginNotFound = errors.New("Plugin is not found in any repositories")

var ErrNameNotEqual = errors.New("Plugin name is not equal to dir name")

var ActionRestore = "restore"
var ActionNewInstall = "install"

type Plugin struct {
	Xml_version int              `xml:"plugin_xml_version"`
	General     PluginGeneral    `xml:"general"`
	Dependency  PluginDependency `xml:"dependency"`
	Actions     PluginActions    `xml:"actions"`
	RepoDir     string
	Installed   bool
	ActionData  *ActionData
}

func (p *Plugin) GetRepoDir() string {
	return p.RepoDir
}

func (p *Plugin) GetDataDir() string {
	return filepath.Join(config.Config.DataDir, p.General.Name)
}

func (p *Plugin) GetTempDir() string {
	return filepath.Join(config.Config.TempDir, p.General.Name)
}

func (p *Plugin) GetAction(name string) string {
	for _, v := range p.Actions.Actions {
		if v.XMLName.Local == name {
			return v.Content
		}
	}
	return ""
}

type PluginGeneral struct {
	Name        string `xml:"name"`
	Version     string `xml:"version"`
	Desctiprion string `xml:"description"`
	Author      string `xml:"author"`
	Licence     string `xml:"licence"`
	// Buttons     map[string]string `xml:"buttons"`
}

type PluginDependency struct {
	Dependent []PluginDependent `xml:"dependent"`
}

type PluginDependent struct {
	Name   string `xml:",chardata"`
	Before bool   `xml:"before,attr"` // trueの場合、actionの実行前に待機します。false(初期値)の場合、actionと並行してインストールします。
}

type PluginActions struct {
	Actions []PluginAction `xml:",any"`
}

type PluginAction struct {
	XMLName xml.Name
	Content string `xml:",chardata"`
}

var Plugins []*Plugin

func SearchPlugin(name string) (*Plugin, error) {
	for _, v := range Plugins {
		if v.General.Name == name {
			return v, nil
		}
	}
	for _, v := range config.Config.Repositories {
		path := filepath.Join(v, name)
		_, err := os.Stat(path)
		if err == nil {
			return LoadPlugin(path)
		}
	}
	return nil, ErrPluginNotFound
}

func LoadPlugin(path string) (*Plugin, error) {
	file, err := os.ReadFile(filepath.Join(path, "plugin.xml"))
	if err != nil {
		return nil, err
	}
	plugin := Plugin{}
	err = xml.Unmarshal(file, &plugin)
	if err != nil {
		return nil, err
	}
	if plugin.General.Name != filepath.Base(path) {
		return nil, ErrNameNotEqual
	}
	plugin.RepoDir = path
	Plugins = append(Plugins, &plugin)
	return &plugin, nil
}
