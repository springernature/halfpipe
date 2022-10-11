package linters

import (
	"fmt"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/manifest"
	"gopkg.in/yaml.v2"
	"strings"
)

func LintDockerComposeTask(dc manifest.DockerCompose, fs afero.Afero) (errs []error) {
	if dc.Retries < 0 || dc.Retries > 5 {
		errs = append(errs, NewErrInvalidField("retries", "must be between 0 and 5"))
	}

	if err := CheckFile(fs, dc.ComposeFile, false); err != nil {
		errs = append(errs, err)
		return errs
	}

	composeContent, err := fs.ReadFile(dc.ComposeFile)
	if err != nil {
		errs = append(errs, err)
		return errs
	}

	e := lintDockerComposeService(dc.Service, dc.ComposeFile, composeContent)
	errs = append(errs, e...)

	return errs
}

func lintDockerComposeService(service string, composeFile string, composeContent []byte) (errs []error) {
	var compose struct {
		Version  string
		Services map[string]interface{}
	}
	err := yaml.Unmarshal(composeContent, &compose)
	if err != nil {
		err = ErrFileInvalid.WithValue(err.Error())
		errs = append(errs, err)
		return errs
	}

	if compose.Services == nil || strings.HasPrefix(compose.Version, "1") {
		errs = append(errs, ErrDockerComposeVersion.WithFile(composeFile).AsWarning())
		return errs
	}

	if _, ok := compose.Services[service]; !ok {
		errs = append(errs, NewErrInvalidField("service", fmt.Sprintf("could not find service '%s' in %s", service, composeFile)))
		return errs
	}

	return
}
