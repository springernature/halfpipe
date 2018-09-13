package tasks

import (
	"github.com/springernature/halfpipe/linters/errors"
	"github.com/springernature/halfpipe/manifest"
)

func LintDeployMLZipTask(mlTask manifest.DeployMLZip) (errs []error, warnings []error) {
	if len(mlTask.Targets) == 0 {
		errs = append(errs, errors.NewMissingField("target"))
	}

	if mlTask.DeployZip == "" {
		errs = append(errs, errors.NewMissingField("deploy_zip"))
	}

	if mlTask.Retries < 0 || mlTask.Retries > 5 {
		errs = append(errs, errors.NewInvalidField("retries", "must be between 0 and 5"))
	}
	return
}

func LintDeployMLModulesTask(mlTask manifest.DeployMLModules) (errs []error, warnings []error) {
	if len(mlTask.Targets) == 0 {
		errs = append(errs, errors.NewMissingField("target"))
	}

	if mlTask.MLModulesVersion == "" {
		errs = append(errs, errors.NewMissingField("ml_modules_version"))
	}

	if mlTask.Retries < 0 || mlTask.Retries > 5 {
		errs = append(errs, errors.NewInvalidField("retries", "must be between 0 and 5"))
	}
	return
}
