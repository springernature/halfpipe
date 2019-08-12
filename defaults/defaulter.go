package defaults

import (
	"fmt"
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
	CfUsernameSnPaas     string
	CfPasswordSnPaas     string
	CfOrgSnPaas          string
	CfAPISnPaas          string
	CfManifest           string
	CfTestDomains        map[string]string
	DockerUsername       string
	DockerPassword       string
	Project              project.Data
	DockerComposeService string
	ArtifactoryUsername  string
	ArtifactoryPassword  string
	ArtifactoryURL       string
	Timeout              string
}

func NewDefaulter(project project.Data) Defaults {
	d := DefaultValues
	d.Project = project
	return d
}

func (d Defaults) getUniqueName(name string, previousTasks manifest.TaskList, counter int) string {
	candidate := name
	if counter > 0 {
		candidate = fmt.Sprintf("%s (%v)", name, counter)
	}

	for _, previousTask := range previousTasks {
		if previousTask.GetName() == candidate {
			return d.getUniqueName(name, previousTasks, counter+1)
		}
	}

	return candidate

}

func (d Defaults) uniqueName(name string, defaultName string, previousTasks manifest.TaskList) string {
	if name == "" {
		name = defaultName
	}
	return d.getUniqueName(name, previousTasks, 0)
}

func (d Defaults) Update(man manifest.Manifest) manifest.Manifest {
	man.Repo.BasePath = d.Project.BasePath

	if man.Repo.URI == "" {
		man.Repo.URI = d.Project.GitURI
	}

	if man.Repo.URI != "" && !man.Repo.IsPublic() && man.Repo.PrivateKey == "" {
		man.Repo.PrivateKey = d.RepoPrivateKey
	}

	var taskSwitcher func(tasks manifest.TaskList) manifest.TaskList

	taskSwitcher = func(tasks manifest.TaskList) (tl manifest.TaskList) {
		tl = make(manifest.TaskList, len(tasks))
		for i, task := range tasks {
			switch task := task.(type) {
			case manifest.DeployCF:
				task.Name = d.uniqueName(task.Name, "deploy-cf", tl[:i])
				if task.API == d.CfAPISnPaas {
					if task.Org == "" {
						task.Org = d.CfOrgSnPaas
					}
					if task.Username == "" {
						task.Username = d.CfUsernameSnPaas
					}
					if task.Password == "" {
						task.Password = d.CfPasswordSnPaas
					}
				} else {
					if task.Org == "" {
						task.Org = man.Team
					}
					if task.Username == "" {
						task.Username = d.CfUsername
					}
					if task.Password == "" {
						task.Password = d.CfPassword
					}
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

				if task.GetTimeout() == "" {
					task.Timeout = d.Timeout
				}

				tl[i] = task

			case manifest.Run:
				task.Name = d.uniqueName(task.Name, fmt.Sprintf("run %s", strings.Replace(task.Script, "./", "", 1)), tl[:i])
				if strings.HasPrefix(task.Docker.Image, config.DockerRegistry) {
					task.Docker.Username = d.DockerUsername
					task.Docker.Password = d.DockerPassword
				}
				task.Vars = d.addArtifactoryCredentialsToVars(task.Vars)

				if task.GetTimeout() == "" {
					task.Timeout = d.Timeout
				}

				tl[i] = task

			case manifest.DockerPush:
				task.Name = d.uniqueName(task.Name, "docker-push", tl[:i])
				if strings.HasPrefix(task.Image, config.DockerRegistry) {
					task.Username = d.DockerUsername
					task.Password = d.DockerPassword
				}

				if task.DockerfilePath == "" {
					task.DockerfilePath = "Dockerfile"
				}

				task.Vars = d.addArtifactoryCredentialsToVars(task.Vars)

				if task.GetTimeout() == "" {
					task.Timeout = d.Timeout
				}

				tl[i] = task

			case manifest.DockerCompose:
				task.Name = d.uniqueName(task.Name, "docker-compose", tl[:i])

				if task.Service == "" {
					task.Service = d.DockerComposeService
				}

				task.Vars = d.addArtifactoryCredentialsToVars(task.Vars)

				if task.GetTimeout() == "" {
					task.Timeout = d.Timeout
				}

				tl[i] = task

			case manifest.ConsumerIntegrationTest:
				task.Name = d.uniqueName(task.Name, "consumer-integration-test", tl[:i])

				task.Vars = d.addArtifactoryCredentialsToVars(task.Vars)

				if task.GetTimeout() == "" {
					task.Timeout = d.Timeout
				}

				tl[i] = task

			case manifest.DeployMLModules:
				task.Name = d.uniqueName(task.Name, "deploy-ml-modules", tl[:i])

				if task.GetTimeout() == "" {
					task.Timeout = d.Timeout
				}

				tl[i] = task

			case manifest.DeployMLZip:
				task.Name = d.uniqueName(task.Name, "deploy-ml-zip", tl[:i])

				if task.GetTimeout() == "" {
					task.Timeout = d.Timeout
				}

				tl[i] = task
			}
		}
		return
	}

	taskList := taskSwitcher(man.Tasks)
	man.Tasks = taskList

	return man

}

func (d Defaults) addArtifactoryCredentialsToVars(vars manifest.Vars) manifest.Vars {
	if len(vars) == 0 {
		vars = make(map[string]string)
	}

	if _, found := vars["ARTIFACTORY_USERNAME"]; !found {
		vars["ARTIFACTORY_USERNAME"] = d.ArtifactoryUsername
	}

	if _, found := vars["ARTIFACTORY_PASSWORD"]; !found {
		vars["ARTIFACTORY_PASSWORD"] = d.ArtifactoryPassword
	}

	if _, found := vars["ARTIFACTORY_URL"]; !found {
		vars["ARTIFACTORY_URL"] = d.ArtifactoryURL
	}

	return vars
}
