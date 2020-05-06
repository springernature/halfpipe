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

type linter struct {
	fs                 afero.Afero
	deprecatedPrefixes []string
	deprecationDate    time.Time
	todaysDate         time.Time
}

func (l linter) Lint(man manifest.Manifest) (result result.LintResult) {
	result.Linter = "Docker Registries"
	result.DocsURL = "https://docs.halfpipe.io/docker-registry/"

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
			if l.todaysDate.Before(l.deprecationDate.AddDate(0, -1, 0)) || man.FeatureToggles.DisableDockerRegistryLinter() {
				result.AddWarning(err)
			} else {
				result.AddError(fmt.Errorf("%s .... To supress this error use the feature toggle as described in '%s', you have until %s to migrate", err.Error(), "https://ee-discourse.springernature.io/t/internal-docker-registries-end-of-life/", l.deprecationDate))
			}

		}
	}
	return
}

func (l linter) lintRunTask(task manifest.Run) (err error) {
	for _, deprecated := range l.deprecatedPrefixes {
		if strings.HasPrefix(task.Docker.Image, deprecated) {
			return linterrors.NewInvalidField("docker.image", fmt.Sprintf("the docker image '%s' references the deprecated docker registry '%s'", task.Docker.Image, deprecated))
		}
	}
	return nil
}

func (l linter) lintDockerCompose(task manifest.DockerCompose) (badError bool, err error) {
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

func (l linter) lintDockerPush(task manifest.DockerPush) (badError bool, err error) {
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

func NewDeprecatedDockerRegistriesLinter(fs afero.Afero, deprecatedPrefixes []string, deprecationDate time.Time, todaysDate time.Time) Linter {
	return linter{
		fs:                 fs,
		deprecatedPrefixes: deprecatedPrefixes,
		deprecationDate:    deprecationDate,
		todaysDate:         todaysDate,
	}
}
