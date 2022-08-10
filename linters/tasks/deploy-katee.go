package tasks

import (
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/linters/filechecker"
	"github.com/springernature/halfpipe/linters/linterrors"
	"github.com/springernature/halfpipe/manifest"
)

func LintDeployKateeTask(task manifest.DeployKatee, man manifest.Manifest, fs afero.Afero) (errs []error, warnings []error) {
	if task.ApplicationName == "" {
		errs = append(errs, linterrors.NewMissingField("application_name"))
	}

	if task.Image == "" {
		errs = append(errs, linterrors.NewMissingField("image"))
	}

	if task.Retries < 0 || task.Retries > 5 {
		errs = append(errs, linterrors.NewInvalidField("retries", "must be between 0 and 5"))
	}

	if err := filechecker.CheckFile(fs, task.VelaManifest, false); err != nil {
		errs = append(errs, err)
	}

	if task.Tag != "" {
		if task.Tag != "version" && task.Tag != "gitref" {
			errs = append(errs, linterrors.NewInvalidField("tag", "must be either 'version' or 'gitref'"))
		}
	}

	if task.Tag == "version" && man.Platform.IsConcourse() && !man.FeatureToggles.UpdatePipeline() {
		errs = append(errs, linterrors.NewInvalidField("tag", "'version' requires the 'update-pipeline' feature toggle"))
	}

	return errs, warnings
}
