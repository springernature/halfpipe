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
	result.DocsURL = "https://docs.halfpipe.io/docs/artifacts/"

	var thereIsAtLeastOneArtifact bool
	for _, t := range man.Tasks {
		switch task := t.(type) {
		case manifest.Run:
			if len(task.SaveArtifacts) > 0 {
				thereIsAtLeastOneArtifact = true
			}
		case manifest.DeployCF:
			if task.DeployArtifact != "" && !thereIsAtLeastOneArtifact {
				msg := fmt.Sprintf("No previous tasks have saved a artifact")
				result.AddError(errors.NewInvalidField("deploy-cf.deploy_artifact", msg))
			}
		}
	}

	return
}
