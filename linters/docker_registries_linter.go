package linters

import (
	"fmt"
	"github.com/springernature/halfpipe/linters/linterrors"
	"github.com/springernature/halfpipe/linters/result"
	"github.com/springernature/halfpipe/manifest"
	"strings"
	"time"
)

type linter struct {
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
		}

		if err != nil {
			if l.todaysDate.Before(l.deprecationDate.AddDate(0,-1, 0)) || man.FeatureToggles.DisableDockerRegistryLinter(){
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

func NewDeprecatedDockerRegistriesLinter(deprecatedPrefixes []string, deprecationDate time.Time, todaysDate time.Time) Linter {
	return linter{
		deprecatedPrefixes: deprecatedPrefixes,
		deprecationDate:    deprecationDate,
		todaysDate:         todaysDate,
	}
}
