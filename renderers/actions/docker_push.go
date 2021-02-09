package actions

import (
	"fmt"
	"github.com/springernature/halfpipe/manifest"
	"path"
)

func (a *Actions) dockerPushSteps(task manifest.DockerPush) (steps Steps) {
	steps = append(steps, Step{
		Name: "Set up Docker Buildx",
		Uses: "docker/setup-buildx-action@v1",
	})

	steps = append(steps, dockerLogin(task.Image, task.Username, task.Password)...)

	steps = append(steps, Step{
		Name: "Build and push",
		Uses: "docker/build-push-action@v2",
		With: With{
			{"context", path.Join(a.workingDir, task.BuildPath)},
			{"file", path.Join(a.workingDir, task.DockerfilePath)},
			{"push", true},
			{"tags", tags(task)},
		},
		Env: Env(task.Vars),
	})

	steps = append(steps, repositoryDispatch(task.Image))
	return steps
}

func tags(task manifest.DockerPush) string {
	tagVar := "${{ env.BUILD_VERSION }}"
	if task.Tag == "gitref" {
		tagVar = "${{ env.GIT_REVISION }}"
	}
	return fmt.Sprintf("%s:latest\n%s:%s\n", task.Image, task.Image, tagVar)
}

func repositoryDispatch(eventName string) Step {
	return Step{
		Name: "Repository dispatch",
		Uses: "peter-evans/repository-dispatch@v1",
		With: With{
			{"token", "${{ secrets.EE_REPOSITORY_DISPATCH_TOKEN }}"},
			{"event-type", "docker-push:" + eventName},
		},
	}

}
