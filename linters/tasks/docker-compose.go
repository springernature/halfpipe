package tasks

import (
	"fmt"

	"github.com/simonjohansson/yaml"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/linters/filechecker"
	"github.com/springernature/halfpipe/linters/linterrors"
	"github.com/springernature/halfpipe/manifest"
)

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
		Services map[string]interface{}
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
