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

	result.Warnings = unsupportedTasks(man.Tasks)

	return result
}

func unsupportedTasks(tasks manifest.TaskList) (errors []error) {
	for i, task := range tasks {
		switch task.(type) {
		case manifest.DockerPush:
			//ok
		default:
			errors = append(errors, linterrors.NewUnsupportedField(fmt.Sprintf("task[%v] %T", i, task)))
		}
	}
	return errors
}
