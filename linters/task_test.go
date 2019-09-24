package linters

import (
	"testing"

	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/linters/tasks"
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
			manifest.Parallel{
				Tasks: manifest.TaskList{

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
				},
			},
			manifest.DockerCompose{},
			manifest.ConsumerIntegrationTest{},
			manifest.DeployMLZip{},
			manifest.DeployMLModules{},
			manifest.Parallel{
				Tasks: manifest.TaskList{
					manifest.Seq{
						Tasks: manifest.TaskList{
							manifest.Run{},
							manifest.Run{},
						},
					},
					manifest.Seq{
						Tasks: manifest.TaskList{
							manifest.Run{},
							manifest.Run{},
						},
					},
				},
			},
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
	calledLintParallelTasks := false
	calledLintParallelTasksNum := 0
	calledLintSeqTasks := false
	var wasCalledFromParallelTask []bool
	calledLintSeqTasksNum := 0

	taskLinter := taskLinter{
		Fs: afero.Afero{
			Fs: nil,
		},
		lintRunTask: func(task manifest.Run, fs afero.Afero, os string) (errs []error, warnings []error) {
			calledLintRunTask = true
			calledLintRunTaskNum++
			return
		},
		lintDeployCFTask: func(task manifest.DeployCF, fs afero.Afero) (errs []error, warnings []error) {
			calledLintDeployCFTask = true
			calledLintDeployCFTaskNum++
			return
		},
		LintPrePromoteTask: func(tasks manifest.Task) (errs []error, warnings []error) {
			calledLintPrePromoteTasks = true
			calledLintPrePromoteTasksNum++
			return
		},
		lintDockerPushTask: func(task manifest.DockerPush, fs afero.Afero) (errs []error, warnings []error) {
			calledLintDockerPushTask = true
			calledLintDockerPushTaskNum++
			return
		},
		lintDockerComposeTask: func(task manifest.DockerCompose, fs afero.Afero) (errs []error, warnings []error) {
			calledLintDockerComposeTask = true
			calledLintDockerComposeTaskNum++
			return
		},
		lintConsumerIntegrationTestTask: func(cit manifest.ConsumerIntegrationTest, providerHostRequired bool) (errs []error, warnings []error) {
			calledLintConsumerIntegrationTestTask = true
			calledLintConsumerIntegrationTestTaskNum++
			return
		},
		lintDeployMLZipTask: func(task manifest.DeployMLZip) (errs []error, warnings []error) {
			calledLintDeployMLZipTask = true
			calledLintDeployMLZipTaskNum++
			return
		},
		lintDeployMLModulesTask: func(task manifest.DeployMLModules) (errs []error, warnings []error) {
			calledLintDeployMLModulesTask = true
			calledLintDeployMLModulesTaskNum++
			return
		},
		lintArtifacts: func(currentTask manifest.Task, previousTasks []manifest.Task) (errs []error, warnings []error) {
			return
		},
		lintParallel: func(parallelTask manifest.Parallel) (errs []error, warnings []error) {
			calledLintParallelTasks = true
			calledLintParallelTasksNum++
			return
		},
		lintSeq: func(seqTask manifest.Seq, cameFromAParallel bool) (errs []error, warnings []error) {
			calledLintSeqTasks = true
			calledLintSeqTasksNum++
			wasCalledFromParallelTask = append(wasCalledFromParallelTask, cameFromAParallel)
			return
		},
	}

	result := taskLinter.Lint(man)
	assert.Len(t, result.Warnings, 0)
	assert.Len(t, result.Errors, 0)

	assert.True(t, calledLintRunTask)
	assert.Equal(t, 7, calledLintRunTaskNum)

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

	assert.True(t, calledLintParallelTasks)
	assert.Equal(t, 2, calledLintParallelTasksNum)

	assert.True(t, calledLintSeqTasks)
	assert.Equal(t, 2, calledLintParallelTasksNum)
	for _, called := range wasCalledFromParallelTask {
		assert.True(t, called)
	}
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
		lintRunTask: func(task manifest.Run, fs afero.Afero, os string) (errs []error, warnings []error) {
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
		lintArtifacts: func(currentTask manifest.Task, previousTasks []manifest.Task) (errs []error, warnings []error) {
			return
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
		fmt.Sprintf("tasks[2] %s", deployMlZipErr),
	}, errorsToStrings(result.Errors))
	assert.Equal(t, []string{
		fmt.Sprintf("tasks[0] %s", runWarn1),
		fmt.Sprintf("tasks[3] %s", deployMlModulesWarn),
	}, errorsToStrings(result.Warnings))
}

