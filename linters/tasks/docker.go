package tasks

import (
	"fmt"
	"github.com/simonjohansson/yaml"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/linters/filechecker"
	"github.com/springernature/halfpipe/linters/linterrors"
	"github.com/springernature/halfpipe/manifest"
	"os"
	"regexp"
)

func LintDockerPushTask(docker manifest.DockerPush, fs afero.Afero) (errs []error, warnings []error) {
	if docker.Username == "" {
		errs = append(errs, linterrors.NewMissingField("username"))
	}
	if docker.Password == "" {
		errs = append(errs, linterrors.NewMissingField("password"))
	}
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

	if err := filechecker.CheckFile(fs, docker.DockerfilePath, false); err != nil {
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

	return errs, warnings
}

func LintDockerComposeTask(dc manifest.DockerCompose, fs afero.Afero) (errs []error, warnings []error) {
	if dc.Retries < 0 || dc.Retries > 5 {
		errs = append(errs, linterrors.NewInvalidField("retries", "must be between 0 and 5"))
	}

	if err := filechecker.CheckFile(fs, dc.ComposeFile, false); err != nil {
		errs = append(errs, err)
		return errs, warnings
	}

	e, w := lintDockerComposeService(dc.Service, dc.ComposeFile, fs)
	errs = append(errs, e...)
	warnings = append(warnings, w...)

	return errs, warnings
}

func lintDockerComposeService(service string, composeFile string, fs afero.Afero) (errs []error, warnings []error) {
	content, err := fs.ReadFile(composeFile)
	if err != nil {
		errs = append(errs, err)
		return errs, warnings
	}

	var compose struct {
		Services map[string]interface{} `yaml:"services"`
	}
	err = yaml.Unmarshal(content, &compose)
	if err != nil {
		err = linterrors.NewFileError(composeFile, err.Error())
		errs = append(errs, err)
		return errs, warnings
	}

	if _, ok := compose.Services[service]; ok {
		return errs, warnings
	}

	var composeWithoutServices map[string]interface{}
	err = yaml.Unmarshal(content, &composeWithoutServices)
	if err != nil {
		errs = append(errs, err)
		return errs, warnings
	}

	if _, ok := composeWithoutServices[service]; ok {
		return errs, warnings
	}

	errs = append(errs, linterrors.NewInvalidField("service", fmt.Sprintf("Could not find service '%s' in %s", service, composeFile)))
	return errs, warnings
}
