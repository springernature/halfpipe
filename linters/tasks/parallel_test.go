package tasks

import (
	"github.com/springernature/halfpipe/helpers"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
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
	errs, _ := LintParallelTask(task)
	helpers.AssertInvalidFieldInErrors(t, "type", errs)
}

func TestWarningIfTaskInsideParallelTaskIsDefined(t *testing.T) {
	task := manifest.Parallel{
		Tasks: manifest.TaskList{
			manifest.Run{},
			manifest.DockerPush{Parallel: "true"},
			manifest.Run{Parallel: "true"},
		},
	}

	errs, warns := LintParallelTask(task)
	assert.Len(t, errs, 0)
	assert.Len(t, warns, 2)
	helpers.AssertInvalidFieldInErrors(t, "parallel", warns)
}
