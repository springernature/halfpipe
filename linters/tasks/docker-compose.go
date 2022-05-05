package tasks

import (
	"fmt"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/linters/filechecker"
	"github.com/springernature/halfpipe/linters/linterrors"
	"github.com/springernature/halfpipe/manifest"
	"gopkg.in/yaml.v2"
	"strings"
)

func LintDockerComposeTask(dc manifest.DockerCompose, fs afero.Afero) (errs []error, warnings []error) {
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

	e, w := lintDockerComposeService(dc.Service, dc.ComposeFile, composeContent)
	errs = append(errs, e...)
	warnings = append(warnings, w...)

	return errs, warnings
}

func lintDockerComposeService(service string, composeFile string, composeContent []byte) (errs []error, warnings []error) {
	var compose struct {
		Version  string
		Services map[string]interface{}
	}
	err := yaml.Unmarshal(composeContent, &compose)
	if err != nil {
		err = linterrors.NewFileError(composeFile, err.Error())
		errs = append(errs, err)
		return errs, warnings
	}

	if compose.Services == nil || strings.HasPrefix(compose.Version, "1") {
		err = linterrors.DeprecatedDockerComposeVersionError{}
		warnings = append(warnings, err)
		return errs, warnings
	}

	if _, ok := compose.Services[service]; !ok {
		errs = append(errs, linterrors.NewInvalidField("service", fmt.Sprintf("could not find service '%s' in %s", service, composeFile)))
		return errs, warnings
	}

	return
}
