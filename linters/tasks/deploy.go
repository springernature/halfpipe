package tasks

import (
	"fmt"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/cf"
	"github.com/springernature/halfpipe/defaults"
	"github.com/springernature/halfpipe/linters/filechecker"
	"github.com/springernature/halfpipe/linters/linterrors"
	"github.com/springernature/halfpipe/manifest"
	"strings"
)

func LintDeployCFTask(cf manifest.DeployCF, man manifest.Manifest, readCfManifest cf.ManifestReader, fs afero.Afero, deprecatedApis []string) (errs []error, warnings []error) {
	if cf.API == "" {
		errs = append(errs, linterrors.NewMissingField("api"))
	}
	if cf.Space == "" {
		errs = append(errs, linterrors.NewMissingField("space"))
	}
	if cf.Org == "" {
		errs = append(errs, linterrors.NewMissingField("org"))
	}
	if cf.TestDomain == "" {
		_, found := defaults.DefaultValues.CF.TestDomains[cf.API]
		if cf.API != "" && !found {
			errs = append(errs, linterrors.NewMissingField("testDomain"))
		}
	}

	if cf.Retries < 0 || cf.Retries > 5 {
		errs = append(errs, linterrors.NewInvalidField("retries", "must be between 0 and 5"))
	}

	if strings.HasPrefix(cf.Manifest, "../artifacts/") {
		warnings = append(warnings, linterrors.NewFileError(cf.Manifest, "this file must be saved as an artifact in a previous task"))
		if len(cf.PrePromote) > 0 {
			errs = append(errs, linterrors.NewInvalidField("pre_promote", "if you are using generated manifest you cannot have pre promote tasks"))

		}
	} else {

		if err := filechecker.CheckFile(fs, cf.Manifest, false); err != nil {
			errs = append(errs, err)
		}
	}

	if cf.Rolling && len(cf.PreStart) > 0 {
		errs = append(errs, linterrors.NewInvalidField("pre_start", "cannot use pre_start with rolling deployment"))
	} else {
		for _, preStartCommand := range cf.PreStart {
			if !strings.HasPrefix(preStartCommand, "cf ") {
				errs = append(errs, linterrors.NewInvalidField("pre_start", fmt.Sprintf("only cf commands are allowed: %s", preStartCommand)))
			}
		}
	}

	for i, prePromoteTask := range cf.PrePromote {
		if prePromoteTask.GetNotifications().NotificationsDefined() {
			errs = append(errs, linterrors.NewInvalidField(
				fmt.Sprintf("pre_promote[%d].notifications", i), "pre_promote tasks are not allowed to specify notifications. Please move them up to the 'deploy-cf' task"))
		}
	}

	for _, api := range deprecatedApis {
		if cf.API == api {
			warnings = append(warnings, linterrors.NewDeprecatedCFApiError(api))
			if cf.Rolling {
				errs = append(errs, linterrors.NewInvalidField("rolling", "cannot use rolling deployment with a deprecated api"))
			}
		}
	}

	if cf.DockerTag != "" {

		apps, err := readCfManifest(cf.Manifest, nil, nil)
		if err != nil {
			errs = append(errs, err)
			return
		}

		if apps[0].DockerImage == "" {
			errs = append(errs, linterrors.NewInvalidField("docker_tag", "you must specify a 'docker.image' in the CF manifest if you want to use this feature"))
			return
		}

		if (cf.DockerTag != "gitref") && (cf.DockerTag != "version") {
			errs = append(errs, linterrors.NewInvalidField("docker_tag", "must be either 'gitref' or 'version'"))
		}

		if cf.DockerTag == "version" && !man.FeatureToggles.Versioned() {
			errs = append(errs, linterrors.NewInvalidField("docker_tag", "'version' requires the update-pipeline feature toggle"))
		}
	}

	if cf.CliVersion != "cf6" && cf.CliVersion != "cf7" {
		errs = append(errs, linterrors.NewInvalidField("cli_version", "must be either 'cf6' or 'cf7'"))
	}

	return errs, warnings
}
