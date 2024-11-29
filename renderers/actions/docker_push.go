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
	}
	for k, v := range task.Vars {
		buildArgs[k] = v
	}

	push := Step{
		Name: "Build and Push",
		Uses: "springernature/ee-action-docker-push@v1",
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

	return Steps{push, repositoryDispatch(task.Image)}
}

func repositoryDispatch(name string) Step {
	return Step{
		Name: "Repository dispatch",
		Uses: "peter-evans/repository-dispatch@ff45666b9427631e3450c54a1bcbee4d9ff4d7c0", // v3
		With: With{
			"token":      githubSecrets.RepositoryDispatchToken,
			"event-type": "docker-push:" + name,
		},
	}
}
