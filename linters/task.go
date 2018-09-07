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
	Fs                              afero.Afero
	lintRunTask                     func(task manifest.Run, taskID string, fs afero.Afero) (errs []error, warnings []error)
	lintDeployCFTask                func(task manifest.DeployCF, taskID string, fs afero.Afero) (errs []error, warnings []error)
	lintDockerPushTask              func(task manifest.DockerPush, taskID string, fs afero.Afero) (errs []error, warnings []error)
	lintDockerComposeTask           func(task manifest.DockerCompose, taskID string, fs afero.Afero) (errs []error, warnings []error)
	lintConsumerIntegrationTestTask func(cit manifest.ConsumerIntegrationTest, taskID string, providerHostRequired bool) (errs []error, warnings []error)
	lintDeployMLZipTask             func(task manifest.DeployMLZip, taskID string) (errs []error, warnings []error)
	lintDeployMLModulesTask         func(task manifest.DeployMLModules, taskID string) (errs []error, warnings []error)
}

func NewTasksLinter(fs afero.Afero) taskLinter {
	return taskLinter{
		Fs:                              fs,
		lintRunTask:                     tasks.LintRunTask,
		lintDeployCFTask:                tasks.LintDeployCFTask,
		lintDockerPushTask:              tasks.LintDockerPushTask,
		lintDockerComposeTask:           tasks.LintDockerComposeTask,
		lintConsumerIntegrationTestTask: tasks.LintConsumerIntegrationTestTask,
		lintDeployMLZipTask:             tasks.LintDeployMLZipTask,
		lintDeployMLModulesTask:         tasks.LintDeployMLModulesTask,
	}
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
		errs, warnings = linter.lintRunTask(task, taskID, linter.Fs)
	case manifest.DeployCF:
		errs, warnings = linter.lintDeployCFTask(task, taskID, linter.Fs)

		if len(task.PrePromote) > 0 {
			subErrors, subWarnings := linter.lintTasks(fmt.Sprintf("%s.pre_promote", taskID), task.PrePromote)
			errs = append(errs, subErrors...)
			warnings = append(warnings, subWarnings...)
		}
	case manifest.DockerPush:
		errs, warnings = linter.lintDockerPushTask(task, taskID, linter.Fs)
	case manifest.DockerCompose:
		errs, warnings = linter.lintDockerComposeTask(task, taskID, linter.Fs)
	case manifest.ConsumerIntegrationTest:
		if listName == "tasks" {
			errs, warnings = linter.lintConsumerIntegrationTestTask(task, taskID, true)
		} else {
			errs, warnings = linter.lintConsumerIntegrationTestTask(task, taskID, false)
		}
	case manifest.DeployMLZip:
		errs, warnings = linter.lintDeployMLZipTask(task, taskID)
	case manifest.DeployMLModules:
		errs, warnings = linter.lintDeployMLModulesTask(task, taskID)
	default:
		errs = append(errs, errors.NewInvalidField("task", fmt.Sprintf("%s is not a known task", taskID)))
	}

	return
}
