package linters

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/cf"
	"github.com/springernature/halfpipe/manifest"
)

func LintDeployCFTask(task manifest.DeployCF, readCfManifest cf.ManifestReader, fs afero.Afero) (errs []error, warnings []error) {
	if task.API == "" {
		errs = append(errs, NewErrMissingField("api"))
	}
	if task.Space == "" {
		errs = append(errs, NewErrMissingField("space"))
	}
	if task.Org == "" {
		errs = append(errs, NewErrMissingField("org"))
	}
	if task.TestDomain == "" {
		errs = append(errs, NewErrMissingField("test_domain"))
	}

	if task.Retries < 0 || task.Retries > 5 {
		errs = append(errs, NewErrInvalidField("retries", "must be between 0 and 5"))
	}

	if strings.HasPrefix(task.Manifest, "../artifacts/") {
		warnings = append(warnings, ErrCFFromArtifact.WithFile(task.Manifest))
		if len(task.PrePromote) > 0 {
			errs = append(errs, ErrCFPrePromoteArtifact)
		}
	} else {

		if err := CheckFile(fs, task.Manifest, false); err != nil {
			errs = append(errs, err)
		}
	}

	if task.Rolling && len(task.PreStart) > 0 {
		errs = append(errs, NewErrInvalidField("pre_start", "cannot use pre_start with rolling deployment"))
	} else {
		for _, preStartCommand := range task.PreStart {
			if !strings.HasPrefix(preStartCommand, "cf ") {
				errs = append(errs, NewErrInvalidField("pre_start", fmt.Sprintf("only cf commands are allowed: %s", preStartCommand)))
			}
		}
	}

	for i, prePromoteTask := range task.PrePromote {
		if prePromoteTask.GetNotifications().NotificationsDefined() {
			errs = append(errs, NewErrInvalidField(
				fmt.Sprintf("pre_promote[%d].notifications", i), "pre_promote tasks are not allowed to specify notifications. Please move them up to the 'deploy-cf' task"))
		}
	}

	if task.DockerTag != "" {

		cfManifest, err := readCfManifest(task.Manifest, nil, nil)
		if err != nil {
			errs = append(errs, err)
			return
		}

		if cfManifest.Applications[0].Docker == nil || cfManifest.Applications[0].Docker.Image == "" {
			errs = append(errs, NewErrInvalidField("docker_tag", "you must specify 'docker.image' in the CF manifest if you want to use this feature"))
			return
		}

		if (task.DockerTag != "gitref") && (task.DockerTag != "version") {
			errs = append(errs, NewErrInvalidField("docker_tag", "must be either 'gitref' or 'version'"))
		}

	}

	if task.CliVersion != "cf6" && task.CliVersion != "cf7" && task.CliVersion != "cf8" {
		errs = append(errs, NewErrInvalidField("cli_version", "must be either 'cf6', 'cf7' or 'cf8'"))
	}

	if task.SSORoute != "" {
		routePattern := regexp.MustCompile(`^[A-Za-z0-9\-]+\.public\.springernature\.app$`)
		if !routePattern.MatchString(task.SSORoute) {
			errs = append(errs, NewErrInvalidField("sso_route", "must be a sub-domain of public.springernature.app"))
		}
	}

	cfManifestErrors, cfManifestWarnings := LintCfManifest(task, readCfManifest)
	errs = append(errs, cfManifestErrors...)
	warnings = append(warnings, cfManifestWarnings...)

	return errs, warnings
}
