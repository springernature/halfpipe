package linters

import (
	"fmt"
	"golang.org/x/exp/slices"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/manifest"
)

func LintDockerPushTask(docker manifest.DockerPush, fs afero.Afero) (errs []error) {
	if docker.Image == "" {
		errs = append(errs, NewErrMissingField("image"))
	} else {
		matched, _ := regexp.Match(`^(.*)/(.*)$`, []byte(docker.Image))
		if !matched {
			errs = append(errs, NewErrInvalidField("image", "must be specified as 'user/image' or 'registry/user/image'"))
		} else {
			// validate the team in repo directory only for halfpipe-io registry
			// Taken from dockerLogin(task.Image, task.Username, task.Password)
			// set registry if not docker hub by counting slashes
			// docker hub format: repository:tag or user/repository:tag
			// other registries:  another.registry/user/repository:tag
			if strings.Count(docker.Image, "/") < 3 && strings.HasPrefix(docker.Image, "eu.gcr.io/halfpipe-io/") {
				errs = append(errs, NewErrInvalidField("image", "recommended to be specified as 'eu.gcr.io/halfpipe-io/<team>/<imageName>'").AsWarning())
			}
		}
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
