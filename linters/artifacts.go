package linters

import (
	"fmt"

	"github.com/springernature/halfpipe/linters/errors"
	"github.com/springernature/halfpipe/manifest"
)

type artifactsLinter struct{}

func NewArtifactsLinter() artifactsLinter {
	return artifactsLinter{}
}

func (linter artifactsLinter) Lint(man manifest.Manifest) (result LintResult) {
	result.Linter = "Artifacts"
	result.DocsURL = "https://docs.halfpipe.io/docs/artifacts/"

	noPreviousArtifactErr := func(field string) error {
		return errors.NewInvalidField(field, fmt.Sprintf("No previous tasks have saved a artifact"))
	}

	var thereIsAtLeastOneArtifact bool
	for _, t := range man.Tasks {
		switch task := t.(type) {
		case manifest.Run:
			if task.RestoreArtifacts && !thereIsAtLeastOneArtifact {
				result.AddError(noPreviousArtifactErr("run.restore_artifacts"))
			}

			if len(task.SaveArtifacts) > 0 {
				thereIsAtLeastOneArtifact = true
			}

		case manifest.DockerCompose:
			if task.RestoreArtifacts && !thereIsAtLeastOneArtifact {
				result.AddError(noPreviousArtifactErr("docker-compose.restore_artifacts"))
			}

			if len(task.SaveArtifacts) > 0 {
				thereIsAtLeastOneArtifact = true
			}

		case manifest.DockerPush:
			if task.RestoreArtifacts && !thereIsAtLeastOneArtifact {
				result.AddError(noPreviousArtifactErr("docker-push.restore_artifacts"))
			}

		case manifest.DeployCF:
			if task.DeployArtifact != "" && !thereIsAtLeastOneArtifact {
				result.AddError(noPreviousArtifactErr("deploy-cf.deploy_artifact"))
			}

		case manifest.DeployMLZip:
			if task.DeployZip != "" && !thereIsAtLeastOneArtifact {
				result.AddError(noPreviousArtifactErr("deploy-ml-zip.deploy_zip"))
			}

		}


	}

	return
}
