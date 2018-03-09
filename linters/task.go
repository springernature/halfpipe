package linters

import (
	"fmt"

	"regexp"

	"strings"

	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/linters/errors"
	"github.com/springernature/halfpipe/linters/filechecker"
	"github.com/springernature/halfpipe/manifest"
)

type taskLinter struct {
	Fs afero.Afero
}

func NewTasksLinter(fs afero.Afero) taskLinter {
	return taskLinter{fs}
}

func (linter taskLinter) Lint(man manifest.Manifest) (result LintResult) {
	result.Linter = "Tasks"

	if len(man.Tasks) == 0 {
		result.AddError(errors.NewMissingField("tasks"))
		return
	}

	for i, t := range man.Tasks {
		switch task := t.(type) {
		case manifest.Run:
			result.AddError(linter.lintRunTask(task)...)
		case manifest.DeployCF:
			result.AddError(linter.lintDeployCFTask(task)...)
		case manifest.DockerPush:
			result.AddError(linter.lintDockerPushTask(task)...)
		default:
			result.AddError(errors.NewInvalidField("task", fmt.Sprintf("task %v is not a known task", i+1)))
		}
	}

	return
}
func (linter taskLinter) lintDeployCFTask(cf manifest.DeployCF) (errs []error) {
	if cf.API == "" {
		errs = append(errs, errors.NewMissingField("deploy-cf.api"))
	}
	if cf.Space == "" {
		errs = append(errs, errors.NewMissingField("deploy-cf.space"))
	}
	if cf.Org == "" {
		errs = append(errs, errors.NewMissingField("deploy-cf.org"))
	}
	if err := filechecker.CheckFile(linter.Fs, cf.Manifest, false); err != nil {
		errs = append(errs, err)
	}

	errs = append(errs, linter.lintEnvVars(cf.Vars)...)

	return
}

func (linter taskLinter) lintDockerPushTask(docker manifest.DockerPush) (errs []error) {
	if docker.Username == "" {
		errs = append(errs, errors.NewMissingField("docker-push.username"))
	}
	if docker.Password == "" {
		errs = append(errs, errors.NewMissingField("docker-push.password"))
	}
	if docker.Image == "" {
		errs = append(errs, errors.NewMissingField("docker-push.image"))
	} else {
		matched, _ := regexp.Match(`^(.*)/(.*)$`, []byte(docker.Image))
		if !matched {
			errs = append(errs, errors.NewInvalidField("docker-push.image", "must be specified as 'user/image' or 'registry/user/image'"))
		}
	}

	if err := filechecker.CheckFile(linter.Fs, "Dockerfile", false); err != nil {
		errs = append(errs, err)
	}

	errs = append(errs, linter.lintEnvVars(docker.Vars)...)

	return
}

func (linter taskLinter) lintRunTask(run manifest.Run) []error {
	var errs []error
	if run.Script == "" {
		errs = append(errs, errors.NewMissingField("run.script"))
	} else {
		if err := filechecker.CheckFile(linter.Fs, run.Script, true); err != nil {
			errs = append(errs, err)
		}
	}

	if run.Docker.Image == "" {
		errs = append(errs, errors.NewMissingField("run.docker.image"))
	}

	if run.Docker.Username != "" && run.Docker.Password == "" {
		errs = append(errs, errors.NewMissingField("run.docker.password"))
	}
	if run.Docker.Password != "" && run.Docker.Username == "" {
		errs = append(errs, errors.NewMissingField("run.docker.username"))
	}

	errs = append(errs, linter.lintEnvVars(run.Vars)...)

	return errs
}

func (linter taskLinter) lintEnvVars(vars map[string]string) (errs []error) {
	for key := range vars {
		if key != strings.ToUpper(key) {
			errs = append(errs, errors.NewInvalidField(key, "Env vars mus be uppercase only"))
		}
	}
	return
}
