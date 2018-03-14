package linters

import (
	"fmt"

	"github.com/springernature/halfpipe/linters/errors"
	"github.com/springernature/halfpipe/manifest"
)

type artifactsLinter struct {
}

func NewArtifactsLinter() artifactsLinter {
	return artifactsLinter{}
}

func (linter artifactsLinter) Lint(man manifest.Manifest) (result LintResult) {
	result.Linter = "Artifacts"

	var artifacts int
	var artifact string
	for _, t := range man.Tasks {
		switch task := t.(type) {
		case manifest.Run:
			if len(task.SaveArtifacts) > 0 {
				artifacts++
				artifact = task.SaveArtifacts[0]
				if artifacts > 1 {
					result.AddError(errors.NewInvalidField("run.save_artifact", "found multiple 'save_artifact', currently halfpipe only supports saving artifacts from on task"))
					return
				}
				if len(task.SaveArtifacts) > 1 {
					result.AddError(errors.NewInvalidField("run.save_artifact", "found multiple artifacts in 'save_artifact', currently halfpipe only supports saving one artifacts"))
					return
				}
			}

		case manifest.DeployCF:
			if task.DeployArtifact != "" && artifact == "" {
				var errorStr string
				errorStr = fmt.Sprintf("No previous tasks have saved a artifact")
				result.AddError(errors.NewInvalidField("deploy-cf.deploy_artifact", errorStr))
			}
		}
	}

	return
}
