package tasks

import (
	"fmt"
	"strings"

	"github.com/simonjohansson/yaml"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/linters/filechecker"
	"github.com/springernature/halfpipe/linters/linterrors"
	"github.com/springernature/halfpipe/manifest"
)

func LintDockerComposeTask(dc manifest.DockerCompose, fs afero.Afero, deprecatedDockerRegistries []string) (errs []error, warnings []error) {
	if dc.Retries < 0 || dc.Retries > 5 {
		errs = append(errs, linterrors.NewInvalidField("retries", "must be between 0 and 5"))
	}

	if err := filechecker.CheckFile(fs, dc.ComposeFile, false); err != nil {
		errs = append(errs, err)
		return errs, warnings
	}

	composeContent, err := fs.ReadFile(dc.ComposeFile)
	if err != nil {
		errs = append(errs, err)
		return errs, warnings
	}

	e, w := lintDockerComposeService(dc.Service, dc.ComposeFile, composeContent, fs)
	errs = append(errs, e...)
	warnings = append(warnings, w...)

	//just check as a string instead of handling all the docker-compose schema variants
	composeString := string(composeContent)
	for _, hostname := range deprecatedDockerRegistries {
		if strings.Contains(composeString, hostname) {
			warnings = append(warnings, linterrors.NewDeprecatedDockerRegistryError(hostname))
		}
	}

	return errs, warnings
}

func lintDockerComposeService(service string, composeFile string, composeContent []byte, fs afero.Afero) (errs []error, warnings []error) {
	var compose struct {
		Services map[string]interface{}
	}
	err := yaml.Unmarshal(composeContent, &compose)
	if err != nil {
		err = linterrors.NewFileError(composeFile, err.Error())
		errs = append(errs, err)
		return errs, warnings
	}

	if _, ok := compose.Services[service]; ok {
		return errs, warnings
	}

	var composeWithoutServices map[string]interface{}
	err = yaml.Unmarshal(composeContent, &composeWithoutServices)
	if err != nil {
		errs = append(errs, err)
		return errs, warnings
	}

	if _, ok := composeWithoutServices[service]; ok {
		return errs, warnings
	}

	errs = append(errs, linterrors.NewInvalidField("service", fmt.Sprintf("could not find service '%s' in %s", service, composeFile)))
	return errs, warnings
}