func TestMergesTheErrorsAndWarningsCorrectlyWithPrePromote(t *testing.T) {
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
		lintRunTask: func(task manifest.Run, fs afero.Afero, os string) (errs []error, warnings []error) {
			return []error{runErr1, runErr2}, []error{runWarn1}
		},
		lintDeployCFTask: func(task manifest.DeployCF, fs afero.Afero) (errs []error, warnings []error) {
			return
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
		lintArtifacts: func(currentTask manifest.Task, previousTasks []manifest.Task) (errs []error, warnings []error) {
			return
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
		fmt.Sprintf("tasks[1].pre_promote[0] %s", prePromoteErr),
		fmt.Sprintf("tasks[1].pre_promote[0] %s", runErr1),
		fmt.Sprintf("tasks[1].pre_promote[0] %s", runErr2),
		fmt.Sprintf("tasks[1].pre_promote[1] %s", prePromoteErr),
		fmt.Sprintf("tasks[1].pre_promote[1] %s", dockerPushErr),
		fmt.Sprintf("tasks[2] %s", deployMlZipErr),
	}, errorsToStrings(result.Errors))
	assert.Equal(t, []string{
		fmt.Sprintf("tasks[0] %s", runWarn1),
		fmt.Sprintf("tasks[1].pre_promote[0] %s", prePromoteWarn),
		fmt.Sprintf("tasks[1].pre_promote[0] %s", runWarn1),
		fmt.Sprintf("tasks[1].pre_promote[1] %s", prePromoteWarn),
		fmt.Sprintf("tasks[1].pre_promote[1] %s", dockerPushWarn),
		fmt.Sprintf("tasks[3] %s", deployMlModulesWarn),
	}, errorsToStrings(result.Warnings))
}

func TestMergesTheErrorsAndWarningsCorrectlyWithParallel(t *testing.T) {
	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.Parallel{
				Tasks: manifest.TaskList{
					manifest.Run{},
					manifest.DeployCF{
						PrePromote: []manifest.Task{
							manifest.Run{},
							manifest.DockerPush{},
						},
					},
				},
			},
			manifest.Parallel{
				Tasks: manifest.TaskList{
					manifest.DeployMLZip{},
					manifest.DeployMLModules{},
				},
			},
		},
	}

	runErr1 := errors.New("runErr1")
	runErr2 := errors.New("runErr2")
	runWarn1 := errors.New("runWarn1")

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
		lintRunTask: func(task manifest.Run, fs afero.Afero, os string) (errs []error, warnings []error) {
			return []error{runErr1, runErr2}, []error{runWarn1}
		},
		lintDeployCFTask: func(task manifest.DeployCF, fs afero.Afero) (errs []error, warnings []error) {
			return
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
		lintArtifacts: func(currentTask manifest.Task, previousTasks []manifest.Task) (errs []error, warnings []error) {
			return
		},
		lintParallel: func(parallelTask manifest.Parallel) (errs []error, warnings []error) {
			return
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
		fmt.Sprintf("tasks[0][0] %s", runErr1),
		fmt.Sprintf("tasks[0][0] %s", runErr2),
		fmt.Sprintf("tasks[0][1].pre_promote[0] %s", prePromoteErr),
		fmt.Sprintf("tasks[0][1].pre_promote[0] %s", runErr1),
		fmt.Sprintf("tasks[0][1].pre_promote[0] %s", runErr2),
		fmt.Sprintf("tasks[0][1].pre_promote[1] %s", prePromoteErr),
		fmt.Sprintf("tasks[0][1].pre_promote[1] %s", dockerPushErr),
		fmt.Sprintf("tasks[1][0] %s", deployMlZipErr),
	}, errorsToStrings(result.Errors))
	assert.Equal(t, []string{
		fmt.Sprintf("tasks[0][0] %s", runWarn1),
		fmt.Sprintf("tasks[0][1].pre_promote[0] %s", prePromoteWarn),
		fmt.Sprintf("tasks[0][1].pre_promote[0] %s", runWarn1),
		fmt.Sprintf("tasks[0][1].pre_promote[1] %s", prePromoteWarn),
		fmt.Sprintf("tasks[0][1].pre_promote[1] %s", dockerPushWarn),
		fmt.Sprintf("tasks[1][1] %s", deployMlModulesWarn),
	}, errorsToStrings(result.Warnings))
}

