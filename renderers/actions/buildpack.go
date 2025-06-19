package actions

import (
	"github.com/gosimple/slug"
	"github.com/springernature/halfpipe/manifest"
)

func (a *Actions) buildpackSteps(task manifest.Buildpack) (steps Steps) {
	step := Step{
		Name: task.GetName(),
		ID:   slug.Make(task.GetName()),
		Env:  Env(task.Vars),
		Uses: "springernature/ee-action-buildpack@v1",
		With: With{
			"builder":    "paketobuildpacks/builder-jammy-full",
			"buildpacks": task.Buildpacks,
			"image":      task.Image,
			"path":       task.Path,
			"tags":       "${{ env.BUILD_VERSION }}",
		},
	}

	return Steps{step}
}
