package linters

import (
	"testing"

	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func testTaskLinter() taskLinter {
	return taskLinter{
		Fs: afero.Afero{Fs: afero.NewMemMapFs()},
	}
}

func TestAtLeastOneTaskExists(t *testing.T) {
	man := manifest.Manifest{}
	taskLinter := testTaskLinter()

	result := taskLinter.Lint(man)
	assert.Len(t, result.Errors, 1)
	assertMissingField(t, "tasks", result.Errors[0])
}

func TestCallsOutToTheLintersCorrectly(t *testing.T) {
	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.Run{},
			manifest.DeployCF{
				PrePromote: []manifest.Task{
					manifest.Run{},
					manifest.DeployCF{
						PrePromote: []manifest.Task{
							manifest.Run{},
						},
					},
					manifest.DockerPush{},
					manifest.DockerCompose{},
					manifest.ConsumerIntegrationTest{},
					manifest.DeployMLZip{},
					manifest.DeployMLModules{},
				},
			},
			manifest.DockerPush{},
			manifest.DockerCompose{},
			manifest.ConsumerIntegrationTest{},
			manifest.DeployMLZip{},
			manifest.DeployMLModules{},
		},
	}

	calledLintRunTask := false
	calledLintRunTaskNum := 0
	calledLintDeployCFTask := false
	calledLintDeployCFTaskNum := 0
	calledLintDockerPushTask := false
	calledLintDockerPushTaskNum := 0
	calledLintDockerComposeTask := false
	calledLintDockerComposeTaskNum := 0
	calledLintConsumerIntegrationTestTask := false
	calledLintConsumerIntegrationTestTaskNum := 0
	calledLintDeployMLZipTask := false
	calledLintDeployMLZipTaskNum := 0
	calledLintDeployMLModulesTask := false
	calledLintDeployMLModulesTaskNum := 0
	calledLintPrePromoteTasks := false
	calledLintPrePromoteTasksNum := 0

	taskLinter := taskLinter{
		Fs: afero.Afero{
			Fs: nil,
		},
		lintRunTask: func(task manifest.Run, fs afero.Afero) (errs []error, warnings []error) {
			calledLintRunTask = true
			calledLintRunTaskNum += 1
			return
		},
		lintDeployCFTask: func(task manifest.DeployCF, fs afero.Afero) (errs []error, warnings []error) {
			calledLintDeployCFTask = true
			calledLintDeployCFTaskNum += 1
			return
		},
		LintPrePromoteTask: func(tasks manifest.Task) (errs []error, warnings []error) {
			calledLintPrePromoteTasks = true
			calledLintPrePromoteTasksNum += 1
			return
		},
		lintDockerPushTask: func(task manifest.DockerPush, fs afero.Afero) (errs []error, warnings []error) {
			calledLintDockerPushTask = true
			calledLintDockerPushTaskNum += 1
			return
		},
		lintDockerComposeTask: func(task manifest.DockerCompose, fs afero.Afero) (errs []error, warnings []error) {
			calledLintDockerComposeTask = true
			calledLintDockerComposeTaskNum += 1
			return
		},
		lintConsumerIntegrationTestTask: func(cit manifest.ConsumerIntegrationTest, providerHostRequired bool) (errs []error, warnings []error) {
			calledLintConsumerIntegrationTestTask = true
			calledLintConsumerIntegrationTestTaskNum += 1
			return
		},
		lintDeployMLZipTask: func(task manifest.DeployMLZip) (errs []error, warnings []error) {
			calledLintDeployMLZipTask = true
			calledLintDeployMLZipTaskNum += 1
			return
		},
		lintDeployMLModulesTask: func(task manifest.DeployMLModules) (errs []error, warnings []error) {
			calledLintDeployMLModulesTask = true
			calledLintDeployMLModulesTaskNum += 1
			return
		},
	}

	taskLinter.Lint(man)

	assert.True(t, calledLintRunTask)
	assert.Equal(t, 3, calledLintRunTaskNum)

	assert.True(t, calledLintDeployCFTask)
	assert.Equal(t, 2, calledLintDeployCFTaskNum)

	assert.True(t, calledLintPrePromoteTasks)
	assert.Equal(t, 8, calledLintPrePromoteTasksNum)

	assert.True(t, calledLintDockerPushTask)
	assert.Equal(t, 2, calledLintDockerPushTaskNum)

	assert.True(t, calledLintDockerComposeTask)
	assert.Equal(t, 2, calledLintDockerComposeTaskNum)

	assert.True(t, calledLintConsumerIntegrationTestTask)
	assert.Equal(t, 2, calledLintConsumerIntegrationTestTaskNum)

	assert.True(t, calledLintDeployMLZipTask)
	assert.Equal(t, 2, calledLintDeployMLZipTaskNum)

	assert.True(t, calledLintDeployMLModulesTask)
	assert.Equal(t, 2, calledLintDeployMLModulesTaskNum)
}

