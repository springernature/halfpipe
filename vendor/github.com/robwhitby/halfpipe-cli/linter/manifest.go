package linter

import (
	"strings"

	"fmt"

	"errors"

	. "github.com/robwhitby/halfpipe-cli/model"
)

func LintManifest(man Manifest) (errs []error) {

	if man.Team == "" {
		errs = append(errs, NewMissingField("team"))
	}

	if man.Repo.Uri == "" {
		errs = append(errs, NewMissingField("repo.uri"))
	} else if !strings.Contains(man.Repo.Uri, "github") {
		errs = append(errs, NewInvalidField("repo.uri", "must contain 'github'"))
	}

	if len(man.Tasks) == 0 {
		errs = append(errs, NewMissingField("tasks"))
	}

	for i, t := range man.Tasks {
		switch task := t.(type) {
		case Run:
			lintRunTask(task, i+1, &errs)
		case DeployCF:
			lintDeployCFTask(task, i+1, &errs)
		case DockerPush:
			lintDockerPushTask(task, i+1, &errs)
		default:
			errs = append(errs, NewInvalidField("task", fmt.Sprintf("task %v '%s' is not a known task", i+1, task.GetName())))
		}
	}

	return
}

func lintRunTask(task Run, taskNumber int, errs *[]error) {
	if task.Script == "" {
		*errs = append(*errs, NewMissingField(fmt.Sprintf("task %v: script", taskNumber)))
	}
	if task.Image == "" {
		*errs = append(*errs, NewMissingField(fmt.Sprintf("task %v: image", taskNumber)))
	}
}

func lintDeployCFTask(task DeployCF, taskNumber int, errs *[]error) {
	*errs = append(*errs, errors.New("todo: lint deploy-cf task"))
	return
}

func lintDockerPushTask(task DockerPush, taskNumber int, errs *[]error) {
	*errs = append(*errs, errors.New("todo: lint docker-push task"))
	return
}
