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
	lintRunTask                     func(task manifest.Run, fs afero.Afero) (errs []error, warnings []error)
	lintDeployCFTask                func(task manifest.DeployCF, fs afero.Afero) (errs []error, warnings []error)
	LintPrePromoteTask              func(task manifest.Task) (errs []error, warnings []error)
	lintDockerPushTask              func(task manifest.DockerPush, fs afero.Afero) (errs []error, warnings []error)
	lintDockerComposeTask           func(task manifest.DockerCompose, fs afero.Afero) (errs []error, warnings []error)
	lintConsumerIntegrationTestTask func(task manifest.ConsumerIntegrationTest, providerHostRequired bool) (errs []error, warnings []error)
	lintDeployMLZipTask             func(task manifest.DeployMLZip) (errs []error, warnings []error)
	lintDeployMLModulesTask         func(task manifest.DeployMLModules) (errs []error, warnings []error)
}

func NewTasksLinter(fs afero.Afero) taskLinter {
	return taskLinter{
		Fs:                              fs,
		lintRunTask:                     tasks.LintRunTask,
		lintDeployCFTask:                tasks.LintDeployCFTask,
		LintPrePromoteTask:              tasks.LintPrePromoteTask,
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

	errs, warnings := linter.lintTasks("", man.Tasks)
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
	var taskID string
	if listName == "" {
		taskID = fmt.Sprintf("tasks[%v]", index)
	} else {
		taskID = fmt.Sprintf("%s[%v]", listName, index)
	}

	prefixErrors := prefixErrorsWithIndex(taskID)
	switch task := t.(type) {
	case manifest.Run:
		errs, warnings = prefixErrors(linter.lintRunTask(task, linter.Fs))
	case manifest.DeployCF:
		errs, warnings = prefixErrors(linter.lintDeployCFTask(task, linter.Fs))

		if len(task.PrePromote) > 0 {
			for pI, preTask := range task.PrePromote {
				e, w := prefixErrorsWithIndex(fmt.Sprintf("%s.pre_promote[%v]", taskID, pI))(linter.LintPrePromoteTask(preTask))
				errs = append(errs, e...)
				warnings = append(warnings, w...)
			}

			subErrors, subWarnings := linter.lintTasks(fmt.Sprintf("%s.pre_promote", taskID), task.PrePromote)
			errs = append(errs, subErrors...)
			warnings = append(warnings, subWarnings...)
		}
	case manifest.DockerPush:
		errs, warnings = prefixErrors(linter.lintDockerPushTask(task, linter.Fs))
	case manifest.DockerCompose:
		errs, warnings = prefixErrors(linter.lintDockerComposeTask(task, linter.Fs))
	case manifest.ConsumerIntegrationTest:
		if listName == "tasks" {
			errs, warnings = prefixErrors(linter.lintConsumerIntegrationTestTask(task, true))
		} else {
			errs, warnings = prefixErrors(linter.lintConsumerIntegrationTestTask(task, false))
		}
	case manifest.DeployMLZip:
		errs, warnings = prefixErrors(linter.lintDeployMLZipTask(task))
	case manifest.DeployMLModules:
		errs, warnings = prefixErrors(linter.lintDeployMLModulesTask(task))
	default:
		errs = append(errs, errors.NewInvalidField("task", fmt.Sprintf("%s is not a known task", taskID)))
	}

	return
}

func prefixErrorsWithIndex(prefix string) func(errs, warns []error) (rE []error, rW []error) {
	return func(errs, warns []error) (rE []error, rW []error) {
		for _, e := range errs {
			rE = append(rE, fmt.Errorf("%s %s", prefix, e))
		}
		for _, w := range warns {
			rW = append(rW, fmt.Errorf("%s %s", prefix, w))
		}
		return
	}
}
