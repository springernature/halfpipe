package linters

import (
	"fmt"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/linters/errors"
	"github.com/springernature/halfpipe/linters/result"
	"github.com/springernature/halfpipe/linters/tasks"
	"github.com/springernature/halfpipe/manifest"
)

type taskLinter struct {
	Fs afero.Afero
}

func NewTasksLinter(fs afero.Afero) taskLinter {
	return taskLinter{fs}
}

func (linter taskLinter) Lint(man manifest.Manifest) (result result.LintResult) {
	result.Linter = "Tasks"
	result.DocsURL = "https://docs.halfpipe.io/manifest/#tasks"

	if len(man.Tasks) == 0 {
		result.AddError(errors.NewMissingField("tasks"))
		return
	}

	errs, warnings := linter.lintTasks("tasks", man.Tasks)
	result.AddError(errs...)
	result.AddWarning(warnings...)

	return
}

func (linter taskLinter) lintTasks(listName string, ts []manifest.Task) (errs []error, warnings []error) {
	for i, t := range ts {
		e, w := linter.lintTask(listName, i, t)
		errs = append(errs, e...)
		warnings = append(warnings, w...)
	}

	return
}

func (linter taskLinter) lintTask(listName string, index int, t manifest.Task) (errs []error, warnings []error) {
	taskID := fmt.Sprintf("%s[%v]", listName, index)
	switch task := t.(type) {
	case manifest.Run:
		errs, warnings = tasks.LintRunTask(task, taskID, linter.Fs)
	case manifest.DeployCF:
		errs, warnings = tasks.LintDeployCFTask(task, taskID, linter.Fs)

		if len(task.PrePromote) > 0 {
			subErrors, subWarnings := linter.lintTasks(fmt.Sprintf("%s.pre_promote", taskID), task.PrePromote)
			errs = append(errs, subErrors...)
			warnings = append(errs, subWarnings...)

		}
	case manifest.DockerPush:
		errs, warnings = tasks.LintDockerPushTask(task, taskID, linter.Fs)
	case manifest.DockerCompose:
		errs, warnings = tasks.LintDockerComposeTask(task, taskID, linter.Fs)
	case manifest.ConsumerIntegrationTest:
		if listName == "tasks" {
			errs, warnings = tasks.LintConsumerIntegrationTestTask(task, taskID, true)
		} else {
			errs, warnings = tasks.LintConsumerIntegrationTestTask(task, taskID, false)
		}
	case manifest.DeployMLZip:
		errs, warnings = tasks.LintDeployMLZipTask(task, taskID)
	case manifest.DeployMLModules:
		errs, warnings = tasks.LintDeployMLModulesTask(task, taskID)
	default:
		errs = append(errs, errors.NewInvalidField("task", fmt.Sprintf("%s is not a known task", taskID)))
	}

	return
}
