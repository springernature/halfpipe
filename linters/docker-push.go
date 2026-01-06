package linters

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/manifest"
)

func LintDockerPushTask(docker manifest.DockerPush, man manifest.Manifest, fs afero.Afero) (errs []error) {
	if docker.Image == "" {
		errs = append(errs, NewErrMissingField("image"))
	}

	// check team is in path if deploying to halfpipe GCR
	if strings.HasPrefix(docker.Image, "eu.gcr.io/halfpipe-io/") && !strings.HasPrefix(docker.Image, fmt.Sprintf("eu.gcr.io/halfpipe-io/%s/", man.Team)) {
		errs = append(errs, ErrDockerRegistry.WithValue(docker.Image).AsWarning())
	}

	if man.Platform.IsActions() {
		// we allow pushing to halfpipe GCR or ECR in github actions
		if !strings.HasPrefix(docker.Image, "eu.gcr.io/halfpipe-io/") && !docker.IsECR() {
			errs = append(errs, ErrDockerRegistry.WithValue(docker.Image))
		}
	}

	// Validate ECR configuration
	if docker.IsECR() {
		errs = append(errs, ErrECRExperimental.AsWarning())
	}

	if docker.Retries < 0 || docker.Retries > 5 {
		errs = append(errs, NewErrInvalidField("retries", "must be between 0 and 5"))
	}

	if docker.DockerfilePath == "" {
		errs = append(errs, NewErrInvalidField("dockerfile_path", "must not be empty"))
	}

	if !docker.RestoreArtifacts {
		_, err := ReadFile(fs, docker.DockerfilePath)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if docker.BuildPath != "" && !docker.RestoreArtifacts {
		isDir, err := fs.IsDir(docker.BuildPath)
		if err != nil {
			if os.IsNotExist(err) {
				errs = append(errs, NewErrInvalidField("build_path", fmt.Sprintf("'%s' does not exist", docker.BuildPath)))
			} else {
				errs = append(errs, err)
			}
		} else if !isDir {
			errs = append(errs, NewErrInvalidField("build_path", fmt.Sprintf("'%s' must be a directory but is a file ", docker.BuildPath)))
		}
	}

	if docker.Tag != "" {
		errs = append(errs, ErrDockerPushTag.AsWarning())
	}

	for _, platform := range docker.Platforms {
		if !slices.Contains([]string{"linux/amd64", "linux/arm64"}, platform) {
			errs = append(errs, ErrDockerPlatformUnknown)
		}
	}

	for k, v := range docker.Vars {
		if strings.HasPrefix(v, "((") && strings.HasSuffix(v, "))") && !strings.HasPrefix(k, "ARTIFACTORY_") {
			errs = append(errs, ErrDockerVarSecret.WithValue(k).AsWarning())
		}
	}

	return errs
}
