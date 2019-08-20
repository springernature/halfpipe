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

func (d Defaults) getUniqueName(name string, previousNames []string, counter int) string {
	candidate := name
	if counter > 0 {
		candidate = fmt.Sprintf("%s (%v)", name, counter)
	}

	for _, previousName := range previousNames {
		if previousName == candidate {
			return d.getUniqueName(name, previousNames, counter+1)
		}
	}

	return candidate

}

func (d Defaults) uniqueName(name string, defaultName string, previousNames []string) string {
	if name == "" {
		name = defaultName
	}
	return d.getUniqueName(name, previousNames, 0)
}

func (d Defaults) updateTasks(tasks manifest.TaskList, man manifest.Manifest) (updated manifest.TaskList) {
	var previousNames []string

	var taskSwitcher func(tasks manifest.TaskList) manifest.TaskList
	taskSwitcher = func(tasks manifest.TaskList) (tl manifest.TaskList) {
		tl = make(manifest.TaskList, len(tasks))
		for i, task := range tasks {
			switch task := task.(type) {
			case manifest.DeployCF:
				task.Name = d.uniqueName(task.Name, "deploy-cf", previousNames)
				previousNames = append(previousNames, task.Name)
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
					task.PrePromote = d.updateTasks(task.PrePromote, man)
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
				task.Name = d.uniqueName(task.Name, fmt.Sprintf("run %s", strings.Replace(task.Script, "./", "", 1)), previousNames)
				previousNames = append(previousNames, task.Name)

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
				task.Name = d.uniqueName(task.Name, "docker-push", previousNames)
				previousNames = append(previousNames, task.Name)

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
				task.Name = d.uniqueName(task.Name, "docker-compose", previousNames)
				previousNames = append(previousNames, task.Name)

				if task.Service == "" {
					task.Service = d.DockerComposeService
				}

				task.Vars = d.addArtifactoryCredentialsToVars(task.Vars)

				if task.GetTimeout() == "" {
					task.Timeout = d.Timeout
				}

				tl[i] = task

			case manifest.ConsumerIntegrationTest:
				task.Name = d.uniqueName(task.Name, "consumer-integration-test", previousNames)
				previousNames = append(previousNames, task.Name)

				task.Vars = d.addArtifactoryCredentialsToVars(task.Vars)

				if task.GetTimeout() == "" {
					task.Timeout = d.Timeout
				}

				tl[i] = task

			case manifest.DeployMLModules:
				task.Name = d.uniqueName(task.Name, "deploy-ml-modules", previousNames)
				previousNames = append(previousNames, task.Name)

				if task.GetTimeout() == "" {
					task.Timeout = d.Timeout
				}

				tl[i] = task

			case manifest.DeployMLZip:
				task.Name = d.uniqueName(task.Name, "deploy-ml-zip", previousNames)
				previousNames = append(previousNames, task.Name)

				if task.GetTimeout() == "" {
					task.Timeout = d.Timeout
				}

				tl[i] = task

			case manifest.Update:
				previousNames = append(previousNames, task.GetName())
				task.Timeout = d.Timeout
				tl[i] = task

			case manifest.Parallel:
				task.Tasks = taskSwitcher(task.Tasks)
				tl[i] = task
			}
		}
		return
	}
	updated = taskSwitcher(tasks)
	return
}

func (d Defaults) updateGitTriggerWithDefaults(man manifest.Manifest) manifest.Manifest {
	// Here the triggers.Translator repo to GitTrigger have already been run.
	// We assume that the translated trigger is the first occurance
	var gitTrigger manifest.GitTrigger
	var gitTriggerIndex int
	var found bool
	for i, trigger := range man.Triggers {
		switch trigger := trigger.(type) {
		case manifest.GitTrigger:
			found = true
			gitTriggerIndex = i
			gitTrigger = trigger
			break
		}
	}

	updatedManifest := man
	gitTrigger.BasePath = d.Project.BasePath

	if gitTrigger.URI == "" {
		gitTrigger.URI = d.Project.GitURI
	}

	if gitTrigger.URI != "" && !gitTrigger.IsPublic() && gitTrigger.PrivateKey == "" {
		gitTrigger.PrivateKey = d.RepoPrivateKey
	}

	if found {
		updatedManifest.Triggers[gitTriggerIndex] = gitTrigger
	} else {
		updatedManifest.Triggers = append(updatedManifest.Triggers, gitTrigger)
	}

	return updatedManifest
}

func (d Defaults) updateDockerTriggerWithDefaults(man manifest.Manifest) manifest.Manifest {
	// We assume that the first docker trigger we find is the right one as we lint later that we only have trigger.
	updatedManifest := man

	var dockerTrigger manifest.DockerTrigger
	var dockerTriggerIndex int
	var found bool
	for i, trigger := range man.Triggers {
		switch trigger := trigger.(type) {
		case manifest.DockerTrigger:
			found = true
			dockerTriggerIndex = i
			dockerTrigger = trigger
			break
		}
	}

	if found {
		if strings.HasPrefix(dockerTrigger.Image, config.DockerRegistry) {
			dockerTrigger.Username = d.DockerUsername
			dockerTrigger.Password = d.DockerPassword
			updatedManifest.Triggers[dockerTriggerIndex] = dockerTrigger
		}
	}

	return updatedManifest
}

func (d Defaults) updateTriggersWithDefaults(man manifest.Manifest) manifest.Manifest {
	man = d.updateGitTriggerWithDefaults(man)
	man = d.updateDockerTriggerWithDefaults(man)
	return man
}

func (d Defaults) Update(man manifest.Manifest) manifest.Manifest {
	man = d.updateTriggersWithDefaults(man)
	man = d.updateGitTriggerWithDefaults(man)

	if man.FeatureToggles.UpdatePipeline() {
		man.Tasks = append(manifest.TaskList{manifest.Update{}}, man.Tasks...)
	}

	man.Tasks = d.updateTasks(man.Tasks, man)

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
