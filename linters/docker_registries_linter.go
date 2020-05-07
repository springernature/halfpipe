package linters

import (
	"fmt"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/linters/linterrors"
	"github.com/springernature/halfpipe/linters/result"
	"github.com/springernature/halfpipe/manifest"
	"strings"
	"time"
)

type dockerRegistriesLinter struct {
	fs                 afero.Afero
	deprecatedPrefixes []string
	deprecationDate    time.Time
	todaysDate         time.Time
}

func (l dockerRegistriesLinter) Lint(man manifest.Manifest) (result result.LintResult) {
	result.Linter = "Deprecated Docker Registries"
	result.DocsURL = "https://ee-discourse.springernature.io/t/internal-docker-registries-end-of-life/"

	for _, task := range man.Tasks.Flatten() {
		var err error
		switch task.(type) {
		case manifest.Run:
			err = l.lintRunTask(task.(manifest.Run))
		case manifest.DockerCompose:
			badErr, e := l.lintDockerCompose(task.(manifest.DockerCompose))
			if badErr {
				result.AddError(e)
				return
			}
			err = e
		case manifest.DockerPush:
			badErr, e := l.lintDockerPush(task.(manifest.DockerPush))
			if badErr {
				result.AddError(e)
				return
			}
			err = e
		}

		if err != nil {
			if man.FeatureToggles.DisableDeprecatedDockerRegistryError() {
				result.AddWarning(err)
			} else {
				result.AddError(fmt.Errorf("%s. To supress this error use the feature toggle as described in <https://ee-discourse.springernature.io/t/internal-docker-registries-end-of-life/>, you have until %s to migrate", err.Error(), l.deprecationDate.Format("02 January 2006")))
			}

		}
	}
	return
}

func (l dockerRegistriesLinter) lintRunTask(task manifest.Run) (err error) {
	for _, deprecated := range l.deprecatedPrefixes {
		if strings.HasPrefix(task.Docker.Image, deprecated) {
			return linterrors.NewInvalidField("docker.image", fmt.Sprintf("the docker image '%s' references the deprecated docker registry '%s'", task.Docker.Image, deprecated))
		}
	}
	return nil
}

func (l dockerRegistriesLinter) lintDockerCompose(task manifest.DockerCompose) (badError bool, err error) {
	composeFile, err := l.fs.ReadFile(task.ComposeFile)
	if err != nil {
		return true, err
	}
	for _, deprecated := range l.deprecatedPrefixes {
		if strings.Contains(string(composeFile), deprecated) {
			return false, linterrors.NewInvalidField("composeFile", fmt.Sprintf("'%s' references the deprecated docker registry '%s'", task.ComposeFile, deprecated))
		}
	}
	return false, nil
}

func (l dockerRegistriesLinter) lintDockerPush(task manifest.DockerPush) (badError bool, err error) {
	dockerContent, err := l.fs.ReadFile(task.DockerfilePath)
	if err != nil {
		return true, err
	}
	for _, deprecated := range l.deprecatedPrefixes {
		if strings.HasPrefix(task.Image, deprecated) {
			return false, linterrors.NewInvalidField("image", fmt.Sprintf("'%s' references the deprecated docker registry '%s'", task.Image, deprecated))
		}
		if strings.Contains(string(dockerContent), deprecated) {
			return false, linterrors.NewInvalidField("dockerfile_path", fmt.Sprintf("'%s' references the deprecated docker registry '%s'", task.DockerfilePath, deprecated))
		}
	}
	return false, nil

}

func NewDeprecatedDockerRegistriesLinter(fs afero.Afero, deprecatedPrefixes []string) Linter {
	return dockerRegistriesLinter{
		fs:                 fs,
		deprecatedPrefixes: deprecatedPrefixes,
	}
}
