package linters

import (
	"strings"

	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/manifest"
)

func LintDeployKateeTask(task manifest.DeployKatee, man manifest.Manifest, fs afero.Afero) (errs []error) {
	if task.Retries < 0 || task.Retries > 5 {
		errs = append(errs, NewErrInvalidField("retries", "must be between 0 and 5"))
	}

	if task.Tag != "" {
		if task.Tag != "version" && task.Tag != "gitref" {
			errs = append(errs, NewErrInvalidField("tag", "must be either 'version' or 'gitref'"))
		}
	}

	if task.Tag == "version" && man.Platform.IsConcourse() && !man.FeatureToggles.UpdatePipeline() {
		errs = append(errs, NewErrInvalidField("tag", "'version' requires the 'update-pipeline' feature toggle"))
	}

	// Check platform_version
	if task.PlatformVersion != "" && task.PlatformVersion != "v1" && task.PlatformVersion != "v2" {
		errs = append(errs, NewErrInvalidField("platform_version", "must be '', 'v1', or 'v2'"))
	}

	// vela manifest checks
	velaAppFile, err := ReadFile(fs, task.VelaManifest)
	if err != nil {
		errs = append(errs, err)
		return errs
	}

	velaManifest, err := unMarshallVelaManifest([]byte(velaAppFile))
	if err != nil {
		errs = append(errs, ErrFileInvalid.WithFile(task.VelaManifest).WithValue(err.Error()))
		return errs
	}

	//check vela env vars are set in halfpipe task
	for _, com := range velaManifest.Spec.Components {
		for _, sec := range com.Properties.Env {
			if strings.HasPrefix(sec.Value, "${") {
				secretName := strings.ReplaceAll(sec.Value, "${", "")
				secretName = strings.ReplaceAll(secretName, "}", "")

				if _, ok := task.Vars[secretName]; !ok {
					if secretName != "BUILD_VERSION" && secretName != "GIT_REVISION" {
						errs = append(errs, ErrVelaVariableMissing.WithValue(secretName).WithFile(task.VelaManifest))
					}
				}
			}
		}
	}

	//vela namespace must start with 'katee-'
	if !strings.HasPrefix(task.Namespace, "katee-") {
		errs = append(errs, ErrVelaNamespace.WithValue(task.Namespace))
	}

	//deployment_check_timeout is deprecated
	if task.DeploymentCheckTimeout != 0 {
		errs = append(errs, ErrVelaDeploymentCheckTimeout.AsWarning())
	}

	return errs
}
