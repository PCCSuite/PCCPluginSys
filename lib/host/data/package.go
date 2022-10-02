package data

import "log"

type PackageType string

const (
	PackageTypeInternal PackageType = "internal"
	PackageTypeExternal PackageType = "external"
)

var ExternalPackages []*Package = make([]*Package, 0)

// includes plugin and external package
type Package struct {
	// package name. not include source repository name.
	Name string

	Type PackageType

	Repo *Repository

	// if type is internal, that plugin.
	// if type is external, source plugin.
	Plugin *Plugin

	Installed bool

	RunningAction *RunningAction
}

func GetExternalPackage(repoName string, name string) *Package {
	for _, v := range ExternalPackages {
		if v.Repo.Name == repoName && v.Name == name {
			return v
		}
	}
	return nil
}

func NewExternalPackage(name string, repo *Repository) *Package {
	if repo.Type != RepositoryTypeExternal {
		log.Panicf("Repository is not external. packageName: %s, repoName: %s", name, repo.Name)
	}
	pack := Package{
		Name:      name,
		Type:      PackageTypeExternal,
		Repo:      repo,
		Plugin:    repo.Source,
		Installed: false,
	}
	ExternalPackages = append(ExternalPackages, &pack)
	return &pack
}
