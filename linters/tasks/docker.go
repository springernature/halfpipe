package tasks

import (
	"fmt"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/linters/errors"
	"github.com/springernature/halfpipe/linters/filechecker"
	"github.com/springernature/halfpipe/manifest"
	"gopkg.in/yaml.v2"
	"regexp"
)

func LintDockerPushTask(docker manifest.DockerPush, fs afero.Afero) (errs []error, warnings []error) {
	if docker.Username == "" {
		errs = append(errs, errors.NewMissingField("username"))
	}
	if docker.Password == "" {
		errs = append(errs, errors.NewMissingField("password"))
	}
	if docker.Image == "" {
		errs = append(errs, errors.NewMissingField("image"))
	} else {
		matched, _ := regexp.Match(`^(.*)/(.*)$`, []byte(docker.Image))
		if !matched {
			errs = append(errs, errors.NewInvalidField("image", "must be specified as 'user/image' or 'registry/user/image'"))
		}
	}

	if docker.Retries < 0 || docker.Retries > 5 {
		errs = append(errs, errors.NewInvalidField("retries", "must be between 0 and 5"))
	}

	if err := filechecker.CheckFile(fs, "Dockerfile", false); err != nil {
		errs = append(errs, err)
	}

	return
}

func LintDockerComposeTask(dc manifest.DockerCompose, fs afero.Afero) (errs []error, warnings []error) {
	if dc.Retries < 0 || dc.Retries > 5 {
		errs = append(errs, errors.NewInvalidField("retries", "must be between 0 and 5"))
	}

	composeFile := "docker-compose.yml"
	if dc.ComposeFile != "" {
		composeFile = dc.ComposeFile
	}

	if err := filechecker.CheckFile(fs, composeFile, false); err != nil {
		errs = append(errs, err)
		return
	}

	e, w := lintDockerComposeService(dc.Service, composeFile, fs)
	errs = append(errs, e...)
	warnings = append(warnings, w...)
	return
}

func lintDockerComposeService(service string, composeFile string, fs afero.Afero) (errs []error, warnings []error) {
	content, err := fs.ReadFile(composeFile)
	if err != nil {
		errs = append(errs, err)
		return
	}

	var compose struct {
		Services map[string]interface{} `yaml:"services"`
	}
	err = yaml.Unmarshal(content, &compose)
	if err != nil {
		err = errors.NewFileError(composeFile, err.Error())
		errs = append(errs, err)
		return
	}

	if _, ok := compose.Services[service]; ok {
		return
	}

	var composeWithoutServices map[string]interface{}
	err = yaml.Unmarshal(content, &composeWithoutServices)
	if err != nil {
		errs = append(errs, err)
		return
	}

	if _, ok := composeWithoutServices[service]; ok {
		return
	}

	errs = append(errs, errors.NewInvalidField("service", fmt.Sprintf("Could not find service '%s' in %s", service, composeFile)))
	return
}
