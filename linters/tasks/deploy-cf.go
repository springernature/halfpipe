package tasks

import (
	"fmt"
	"strings"

	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/cf"
	"github.com/springernature/halfpipe/defaults"
	"github.com/springernature/halfpipe/linters/filechecker"
	"github.com/springernature/halfpipe/linters/linterrors"
	"github.com/springernature/halfpipe/manifest"
)

func LintDeployCFTask(cf manifest.DeployCF, man manifest.Manifest, readCfManifest cf.ManifestReader, fs afero.Afero) (errs []error, warnings []error) {
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
		_, found := defaults.Concourse.CF.TestDomains[cf.API]
		if cf.API != "" && !found {
			errs = append(errs, linterrors.NewMissingField("test_domain"))
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

	if cf.DockerTag != "" {

		cfManifest, err := readCfManifest(cf.Manifest, nil, nil)
		if err != nil {
			errs = append(errs, err)
			return
		}

		if cfManifest.Applications[0].Docker == nil || cfManifest.Applications[0].Docker.Image == "" {
			errs = append(errs, linterrors.NewInvalidField("docker_tag", "you must specify 'docker.image' in the CF manifest if you want to use this feature"))
			return
		}

		if (cf.DockerTag != "gitref") && (cf.DockerTag != "version") {
			errs = append(errs, linterrors.NewInvalidField("docker_tag", "must be either 'gitref' or 'version'"))
		}

	}

	if cf.CliVersion != "cf6" && cf.CliVersion != "cf7" && cf.CliVersion != "cf8" {
		errs = append(errs, linterrors.NewInvalidField("cli_version", "must be either 'cf6', 'cf7' or 'cf8'"))
	}

	return errs, warnings
}
