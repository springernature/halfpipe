package tasks

import (
	"fmt"
	"os"
	"regexp"

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
		}
	}

	if docker.Retries < 0 || docker.Retries > 5 {
		errs = append(errs, linterrors.NewInvalidField("retries", "must be between 0 and 5"))
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
		if (docker.Tag != "gitref") && (docker.Tag != "version") {
			errs = append(errs, linterrors.NewInvalidField("tag", "must be either 'gitref' or 'version'"))
		}
	}

	return errs, warnings
}
