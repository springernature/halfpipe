package linters

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/springernature/halfpipe/linters/linterrors"
	"github.com/springernature/halfpipe/linters/result"
	"github.com/springernature/halfpipe/manifest"
)

type ActionsLinter struct{}

func NewActionsLinter() Linter {
	return ActionsLinter{}
}

func (linter ActionsLinter) Lint(man manifest.Manifest) (result result.LintResult) {
	result.Linter = "GitHub Actions"
	result.DocsURL = "https://ee.public.springernature.app/rel-eng/github-actions/overview/"
	if man.Platform.IsActions() {
		result.AddWarning(unsupportedTasks(man.Tasks, man)...)
		result.AddWarning(unsupportedTriggers(man.Triggers)...)
		result.AddWarning(unsupportedFeatures(man.FeatureToggles)...)
	}
	return result
}

func unsupportedTasks(tasks manifest.TaskList, man manifest.Manifest) (errors []error) {
	for i, t := range tasks {
		switch task := t.(type) {
		case manifest.Parallel:
			errors = append(errors, unsupportedTasks(task.Tasks, man)...)
		case manifest.Sequence:
			errors = append(errors, unsupportedTasks(task.Tasks, man)...)
		case manifest.ConsumerIntegrationTest:
			if task.UseCovenant {
				errors = append(errors, linterrors.NewUnsupportedField(fmt.Sprintf("tasks[%v] %T.use_covenant", i, task)))
			}
		}
	}

	for i, t := range tasks {
		switch task := t.(type) {
		case manifest.Parallel, manifest.Sequence:
			//skip
		default:
			if task.IsManualTrigger() {
				errors = append(errors, linterrors.NewUnsupportedField(fmt.Sprintf("tasks[%v] %T.manual_trigger", i, task)))
			}
		}
	}

	for i, t := range tasks {
		switch task := t.(type) {
		case manifest.DeployCF:
			if task.Rolling {
				errors = append(errors, linterrors.NewUnsupportedField(fmt.Sprintf("tasks[%v] %T.rolling", i, task)))
			}
		case manifest.DockerPush:
			for _, trigger := range man.Triggers {
				if t, ok := trigger.(manifest.DockerTrigger); ok {
					if t.Image == task.Image {
						errors = append(errors, linterrors.NewUnsupportedField(fmt.Sprintf("tasks[%v] %T.image '%s' is also a trigger. This will create a loop.", i, task, t.Image)))
					}
				}
			}
		}
	}
	return errors
}

func unsupportedTriggers(triggers manifest.TriggerList) (errors []error) {

	addError := func(i int, name string) {
		errors = append(errors, linterrors.NewUnsupportedField(fmt.Sprintf("triggers[%v] %s", i, name)))
	}

	for i, trigger := range triggers {
		switch t := trigger.(type) {
		case manifest.GitTrigger:
			if t.PrivateKey != "" {
				addError(i, "private_key")
			}
			if t.URI != "" {
				addError(i, "uri")
			}
		case manifest.TimerTrigger, manifest.DockerTrigger:
			// ok
		default:
			addError(i, fmt.Sprintf("%T", trigger))
		}
	}
	return errors
}

var ErrUpdatePiplineNotImplemented = errors.New("the update-pipeline feature is not implemented, so you must always run 'halfpipe' to keep the workflow file up to date")

func unsupportedFeatures(features manifest.FeatureToggles) (errors []error) {
	if features.UpdatePipeline() {
		errors = []error{ErrUpdatePiplineNotImplemented}
	}
	return errors
}
