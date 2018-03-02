package defaults

import (
	"strings"

	"github.com/springernature/halfpipe/manifest"
)

type Defaulter func(manifest.Manifest, Project) manifest.Manifest

type Defaults struct {
	RepoPrivateKey string
	CfUsername     string
	CfPassword     string
	CfManifest     string
	DockerUsername string
	DockerPassword string
	Project        Project
}

func NewDefaulter(project Project) Defaults {
	d := DefaultValues
	d.Project = project
	return d
}

func (d Defaults) Update(man manifest.Manifest) manifest.Manifest {
	man.Repo.Uri = d.Project.GitUri
	man.Repo.BasePath = d.Project.BasePath

	if man.Repo.Uri != "" && !man.Repo.IsPublic() && man.Repo.PrivateKey == "" {
		man.Repo.PrivateKey = d.RepoPrivateKey
	}

	for i, t := range man.Tasks {
		switch task := t.(type) {
		case manifest.DeployCF:
			if task.Org == "" {
				task.Org = man.Team
			}
			if task.Username == "" {
				task.Username = d.CfUsername
			}
			if task.Password == "" {
				task.Password = d.CfPassword
			}
			if task.Manifest == "" {
				task.Manifest = d.CfManifest
			}
			man.Tasks[i] = task
		case manifest.Run:
			if strings.HasPrefix(task.Docker.Image, "eu.gcr.io/halfpipe-io/") {
				task.Docker.Username = d.DockerUsername
				task.Docker.Password = d.DockerPassword
			}
			man.Tasks[i] = task
		}
	}

	return man
}
