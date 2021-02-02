package actions

import (
	"fmt"
	"github.com/springernature/halfpipe/manifest"
	"path"
)

func (a *Actions) dockerPushSteps(task manifest.DockerPush, man manifest.Manifest) (steps Steps) {
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
			{"outputs", "type=image,oci-mediatypes=true,push=true"},
		},
		Env: Env(task.Vars),
	})

	steps = append(steps, repositoryDispatch(man.PipelineName()))
	return steps
}

func tags(task manifest.DockerPush) string {
	tagVar := "${{ env.BUILD_VERSION }}"
	if task.Tag == "gitref" {
		tagVar = "${{ env.GIT_REVISION }}"
	}
	return fmt.Sprintf("%s:latest\n%s:%s\n", task.Image, task.Image, tagVar)
}

func repositoryDispatch(name string) Step {
	return Step{
		Name: "Repository dispatch",
		Uses: "peter-evans/repository-dispatch@v1",
		With: With{
			{"token", repoAccessToken},
			{"event-type", "docker-push:" + name},
		},
	}

}