func TestLintArtifactsWithParallelSeq(t *testing.T) {

	t.Run("no previous steps saves artifacts", func(t *testing.T) {
		taskLinter := taskLinter{Fs: afero.Afero{},
			lintRunTask:   func(task manifest.Run, fs afero.Afero, os string) (errs []error, warnings []error) { return },
			lintParallel:  func(parallelTask manifest.Parallel) (errs []error, warnings []error) { return },
			lintSeq:       func(seqTask manifest.Seq, cameFromAParallel bool) (errs []error, warnings []error) { return },
			lintArtifacts: tasks.LintArtifacts,
		}

		man := manifest.Manifest{
			Tasks: []manifest.Task{
				manifest.Run{},
				manifest.Parallel{
					Tasks: manifest.TaskList{
						manifest.Seq{
							Tasks: manifest.TaskList{
								manifest.Run{},
								manifest.Run{RestoreArtifacts: true},
							},
						},
					},
				},
			},
		}

		result := taskLinter.Lint(man)

		assert.Len(t, result.Errors, 1)
		assert.Len(t, result.Warnings, 0)
		assert.Equal(t, "tasks[1][0][1] reads from saved artifacts, but there are no previous tasks that saves any", result.Errors[0].Error())
	})

	t.Run("a previous steps saves artifacts", func(t *testing.T) {
		taskLinter := taskLinter{Fs: afero.Afero{},
			lintRunTask:   func(task manifest.Run, fs afero.Afero, os string) (errs []error, warnings []error) { return },
			lintParallel:  func(parallelTask manifest.Parallel) (errs []error, warnings []error) { return },
			lintSeq:       func(seqTask manifest.Seq, cameFromAParallel bool) (errs []error, warnings []error) { return },
			lintArtifacts: tasks.LintArtifacts,
		}

		man := manifest.Manifest{
			Tasks: []manifest.Task{
				manifest.Run{SaveArtifacts: []string{"something"}},
				manifest.Parallel{
					Tasks: manifest.TaskList{
						manifest.Seq{
							Tasks: manifest.TaskList{
								manifest.Run{},
								manifest.Run{RestoreArtifacts: true},
							},
						},
					},
				},
			},
		}

		result := taskLinter.Lint(man)

		assert.Len(t, result.Errors, 0)
		assert.Len(t, result.Warnings, 0)
	})

	t.Run("a previous steps in the sequence saves artifacts", func(t *testing.T) {
		taskLinter := taskLinter{Fs: afero.Afero{},
			lintRunTask:   func(task manifest.Run, fs afero.Afero, os string) (errs []error, warnings []error) { return },
			lintParallel:  func(parallelTask manifest.Parallel) (errs []error, warnings []error) { return },
			lintSeq:       func(seqTask manifest.Seq, cameFromAParallel bool) (errs []error, warnings []error) { return },
			lintArtifacts: tasks.LintArtifacts,
		}

		man := manifest.Manifest{
			Tasks: []manifest.Task{
				manifest.Run{},
				manifest.Parallel{
					Tasks: manifest.TaskList{
						manifest.Seq{
							Tasks: manifest.TaskList{
								manifest.Run{SaveArtifacts: []string{"something"}},
								manifest.Run{RestoreArtifacts: true},
							},
						},
					},
				},
			},
		}

		result := taskLinter.Lint(man)

		assert.Len(t, result.Errors, 0)
		assert.Len(t, result.Warnings, 0)
	})
}

