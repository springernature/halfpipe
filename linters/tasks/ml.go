package tasks

import (
	"github.com/springernature/halfpipe/linters/linterrors"
	"github.com/springernature/halfpipe/manifest"
)

func LintDeployMLZipTask(mlTask manifest.DeployMLZip) (errs []error, warnings []error) {
	if len(mlTask.Targets) == 0 {
		errs = append(errs, linterrors.NewMissingField("target"))
	}

	if mlTask.DeployZip == "" {
		errs = append(errs, linterrors.NewMissingField("deploy_zip"))
	}

	if mlTask.Retries < 0 || mlTask.Retries > 5 {
		errs = append(errs, linterrors.NewInvalidField("retries", "must be between 0 and 5"))
	}

	if mlTask.AppVersion != "" && mlTask.UseBuildVersion {
		errs = append(errs, linterrors.NewInvalidField("use_build_version", "cannot set both app_version and use_build_version"))
	}
	return errs, warnings
}

func LintDeployMLModulesTask(mlTask manifest.DeployMLModules) (errs []error, warnings []error) {
	if len(mlTask.Targets) == 0 {
		errs = append(errs, linterrors.NewMissingField("target"))
	}

	if mlTask.MLModulesVersion == "" {
		errs = append(errs, linterrors.NewMissingField("ml_modules_version"))
	}

	if mlTask.Retries < 0 || mlTask.Retries > 5 {
		errs = append(errs, linterrors.NewInvalidField("retries", "must be between 0 and 5"))
	}

	if mlTask.AppVersion != "" && mlTask.UseBuildVersion {
		errs = append(errs, linterrors.NewInvalidField("use_build_version", "cannot set both app_version and use_build_version"))
	}
	return errs, warnings
}
