package tasks

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/linters/filechecker"
	"github.com/springernature/halfpipe/linters/linterrors"
	"github.com/springernature/halfpipe/manifest"
)

func LintDockerPushTask(docker manifest.DockerPush, manifest manifest.Manifest, fs afero.Afero) (errs []error, warnings []error) {
	if docker.Image == "" {
		errs = append(errs, linterrors.NewMissingField("image"))
	} else {
		matched, _ := regexp.Match(`^(.*)/(.*)$`, []byte(docker.Image))
		if !matched {
			errs = append(errs, linterrors.NewInvalidField("image", "must be specified as 'user/image' or 'registry/user/image'"))
		} else {
			// validate the team in repo directory only for halfpipe-io registry
			// Taken from dockerLogin(task.Image, task.Username, task.Password)
			// set registry if not docker hub by counting slashes
			// docker hub format: repository:tag or user/repository:tag
			// other registries:  another.registry/user/repository:tag
			if strings.Count(docker.Image, "/") < 3 && strings.HasPrefix(docker.Image, "eu.gcr.io/halfpipe-io/") {
				warnings = append(warnings, linterrors.NewInvalidField("image", "recommended to be specified as 'eu.gcr.io/halfpip-io/<team>/<imageName>'"))
			}
		}
	}

	if docker.Retries < 0 || docker.Retries > 5 {
		errs = append(errs, linterrors.NewInvalidField("retries", "must be between 0 and 5"))
	}

	severities := map[string]bool{
		"CRITICAL": true,
		"HIGH":     true,
		"MEDIUM":   true,
		"LOW":      true,
		"SKIP":     true,
	}
	if docker.ImageScanSeverity != "" {
		if !severities[strings.ToUpper(docker.ImageScanSeverity)] {
			errs = append(
				errs,
				linterrors.NewInvalidField("image_scan_severity",
					"Unknown image_scan_severity, please use CRITICAL, HIGH, MEDIUM, LOW or SKIP to skip the scan step"))
		}
	}

	if docker.DockerfilePath == "" {
		errs = append(errs, linterrors.NewInvalidField("dockerfile_path", "must not be empty"))
	}

	_, err := filechecker.ReadFile(fs, docker.DockerfilePath)
	if err != nil {
		errs = append(errs, err)
	}

	if docker.BuildPath != "" {
		isDir, err := fs.IsDir(docker.BuildPath)
		if err != nil {
			if os.IsNotExist(err) {
				errs = append(errs, linterrors.NewInvalidField("build_path", fmt.Sprintf("'%s' does not exist", docker.BuildPath)))
			} else {
				errs = append(errs, err)
			}
		} else if !isDir {
			errs = append(errs, linterrors.NewInvalidField("build_path", fmt.Sprintf("'%s' must be a directory but is a file ", docker.BuildPath)))
		}
	}

	if docker.Tag != "" {
		warnings = append(warnings, linterrors.NewDeprecatedField("tag", "this field is not needed anymore"))
	}

	return errs, warnings
}
