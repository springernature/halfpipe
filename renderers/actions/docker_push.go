package actions

import (
	"path"
	"strings"

	"github.com/springernature/halfpipe/manifest"
)

func (a *Actions) dockerPushSteps(task manifest.DockerPush, man manifest.Manifest) Steps {
	buildArgs := map[string]string{
		"ARTIFACTORY_PASSWORD": "",
		"ARTIFACTORY_URL":      "",
		"ARTIFACTORY_USERNAME": "",
		"BUILD_VERSION":        "",
		"GIT_REVISION":         "",
		"RUNNING_IN_CI":        "",
		"CI":                   "",
	}
	for k, v := range task.Vars {
		buildArgs[k] = v
	}

	push := Step{
		Name: "Build and Push",
		Uses: ExternalActions.DockerPush,
		With: With{
			"image":      task.Image,
			"tags":       "latest\n${{ env.BUILD_VERSION }}\n${{ env.GIT_REVISION }}\n",
			"context":    path.Join(a.workingDir, task.BuildPath),
			"dockerfile": path.Join(a.workingDir, task.DockerfilePath),
			"buildArgs":  MultiLine{buildArgs},
			"secrets":    MultiLine{task.Secrets},
			"platforms":  strings.Join(task.Platforms, ","),
		},
	}

	// useCache will be set on manual "workflow dispatch" trigger.
	// otherwise it will be an empty string and we default it to true
	if task.UseCache {
		push.With["useCache"] = "${{ inputs.useCache == '' || inputs.useCache == 'true' }}"
	}

	if man.FeatureToggles.Ghas() {
		push.With["ghas"] = "true"
		push.With["githubPat"] = "${{ secrets.GITHUB_TOKEN }}"
	}

	return Steps{push, repositoryDispatch(task.Image)}
}

func repositoryDispatch(name string) Step {
	return Step{
		Name: "Repository dispatch",
		Uses: ExternalActions.RepositoryDispatch,
		With: With{
			"token":      githubSecrets.RepositoryDispatchToken,
			"event-type": "docker-push:" + name,
		},
	}
}
