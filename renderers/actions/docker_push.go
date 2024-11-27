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

	step := Step{
		Name: "Build and Push",
		Uses: "springernature/ee-action-docker-push@v1",
		With: With{
			"image":      task.Image,
			"tags":       strings.Join([]string{"latest", "${{ env.BUILD_VERSION }}", "${{ env.GIT_REVISION }}"}, "\n"),
			"context":    path.Join(a.workingDir, task.BuildPath),
			"dockerfile": path.Join(a.workingDir, task.DockerfilePath),
			"buildArgs":  MultiLine{buildArgs},
			"secrets":    MultiLine{task.Secrets},
			"platforms":  strings.Join(task.Platforms, ","),
		},
	}

	return Steps{step}
}
