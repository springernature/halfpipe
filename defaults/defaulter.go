package defaults

import (
	"strings"

	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/project"
)

type Defaulter func(manifest.Manifest, project.Data) manifest.Manifest

type Defaults struct {
	RepoPrivateKey       string
	CfUsername           string
	CfPassword           string
	CfManifest           string
	CfTestDomains        map[string]string
	DockerUsername       string
	DockerPassword       string
	Project              project.Data
	DockerComposeService string
}

func NewDefaulter(project project.Data) Defaults {
	d := DefaultValues
	d.Project = project
	return d
}

func (d Defaults) Update(man manifest.Manifest) manifest.Manifest {
	if man.Repo.URI == "" {
		man.Repo.URI = d.Project.GitURI
		man.Repo.BasePath = d.Project.BasePath
	}

	if man.Repo.URI != "" && !man.Repo.IsPublic() && man.Repo.PrivateKey == "" {
		man.Repo.PrivateKey = d.RepoPrivateKey
	}

	var taskSwitcher func(task manifest.TaskList) manifest.TaskList

	taskSwitcher = func(task manifest.TaskList) (tl manifest.TaskList) {
		tl = make(manifest.TaskList, len(task))
		for i, t := range task {
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
				if task.PrePromote != nil {
					task.PrePromote = taskSwitcher(task.PrePromote)
				}
				if task.TestDomain == "" {
					if domain, ok := d.CfTestDomains[task.API]; ok {
						task.TestDomain = domain
					}
				}

				tl[i] = task

			case manifest.Run:
				if strings.HasPrefix(task.Docker.Image, config.DockerRegistry) {
					task.Docker.Username = d.DockerUsername
					task.Docker.Password = d.DockerPassword
				}
				tl[i] = task

			case manifest.DockerPush:
				if strings.HasPrefix(task.Image, config.DockerRegistry) {
					task.Username = d.DockerUsername
					task.Password = d.DockerPassword
				}
				tl[i] = task

			case manifest.DockerCompose:
				if task.Service == "" {
					task.Service = d.DockerComposeService
				}
				tl[i] = task

			default:
				tl[i] = task
			}
		}
		return
	}

	taskList := taskSwitcher(man.Tasks)
	man.Tasks = taskList

	return man

}
