package actions

import (
	"github.com/gosimple/slug"
	"github.com/springernature/halfpipe/manifest"
	"path"
)

func (a *Actions) buildpackSteps(task manifest.Buildpack) (steps Steps) {

	appPath := a.workingDir
	if len(task.Path) > 0 {
		appPath = path.Join(appPath, task.Path)
	}

	step := Step{
		Name: task.GetName(),
		ID:   slug.Make(task.GetName()),
		Env:  Env(task.Vars),
		Uses: "springernature/ee-action-buildpack@v1",
		With: With{
			"builder":    task.Builder,
			"buildpacks": task.Buildpacks,
			"image":      task.Image,
			"path":       appPath,
			"tags":       "${{ env.BUILD_VERSION }},${{ env.GIT_REVISION }}",
			"buildEnv":   MultiLine{task.Vars},
		},
	}

	return Steps{step}
}
