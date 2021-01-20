package actions

import (
	"fmt"
	"github.com/springernature/halfpipe/manifest"
	"path"
)

func (a *Actions) dockerPushJob(task manifest.DockerPush, man manifest.Manifest) Job {
	steps := []Step{checkoutCode}
	if task.ReadsFromArtifacts() {
		steps = append(steps, a.restoreArtifacts())
	}
	steps = append(steps,
		Step{
			Name: "Set up Docker Buildx",
			Uses: "docker/setup-buildx-action@v1",
		},
		dockerLogin(task.Image, task.Username, task.Password),
		Step{
			Name: "Build and push",
			Uses: "docker/build-push-action@v2",
			With: With{
				{"context", path.Join(a.workingDir, task.BuildPath)},
				{"file", path.Join(a.workingDir, task.DockerfilePath)},
				{"push", true},
				{"tags", tags(task)},
				{"outputs", "type=image,oci-mediatypes=true,push=true"},
			},
		},
		repositoryDispatch(man.PipelineName()),
	)

	return Job{
		Name:   task.GetName(),
		RunsOn: defaultRunner,
		Steps:  steps,
		Env:    Env(task.Vars),
	}
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
