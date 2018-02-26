package linters

import (
	"fmt"

	"regexp"

	"strings"

	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/errors"
	"github.com/springernature/halfpipe/helpers/file_checker"
	"github.com/springernature/halfpipe/model"
)

type TaskLinter struct {
	Fs afero.Afero
}

func (linter TaskLinter) Lint(man model.Manifest) (result model.LintResult) {
	result.Linter = "Tasks Linter"

	if len(man.Tasks) == 0 {
		result.AddError(errors.NewMissingField("tasks"))
		return
	}

	for i, t := range man.Tasks {
		switch task := t.(type) {
		case model.Run:
			result.AddError(linter.lintRunTask(task)...)
		case model.DeployCF:
			result.AddError(linter.lintDeployCFTask(task)...)
		case model.DockerPush:
			result.AddError(linter.lintDockerPushTask(task)...)
		default:
			result.AddError(errors.NewInvalidField("task", fmt.Sprintf("task %v is not a known task", i+1)))
		}
	}

	result.AddError(linter.lintArtifact(man)...)

	return
}
func (linter TaskLinter) lintDeployCFTask(cf model.DeployCF) (errs []error) {
	if cf.Api == "" {
		errs = append(errs, errors.NewMissingField("deploy-cf.api"))
	}
	if cf.Space == "" {
		errs = append(errs, errors.NewMissingField("deploy-cf.space"))
	}
	if cf.Org == "" {
		errs = append(errs, errors.NewMissingField("deploy-cf.org"))
	}
	if err := file_checker.CheckFile(linter.Fs, cf.Manifest, false); err != nil {
		errs = append(errs, err)
	}

	errs = append(errs, linter.lintEnvVars(cf.Vars)...)

	return
}

func (linter TaskLinter) lintDockerPushTask(docker model.DockerPush) (errs []error) {
	if docker.Username == "" {
		errs = append(errs, errors.NewMissingField("docker-push.username"))
	}
	if docker.Password == "" {
		errs = append(errs, errors.NewMissingField("docker-push.password"))
	}
	if docker.Repo == "" {
		errs = append(errs, errors.NewMissingField("docker-push.repo"))
	} else {
		matched, _ := regexp.Match(`^(.*)/(.*)$`, []byte(docker.Repo))
		if !matched {
			errs = append(errs, errors.NewInvalidField("docker-push.repo", "must be specified as 'owner/image'"))
		}
	}

	if err := file_checker.CheckFile(linter.Fs, "Dockerfile", false); err != nil {
		errs = append(errs, err)
	}

	errs = append(errs, linter.lintEnvVars(docker.Vars)...)

	return
}

func (linter TaskLinter) lintRunTask(run model.Run) []error {
	var errs []error
	if run.Script == "" {
		errs = append(errs, errors.NewMissingField("run.script"))
	} else {
		if err := file_checker.CheckFile(linter.Fs, run.Script, true); err != nil {
			errs = append(errs, err)
		}
	}

	if run.Image == "" {
		errs = append(errs, errors.NewMissingField("run.image"))
	}

	errs = append(errs, linter.lintEnvVars(run.Vars)...)

	return errs
}

func (linter TaskLinter) lintEnvVars(vars map[string]string) (errs []error) {
	for key, _ := range vars {
		if key != strings.ToUpper(key) {
			errs = append(errs, errors.NewInvalidField(key, "Env vars mus be uppercase only"))
		}
	}
	return
}

func (TaskLinter) lintArtifact(manifest model.Manifest) (errs []error) {
	var numberOfSaves int
	for _, t := range manifest.Tasks {
		switch task := t.(type) {
		case model.Run:
			if task.SaveArtifact != "" {
				numberOfSaves += 1
			}
		}
	}

	if numberOfSaves > 1 {
		errs = append(errs, errors.NewInvalidField("run.save_artifact", "Currently halfpipe only supports saving of one artifact per pipeline."))
	}

	return
}
