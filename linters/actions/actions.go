package linters

import (
	"fmt"
	"github.com/springernature/halfpipe/linters/linterrors"
	"github.com/springernature/halfpipe/linters/result"
	"github.com/springernature/halfpipe/manifest"
)

type ActionsLinter struct{}

func (linter ActionsLinter) Lint(man manifest.Manifest) (result result.LintResult) {
	result.Linter = "GitHub Actions"
	result.DocsURL = "https://ee.public.springernature.app/rel-eng/github-actions/overview/"

	result.AddWarning(unsupportedTasks(man.Tasks, man)...)
	result.AddWarning(unsupportedTriggers(man.Triggers)...)

	return result
}

func unsupportedTasks(tasks manifest.TaskList, man manifest.Manifest) (errors []error) {
	for i, t := range tasks {
		switch task := t.(type) {
		case manifest.Parallel:
			errors = append(errors, unsupportedTasks(task.Tasks, man)...)
		case manifest.Sequence:
			errors = append(errors, unsupportedTasks(task.Tasks, man)...)
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
			if t.GitCryptKey != "" {
				addError(i, "git_crypt_key")
			}
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
