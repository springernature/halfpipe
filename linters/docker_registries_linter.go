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
	for _, task := range man.Tasks.Flatten() {
		var err error
		switch task.(type) {
		case manifest.Run:
			err = l.lintRunTask(task.(manifest.Run))
		case manifest.DockerCompose:
			e, badErr := l.lintDockerCompose(task.(manifest.DockerCompose))
			if badErr {
				result.AddError(e)
				return
			}
			err = e
		case manifest.DockerPush:
			e, badErr := l.lintDockerPush(task.(manifest.DockerPush))
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
				result.AddError(fmt.Errorf("%s .... To supress this error use the feature toggle '%s', you have until %s to migrate", err.Error(), manifest.FeatureToggleDisableDeprecatedDockerRegistryError, l.deprecationDate))
			}

		}
	}
	return
}

func (l linter) lintRunTask(task manifest.Run) (err error) {
	for _, deprecated := range l.deprecatedPrefixes {
		if strings.HasPrefix(task.Docker.Image, deprecated) {
			return linterrors.NewInvalidField("docker.image", fmt.Sprintf("The docker image '%s' references the deprecated docker registry '%s'", task.Docker.Image, deprecated))
		}
	}
	return nil
}

func (l linter) lintDockerCompose(task manifest.DockerCompose) (err error, badError bool) {
	composeFile, err := l.fs.ReadFile(task.ComposeFile)
	if err != nil {
		return err, true
	}
	for _, deprecated := range l.deprecatedPrefixes {
		if strings.Contains(string(composeFile), deprecated) {
			return linterrors.NewInvalidField("composeFile", fmt.Sprintf("'%s' references the deprecated docker registry '%s'", task.ComposeFile, deprecated)), false
		}
	}
	return nil, false
}

func (l linter) lintDockerPush(task manifest.DockerPush) (err error, badError bool) {
	dockerContent, err := l.fs.ReadFile(task.DockerfilePath)
	if err != nil {
		return err, true
	}
	for _, deprecated := range l.deprecatedPrefixes {
		if strings.HasPrefix(task.Image, deprecated) {
			return linterrors.NewInvalidField("image", fmt.Sprintf("'%s' references the deprecated docker registry '%s'", task.Image, deprecated)), false
		}
		if strings.Contains(string(dockerContent), deprecated) {
			return linterrors.NewInvalidField("dockerfile_path", fmt.Sprintf("'%s' references the deprecated docker registry '%s'", task.DockerfilePath, deprecated)), false
		}
	}
	return nil, false

}

func NewDeprecatedDockerRegistriesLinter(fs afero.Afero, deprecatedPrefixes []string, deprecationDate time.Time, todaysDate time.Time) Linter {
	return linter{
		fs:                 fs,
		deprecatedPrefixes: deprecatedPrefixes,
		deprecationDate:    deprecationDate,
		todaysDate:         todaysDate,
	}
}
