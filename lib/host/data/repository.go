package data

import "github.com/PCCSuite/PCCPluginSys/lib/host/config"

type RepositoryType string

const (
	RepositoryTypeDirectory RepositoryType = "directory"
	RepositoryTypeExternal  RepositoryType = "external"
)

var Repositories map[string]*Repository = map[string]*Repository{}

type Repository struct {
	Name string
	Type RepositoryType

	// if only external
	Source *Plugin

	// if only directory
	Directory string
}

func InitInternalRepositories() {
	for k, v := range config.Config.Repositories {
		Repositories[k] = &Repository{
			Name:      k,
			Type:      RepositoryTypeDirectory,
			Directory: v,
		}
	}
}

func NewExternalRepository(source *Plugin) *Repository {
	repo := Repository{
		Name:   source.Name,
		Type:   RepositoryTypeExternal,
		Source: source,
	}
	Repositories[source.Name] = &repo
	return &repo
}
