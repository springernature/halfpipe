package linters

import (
	"fmt"

	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/errors"
	"github.com/springernature/halfpipe/model"
	"regexp"
)

type TaskLinter struct {
	Fs afero.Afero
}

func (taskLinter TaskLinter) Lint(man model.Manifest) (result errors.LintResult) {
	result.Linter = "Tasks Linter"

	if len(man.Tasks) == 0 {
		result.Errors = append(result.Errors, errors.NewMissingField("tasks"))
		return
	}

	for i, t := range man.Tasks {
		switch task := t.(type) {
		case model.Run:
			result.Errors = append(result.Errors, lintRunTask(taskLinter, task)...)
		case model.DeployCF:
			result.Errors = append(result.Errors, lintDeployCFTask(task)...)
		case model.DockerPush:
			result.Errors = append(result.Errors, lintDockerPushTask(taskLinter, task)...)
		default:
			result.Errors = append(result.Errors, errors.NewInvalidField("task", fmt.Sprintf("task %v '%s' is not a known task", i+1, task.GetName())))
		}
	}

	return
}
func lintDeployCFTask(cf model.DeployCF) (errs []error) {
	if cf.Api == "" {
		errs = append(errs, errors.NewMissingField("api"))
	}

	if cf.Space == "" {
		errs = append(errs, errors.NewMissingField("space"))
	}

	if cf.Org == "" {
		errs = append(errs, errors.NewMissingField("org"))
	}
	return
}

func lintDockerPushTask(t TaskLinter, docker model.DockerPush) (errs []error) {
	if docker.Username == "" {
		errs = append(errs, errors.NewMissingField("username"))
	}
	if docker.Password == "" {
		errs = append(errs, errors.NewMissingField("password"))
	}
	if docker.Repo == "" {
		errs = append(errs, errors.NewMissingField("repo"))
	} else {
		matched, _ := regexp.Match(`^(.*)/(.*)$`, []byte(docker.Repo))
		if !matched {
			errs = append(errs, errors.NewInvalidField("repo", "must be specified as 'owner/image'"))
		}
	}

	if err := CheckFile(t.Fs, "Dockerfile", false); err != nil {
		errs = append(errs, err)
	}

	return
}

func lintRunTask(t TaskLinter, run model.Run) []error {
	var errs []error
	if run.Script == "" {
		errs = append(errs, errors.NewMissingField("script"))
	} else {
		if err := CheckFile(t.Fs, run.Script, true); err != nil {
			errs = append(errs, err)
		}
	}

	if run.Image == "" {
		errs = append(errs, errors.NewMissingField("image"))
	}

	return errs
}
