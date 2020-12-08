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
	result.DocsURL = "https://ee.public.springernature.app/rel-eng/halfpipe/github-actions"

	result.AddWarning(unsupportedTasks(man.Tasks)...)
	result.AddWarning(unsupportedTriggers(man.Triggers)...)

	return result
}

func unsupportedTasks(tasks manifest.TaskList) (errors []error) {
	for i, t := range tasks {
		switch task := t.(type) {
		case manifest.DockerPush, manifest.Run:
			if task.IsManualTrigger() {
				errors = append(errors, linterrors.NewUnsupportedField(fmt.Sprintf("tasks[%v] %T.manual_trigger", i, task)))
			}
		case manifest.Parallel:
			errors = append(errors, unsupportedTasks(task.Tasks)...)
		case manifest.Sequence:
			errors = append(errors, unsupportedTasks(task.Tasks)...)
		default:
			errors = append(errors, linterrors.NewUnsupportedField(fmt.Sprintf("tasks[%v] %T", i, task)))
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
