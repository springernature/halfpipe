package tasks

import (
	"github.com/springernature/halfpipe/helpers"
	"github.com/springernature/halfpipe/manifest"
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
