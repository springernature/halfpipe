package linters

import (
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/manifest"
)

func LintDockerPushAWSTask(task manifest.DockerPushAWS, man manifest.Manifest, fs afero.Afero) (errs []error) {
	errs = append(errs, ErrDockerPushAWSExperimental)

	if !man.Platform.IsActions() {
		errs = append(errs, ErrDockerPushAWSActionsOnly)
	}

	if task.Repository == "" {
		errs = append(errs, NewErrMissingField("repository"))
	}

	if !task.RestoreArtifacts {
		if err := CheckFile(fs, task.DockerfilePath, false); err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}
