package actions

import (
	"github.com/gosimple/slug"
	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/renderers/shared"
)

func (a *Actions) copyContainerImageSteps(task manifest.CopyContainerImage) (steps Steps) {

	step := Step{
		Name: task.GetName(),
		ID:   slug.Make(task.GetName()),
		Run:  shared.CopyContainerImageScript,
		Env: Env{
			"SOURCE_URL":            task.Source,
			"TARGET_URL":            task.Target,
			"AWS_ACCESS_KEY_ID":     task.AwsAccessKeyID,
			"AWS_SECRET_ACCESS_KEY": task.AwsSecretAccessKey,
		},
	}

	return Steps{step}
}
