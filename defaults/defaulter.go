package defaults

import (
	"strings"

	"github.com/springernature/halfpipe/parser"
)

type Defaulter func(parser.Manifest, Project) parser.Manifest

type Defaults struct {
	RepoPrivateKey string
	CfUsername     string
	CfPassword     string
	CfManifest     string
	DockerUsername string
	DockerPassword string
}

func (d Defaults) Update(man parser.Manifest, project Project) parser.Manifest {

	man.Repo.Uri = project.GitUri
	man.Repo.BasePath = project.BasePath

	if man.Repo.Uri != "" && !man.Repo.IsPublic() && man.Repo.PrivateKey == "" {
		man.Repo.PrivateKey = d.RepoPrivateKey
	}

	for i, t := range man.Tasks {
		switch task := t.(type) {
		case parser.DeployCF:
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
		case parser.Run:
			if strings.HasPrefix(task.Docker.Image, "eu.gcr.io/halfpipe-io/") {
				task.Docker.Username = d.DockerUsername
				task.Docker.Password = d.DockerPassword
			}
			man.Tasks[i] = task
		}
	}

	return man
}
