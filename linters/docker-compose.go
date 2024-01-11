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

	serviceExists := false
	for _, f := range dc.ComposeFiles {
		fServiceExists, fErr := lintComposeFile(f, dc, fs)
		if fErr != nil {
			errs = append(errs, fErr)
		}
		if fServiceExists {
			serviceExists = true
		}
	}

	if !serviceExists {
		errs = append(errs, NewErrInvalidField("service", fmt.Sprintf("could not find service '%s' in %s", dc.Service, dc.ComposeFiles)))
	}
	return errs
}

func lintComposeFile(path string, dc manifest.DockerCompose, fs afero.Afero) (serviceExists bool, err error) {
	if err := CheckFile(fs, path, false); err != nil {
		return false, err
	}

	composeContent, err := fs.ReadFile(path)
	if err != nil {
		return false, err
	}

	serviceExists, err = lintDockerComposeService(dc.Service, path, composeContent)
	if err != nil {
		return false, err
	}

	return serviceExists, nil
}

func lintDockerComposeService(service string, composeFile string, composeContent []byte) (serviceExists bool, err error) {
	var compose struct {
		Version  string
		Services map[string]interface{}
	}
	e := yaml.Unmarshal(composeContent, &compose)
	if e != nil {
		return false, ErrFileInvalid.WithValue(e.Error())
	}

	if compose.Services == nil || strings.HasPrefix(compose.Version, "1") {
		fmt.Println("SSS", compose.Services == nil)
		return false, ErrDockerComposeVersion.WithFile(composeFile).AsWarning()
	}

	_, serviceExists = compose.Services[service]
	return serviceExists, nil
}
