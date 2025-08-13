package linters

import (
	"testing"

	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestParallelTaskInParallelTask(t *testing.T) {
	task := manifest.Parallel{
		Tasks: manifest.TaskList{
			manifest.Run{},
			manifest.DockerPush{},
			manifest.Parallel{
				Tasks: manifest.TaskList{
					manifest.Run{},
				},
			},
		},
	}
	errs := LintParallelTask(task, "concourse")
	assertContainsError(t, errs, ErrInvalidField.WithValue("type"))
}

func TestErrorIfParallelIsEmpty(t *testing.T) {
	task := manifest.Parallel{
		Tasks: manifest.TaskList{},
	}
	errs := LintParallelTask(task, "concourse")
	assert.Len(t, errs, 1)
	assertContainsError(t, errs, ErrInvalidField.WithValue("tasks"))
}

func TestWarningIfParallelOnlyContainsOneItem(t *testing.T) {
	task := manifest.Parallel{
		Tasks: manifest.TaskList{
			manifest.Run{},
		},
	}
	errs := LintParallelTask(task, "concourse")
	assertContainsError(t, errs, ErrInvalidField.WithValue("tasks"))
}

func TestWarnIfMultipleTasksInsideParallelSavesArtifact(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		task := manifest.Parallel{
			Tasks: manifest.TaskList{
				manifest.Run{SaveArtifactsOnFailure: []string{"blah"}},
				manifest.Run{SaveArtifacts: []string{"."}},
				manifest.Run{},
			},
		}

		errs := LintParallelTask(task, "concourse")
		assert.Len(t, errs, 0)
	})

	t.Run("bad", func(t *testing.T) {
		task := manifest.Parallel{
			Tasks: manifest.TaskList{
				manifest.Run{},
				manifest.Run{SaveArtifacts: []string{"."}},
				manifest.Run{},
				manifest.Run{},
				manifest.Run{SaveArtifacts: []string{"."}},
			},
		}

		errs := LintParallelTask(task, "concourse")
		assertContainsError(t, errs, ErrInvalidField.WithValue("tasks"))

		assert.Len(t, LintParallelTask(task, "actions"), 0)
	})
}

func TestWarnIfMultipleTasksInsideParallelSavesArtifactOnFailure(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		task := manifest.Parallel{
			Tasks: manifest.TaskList{
				manifest.Run{},
				manifest.Run{SaveArtifacts: []string{"."}},
				manifest.Run{},
			},
		}

		errs := LintParallelTask(task, "concourse")
		assert.Len(t, errs, 0)
	})

	t.Run("bad", func(t *testing.T) {
		task := manifest.Parallel{
			Tasks: manifest.TaskList{
				manifest.Run{},
				manifest.Run{SaveArtifacts: []string{"."}, SaveArtifactsOnFailure: []string{"blurgh"}},
				manifest.Run{},
				manifest.Run{},
				manifest.Run{SaveArtifactsOnFailure: []string{"."}},
			},
		}

		errs := LintParallelTask(task, "concourse")
		assertContainsError(t, errs, ErrInvalidField.WithValue("tasks"))

		assert.Len(t, LintParallelTask(task, "actions"), 0)
	})
}
