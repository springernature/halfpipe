package tasks

import (
	"github.com/pkg/errors"
	"github.com/springernature/halfpipe/manifest"
)

func LintArtifacts(currentTask manifest.Task, previousTasks []manifest.Task) (errs []error, warnings []error) {
	if currentTask.ReadsFromArtifacts() && !previousTasksSavesArtifact(previousTasks){
		errs = append(errs, errors.New("reads from saved artifacts, but there are no previous tasks that saves any"))
	}
	return
}

func previousTasksSavesArtifact(tasks []manifest.Task) bool {
	for _, task := range tasks {
		if task.SavesArtifacts() {
			return true
		}
	}
	return false
}

