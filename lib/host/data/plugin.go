package data

import (
	"encoding/xml"
	"errors"
	"os"
	"path/filepath"

	"github.com/PCCSuite/PCCPluginSys/lib/host/config"
)

var ErrPluginNotFound = errors.New("Plugin not found")

var ErrNameNotEqual = errors.New("Plugin name is not equal to dir name")

var ActionRestore = "restore"
var ActionNewInstall = "install"
var ActionExternal = "external"

var Plugins []*Plugin = make([]*Plugin, 0)

type Plugin struct {
	*Package
	Xml_version int              `xml:"plugin_xml_version"`
	General     PluginGeneral    `xml:"general"`
	Dependency  PluginDependency `xml:"dependency"`
	Actions     PluginActions    `xml:"actions"`
}

func (p *Plugin) GetRepoDir() string {
	return filepath.Join(p.Repo.Directory, p.General.Name)
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

func SearchPlugin(name string) (*Plugin, error) {
	pl := GetPlugin(name)
	if pl != nil {
		return pl, nil
	}
	for _, v := range Repositories {
		if v.Type != RepositoryTypeDirectory {
			continue
		}
		pl, err := LoadPlugin(v, name)
		if err == nil {
			return pl, nil
		}
	}
	return nil, ErrPluginNotFound
}

func GetPlugin(name string) *Plugin {
	for _, v := range Plugins {
		if v.Name == name {
			return v
		}
	}
	return nil
}

func LoadPlugin(repo *Repository, name string) (*Plugin, error) {
	path := filepath.Join(repo.Directory, name)
	_, err := os.Stat(path)
	if err != nil {
		return nil, ErrPluginNotFound
	}
	file, err := os.ReadFile(filepath.Join(path, "plugin.xml"))
	if err != nil {
		return nil, err
	}
	plugin := Plugin{}
	err = xml.Unmarshal(file, &plugin)
	if err != nil {
		return nil, err
	}
	if plugin.General.Name != name {
		return nil, ErrNameNotEqual
	}
	plugin.Package = &Package{
		Name:      name,
		Type:      PackageTypeInternal,
		Repo:      repo,
		Plugin:    &plugin,
		Installed: false,
	}
	Plugins = append(Plugins, &plugin)
	return &plugin, nil
}
