package linters

import (
	"github.com/springernature/halfpipe/manifest"
)

func LintDeployMLZipTask(mlTask manifest.DeployMLZip) (errs []error, warnings []error) {
	if len(mlTask.Targets) == 0 {
		errs = append(errs, NewErrMissingField("target"))
	}

	if mlTask.DeployZip == "" {
		errs = append(errs, NewErrMissingField("deploy_zip"))
	}

	if mlTask.Retries < 0 || mlTask.Retries > 5 {
		errs = append(errs, NewErrInvalidField("retries", "must be between 0 and 5"))
	}

	if mlTask.AppVersion != "" && mlTask.UseBuildVersion {
		errs = append(errs, NewErrInvalidField("use_build_version", "cannot set both app_version and use_build_version"))
	}
	return errs, warnings
}

func LintDeployMLModulesTask(mlTask manifest.DeployMLModules) (errs []error, warnings []error) {
	if len(mlTask.Targets) == 0 {
		errs = append(errs, NewErrMissingField("target"))
	}

	if mlTask.MLModulesVersion == "" {
		errs = append(errs, NewErrMissingField("ml_modules_version"))
	}

	if mlTask.Retries < 0 || mlTask.Retries > 5 {
		errs = append(errs, NewErrInvalidField("retries", "must be between 0 and 5"))
	}

	if mlTask.AppVersion != "" && mlTask.UseBuildVersion {
		errs = append(errs, NewErrInvalidField("use_build_version", "cannot set both app_version and use_build_version"))
	}
	return errs, warnings
}