func TestMergesTheErrorsAndWarningsCorrectly(t *testing.T) {
	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.Run{},
			manifest.DeployCF{
				PrePromote: []manifest.Task{
					manifest.Run{},
					manifest.DockerPush{},
				},
			},
			manifest.DeployMLZip{},
			manifest.DeployMLModules{},
		},
	}

	runErr1 := errors.New("runErr1")
	runErr2 := errors.New("runErr2")
	runWarn1 := errors.New("runWarn1")

	deployErr := errors.New("deployErr")

	prePromoteErr := errors.New("prePromoteErr")
	prePromoteWarn := errors.New("prePromoteWarn")

	dockerPushErr := errors.New("dockerPushErr")
	dockerPushWarn := errors.New("dockerPushWarn")

	deployMlZipErr := errors.New("deployMlZipErr")

	deployMlModulesWarn := errors.New("deployMlModulesWarn")
	taskLinter := taskLinter{
		Fs: afero.Afero{
			Fs: nil,
		},
		lintRunTask: func(task manifest.Run, fs afero.Afero) (errs []error, warnings []error) {
			return []error{runErr1, runErr2}, []error{runWarn1}
		},
		lintDeployCFTask: func(task manifest.DeployCF, fs afero.Afero) (errs []error, warnings []error) {
			return []error{deployErr}, []error{}
		},
		LintPrePromoteTask: func(tasks manifest.Task) (errs []error, warnings []error) {
			return []error{prePromoteErr}, []error{prePromoteWarn}
		},
		lintDockerPushTask: func(task manifest.DockerPush, fs afero.Afero) (errs []error, warnings []error) {
			return []error{dockerPushErr}, []error{dockerPushWarn}
		},
		lintDeployMLZipTask: func(task manifest.DeployMLZip) (errs []error, warnings []error) {
			return []error{deployMlZipErr}, []error{}
		},
		lintDeployMLModulesTask: func(task manifest.DeployMLModules) (errs []error, warnings []error) {
			return []error{}, []error{deployMlModulesWarn}

		},
	}

	result := taskLinter.Lint(man)

	errorsToStrings := func(errs []error) (out []string) {
		for _, e := range errs {
			out = append(out, e.Error())
		}
		return
	}

	assert.Equal(t, []string{
		fmt.Sprintf("tasks[0] %s", runErr1),
		fmt.Sprintf("tasks[0] %s", runErr2),
		fmt.Sprintf("tasks[1] %s", deployErr),
		fmt.Sprintf("tasks[1].pre_promote[0] %s", prePromoteErr),
		fmt.Sprintf("tasks[1].pre_promote[1] %s", prePromoteErr),
		fmt.Sprintf("tasks[1].pre_promote[0] %s", runErr1),
		fmt.Sprintf("tasks[1].pre_promote[0] %s", runErr2),
		fmt.Sprintf("tasks[1].pre_promote[1] %s", dockerPushErr),
		fmt.Sprintf("tasks[2] %s", deployMlZipErr),
	}, errorsToStrings(result.Errors))
	assert.Equal(t, []string{
		fmt.Sprintf("tasks[0] %s", runWarn1),
		fmt.Sprintf("tasks[1].pre_promote[0] %s", prePromoteWarn),
		fmt.Sprintf("tasks[1].pre_promote[1] %s", prePromoteWarn),
		fmt.Sprintf("tasks[1].pre_promote[0] %s", runWarn1),
		fmt.Sprintf("tasks[1].pre_promote[1] %s", dockerPushWarn),
		fmt.Sprintf("tasks[3] %s", deployMlModulesWarn),
	}, errorsToStrings(result.Warnings))
}
