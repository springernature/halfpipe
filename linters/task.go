package linters

import (
	"code.cloudfoundry.org/cli/util/manifestparser"
	"fmt"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/cf"
	"github.com/springernature/halfpipe/manifest"
	"sort"
	"strings"
	"time"
)

type taskLinter struct {
	Fs                              afero.Afero
	lintRunTask                     func(task manifest.Run, fs afero.Afero, os string) []error
	lintDeployCFTask                func(task manifest.DeployCF, readCfManifest cf.ManifestReader, fs afero.Afero) []error
	lintDeployKateeTask             func(task manifest.DeployKatee, man manifest.Manifest, fs afero.Afero) []error
	LintPrePromoteTask              func(task manifest.Task) []error
	lintDockerPushTask              func(task manifest.DockerPush, fs afero.Afero) []error
	lintDockerComposeTask           func(task manifest.DockerCompose, fs afero.Afero) []error
	lintConsumerIntegrationTestTask func(task manifest.ConsumerIntegrationTest, providerHostRequired bool) []error
	lintDeployMLZipTask             func(task manifest.DeployMLZip) []error
	lintDeployMLModulesTask         func(task manifest.DeployMLModules) []error
	lintArtifacts                   func(currentTask manifest.Task, previousTasks []manifest.Task) []error
	lintParallel                    func(parallelTask manifest.Parallel) []error
	lintSequence                    func(seqTask manifest.Sequence, cameFromAParallel bool) []error
	lintNotifications               func(task manifest.Task) []error
	os                              string
}

func NewTasksLinter(fs afero.Afero, os string) taskLinter {
	return taskLinter{
		Fs:                              fs,
		lintRunTask:                     LintRunTask,
		lintDeployCFTask:                LintDeployCFTask,
		lintDeployKateeTask:             LintDeployKateeTask,
		LintPrePromoteTask:              LintPrePromoteTask,
		lintDockerPushTask:              LintDockerPushTask,
		lintDockerComposeTask:           LintDockerComposeTask,
		lintConsumerIntegrationTestTask: LintConsumerIntegrationTestTask,
		lintDeployMLZipTask:             LintDeployMLZipTask,
		lintDeployMLModulesTask:         LintDeployMLModulesTask,
		lintArtifacts:                   LintArtifacts,
		lintParallel:                    LintParallelTask,
		lintSequence:                    LintSequenceTask,
		lintNotifications:               LintNotifications,
		os:                              os,
	}
}

func (linter taskLinter) Lint(man manifest.Manifest) (result LintResult) {
	result.Linter = "Tasks"
	result.DocsURL = "https://ee.public.springernature.app/rel-eng/halfpipe/manifest/#tasks"

	if len(man.Tasks) == 0 {
		result.Add(NewErrMissingField("tasks").AsWarning())
		return result
	}

	errs := linter.lintTasks("", man.Tasks, man, []manifest.Task{}, true, false)
	sortErrors(errs)
	result.Add(errs...)
	return result
}

func (linter taskLinter) lintTasks(listName string, ts []manifest.Task, man manifest.Manifest, previousTasks []manifest.Task, lintArtifact, cameFromParallel bool) (rE []error) {
	for index, t := range ts {
		previousTasks = append(previousTasks, ts[:index]...)

		var taskID string
		if listName == "" {
			taskID = fmt.Sprintf("tasks[%v]", index)
		} else {
			taskID = fmt.Sprintf("%s[%v]", listName, index)
		}

		wrapWithIndex := wrapErrorsWithIndex(taskID)

		lintTimeout := true

		var errs []error
		switch task := t.(type) {
		case manifest.Run:
			errs = linter.lintRunTask(task, linter.Fs, linter.os)
		case manifest.DeployCF:
			errs = linter.lintDeployCFTask(task, manifestparser.ManifestParser{}.InterpolateAndParse, linter.Fs)

			if len(errs) == 0 && len(task.PrePromote) > 0 {
				for pI, preTask := range task.PrePromote {
					prePromotePrefixer := wrapErrorsWithIndex(fmt.Sprintf("%s.pre_promote[%v]", taskID, pI))
					e := linter.LintPrePromoteTask(preTask)
					errs = append(errs, prePromotePrefixer(e)...)
				}

				subErrors := linter.lintTasks(fmt.Sprintf("%s.pre_promote", taskID), task.PrePromote, man, previousTasks, false, false)
				errs = append(errs, subErrors...)
			}
		case manifest.DeployKatee:
			errs = linter.lintDeployKateeTask(task, man, linter.Fs)
		case manifest.DockerPush:
			errs = linter.lintDockerPushTask(task, linter.Fs)
		case manifest.DockerCompose:
			errs = linter.lintDockerComposeTask(task, linter.Fs)
		case manifest.ConsumerIntegrationTest:
			if listName == "tasks" {
				errs = linter.lintConsumerIntegrationTestTask(task, true)
			} else {
				errs = linter.lintConsumerIntegrationTestTask(task, false)
			}
		case manifest.DeployMLZip:
			errs = linter.lintDeployMLZipTask(task)
		case manifest.DeployMLModules:
			errs = linter.lintDeployMLModulesTask(task)
		case manifest.Update:
		case manifest.Parallel:
			errs = linter.lintParallel(task)
			subErrors := linter.lintTasks(taskID, task.Tasks, man, previousTasks, true, true)
			errs = append(errs, subErrors...)
			lintTimeout = false
			lintArtifact = false
		case manifest.Sequence:
			errs = linter.lintSequence(task, cameFromParallel)
			subErrors := linter.lintTasks(taskID, task.Tasks, man, previousTasks, true, false)
			errs = append(errs, subErrors...)
			lintTimeout = false
			lintArtifact = false
		default:
			errs = append(errs, NewErrInvalidField("task", fmt.Sprintf("%s is not a known task", taskID)))
		}

		if t.ReadsFromArtifacts() && lintArtifact {
			artifactErr := linter.lintArtifacts(t, previousTasks)
			errs = append(errs, artifactErr...)
		}

		if lintTimeout && t.GetTimeout() != "" {
			_, err := time.ParseDuration(t.GetTimeout())
			if err != nil {
				errs = append(errs, NewErrInvalidField("timeout", err.Error()))
			}
		}

		errs = append(errs, linter.lintNotifications(t)...)

		rE = append(rE, wrapWithIndex(errs)...)
	}

	return rE
}

func sortErrors(errs []error) {
	getPrefix := func(err error) string {
		return strings.Split(err.Error(), " ")[0]
	}

	sort.Slice(errs, func(i, j int) bool {
		return getPrefix(errs[i]) < getPrefix(errs[j])
	})
}

func wrapErrorsWithIndex(prefix string) func(errs []error) (rE []error) {
	// Since we are calling lintTasks recursively we end up in a situation where
	// error already contains the prefix.
	return func(errs []error) (rE []error) {
		for _, e := range errs {
			if strings.HasPrefix(e.Error(), prefix) {
				rE = append(rE, e)
			} else {
				rE = append(rE, fmt.Errorf("%s %w", prefix, e))
			}

		}
		return rE
	}
}
