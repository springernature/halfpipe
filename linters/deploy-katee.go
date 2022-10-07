package linters

import (
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/manifest"
)

func LintDeployKateeTask(task manifest.DeployKatee, man manifest.Manifest, fs afero.Afero) (errs []error, warnings []error) {
	if task.ApplicationName == "" {
		errs = append(errs, NewErrMissingField("application_name"))
	}

	if task.Image == "" {
		errs = append(errs, NewErrMissingField("image"))
	}

	if task.Retries < 0 || task.Retries > 5 {
		errs = append(errs, NewErrInvalidField("retries", "must be between 0 and 5"))
	}

	if err := CheckFile(fs, task.VelaManifest, false); err != nil {
		errs = append(errs, err)
	}

	if task.Tag != "" {
		if task.Tag != "version" && task.Tag != "gitref" {
			errs = append(errs, NewErrInvalidField("tag", "must be either 'version' or 'gitref'"))
		}
	}

	if task.Tag == "version" && man.Platform.IsConcourse() && !man.FeatureToggles.UpdatePipeline() {
		errs = append(errs, NewErrInvalidField("tag", "'version' requires the 'update-pipeline' feature toggle"))
	}

	return errs, warnings
}