func TestLintArtifactsWithPrePromote(t *testing.T) {

	t.Run("A previous non pre promote step have saved artifact", func(t *testing.T) {
		taskLinter := taskLinter{Fs: afero.Afero{},
			lintRunTask:        func(task manifest.Run, fs afero.Afero, os string) (errs []error, warnings []error) { return },
			lintDeployCFTask:   func(task manifest.DeployCF, fs afero.Afero) (errs []error, warnings []error) { return },
			LintPrePromoteTask: func(tasks manifest.Task) (errs []error, warnings []error) { return },
			lintArtifacts:      tasks.LintArtifacts,
		}
		man := manifest.Manifest{
			Tasks: []manifest.Task{
				manifest.Run{SaveArtifacts: []string{"."}},
				manifest.DeployCF{
					PrePromote: []manifest.Task{
						manifest.Run{},
						manifest.Run{},
						manifest.Run{RestoreArtifacts: true},
					},
				},
				manifest.Run{},
			},
		}

		result := taskLinter.Lint(man)

		assert.Len(t, result.Errors, 0)
		assert.Len(t, result.Warnings, 0)

	})

	t.Run("A previous step haven't saved artifacts and the deploy uses a generated manifest manifest path", func(t *testing.T) {
		taskLinter := taskLinter{Fs: afero.Afero{},
			lintRunTask:        func(task manifest.Run, fs afero.Afero, os string) (errs []error, warnings []error) { return },
			lintDeployCFTask:   func(task manifest.DeployCF, fs afero.Afero) (errs []error, warnings []error) { return },
			LintPrePromoteTask: func(tasks manifest.Task) (errs []error, warnings []error) { return },
			lintArtifacts:      tasks.LintArtifacts,
		}
		man := manifest.Manifest{
			Tasks: []manifest.Task{
				manifest.Run{},
				manifest.DeployCF{Manifest: "../artifacts/some/path/manifest.yml"},
			},
		}

		result := taskLinter.Lint(man)

		assert.Len(t, result.Errors, 1)
		assert.Len(t, result.Warnings, 0)
		assert.Equal(t, "tasks[1] reads from saved artifacts, but there are no previous tasks that saves any", result.Errors[0].Error())
	})

	t.Run("A previous step have saved artifacts and the deploy uses a generated manifest manifest path", func(t *testing.T) {
		taskLinter := taskLinter{Fs: afero.Afero{},
			lintRunTask:        func(task manifest.Run, fs afero.Afero, os string) (errs []error, warnings []error) { return },
			lintDeployCFTask:   func(task manifest.DeployCF, fs afero.Afero) (errs []error, warnings []error) { return },
			LintPrePromoteTask: func(tasks manifest.Task) (errs []error, warnings []error) { return },
			lintArtifacts:      tasks.LintArtifacts,
		}
		man := manifest.Manifest{
			Tasks: []manifest.Task{
				manifest.Run{SaveArtifacts: []string{"."}},
				manifest.DeployCF{Manifest: "../artifacts/some/path/manifest.yml"},
			},
		}

		result := taskLinter.Lint(man)

		assert.Len(t, result.Errors, 0)
		assert.Len(t, result.Warnings, 0)
	})

}

func TestLintTimeout(t *testing.T) {
	taskLinter := taskLinter{
		lintRunTask:           func(task manifest.Run, fs afero.Afero, os string) (errs []error, warnings []error) { return },
		lintDeployCFTask:      func(task manifest.DeployCF, fs afero.Afero) (errs []error, warnings []error) { return },
		LintPrePromoteTask:    func(task manifest.Task) (errs []error, warnings []error) { return },
		lintDockerPushTask:    func(task manifest.DockerPush, fs afero.Afero) (errs []error, warnings []error) { return },
		lintDockerComposeTask: func(task manifest.DockerCompose, fs afero.Afero) (errs []error, warnings []error) { return },
		lintConsumerIntegrationTestTask: func(task manifest.ConsumerIntegrationTest, providerHostRequired bool) (errs []error, warnings []error) {
			return
		},
		lintDeployMLZipTask:     func(task manifest.DeployMLZip) (errs []error, warnings []error) { return },
		lintDeployMLModulesTask: func(task manifest.DeployMLModules) (errs []error, warnings []error) { return },
		lintArtifacts: func(currentTask manifest.Task, previousTasks []manifest.Task) (errs []error, warnings []error) {
			return
		},
	}

	badTime := "immaBadTime"

	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.Run{Timeout: badTime},
			manifest.DeployCF{
				PrePromote: []manifest.Task{
					manifest.Run{
						Timeout: badTime,
					},
				},
				Timeout: badTime,
			},
			manifest.DockerPush{
				Timeout: badTime,
			},
		},
	}

	result := taskLinter.Lint(man)

	assert.Len(t, result.Errors, 4)
	assert.Equal(t, "tasks[0] Invalid field 'timeout': time: invalid duration immaBadTime", result.Errors[0].Error())
	assert.Equal(t, "tasks[1] Invalid field 'timeout': time: invalid duration immaBadTime", result.Errors[1].Error())
	assert.Equal(t, "tasks[1].pre_promote[0] Invalid field 'timeout': time: invalid duration immaBadTime", result.Errors[2].Error())
	assert.Equal(t, "tasks[2] Invalid field 'timeout': time: invalid duration immaBadTime", result.Errors[3].Error())
}
