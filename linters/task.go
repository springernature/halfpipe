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
			linter.lintRunTask(task, &result)
		case manifest.DeployCF:
			linter.lintDeployCFTask(task, &result)
		case manifest.DockerPush:
			linter.lintDockerPushTask(task, &result)
		case manifest.DockerCompose:
			linter.lintDockerComposeTask(task, &result)
		default:
			result.AddError(errors.NewInvalidField("task", fmt.Sprintf("task %v is not a known task", i+1)))
		}
	}
	return
}
func (linter taskLinter) lintDeployCFTask(cf manifest.DeployCF, result *LintResult) {
	if cf.API == "" {
		result.AddError(errors.NewMissingField("deploy-cf.api"))
	}
	if cf.Space == "" {
		result.AddError(errors.NewMissingField("deploy-cf.space"))
	}
	if cf.Org == "" {
		result.AddError(errors.NewMissingField("deploy-cf.org"))
	}
	if err := filechecker.CheckFile(linter.Fs, cf.Manifest, false); err != nil {
		result.AddError(err)
	}

	linter.lintEnvVars(cf.Vars, result)
	return
}

func (linter taskLinter) lintDockerPushTask(docker manifest.DockerPush, result *LintResult) {
	if docker.Username == "" {
		result.AddError(errors.NewMissingField("docker-push.username"))
	}
	if docker.Password == "" {
		result.AddError(errors.NewMissingField("docker-push.password"))
	}
	if docker.Image == "" {
		result.AddError(errors.NewMissingField("docker-push.image"))
	} else {
		matched, _ := regexp.Match(`^(.*)/(.*)$`, []byte(docker.Image))
		if !matched {
			result.AddError(errors.NewInvalidField("docker-push.image", "must be specified as 'user/image' or 'registry/user/image'"))
		}
	}

	if err := filechecker.CheckFile(linter.Fs, "Dockerfile", false); err != nil {
		result.AddError(err)
	}

	linter.lintEnvVars(docker.Vars, result)
	return
}

func (linter taskLinter) lintRunTask(run manifest.Run, result *LintResult) {
	if run.Script == "" {
		result.AddError(errors.NewMissingField("run.script"))
	} else {
		// Possible for script to have args,
		fields := strings.Fields(strings.TrimSpace(run.Script))
		command := fields[0]
		if err := filechecker.CheckFile(linter.Fs, command, true); err != nil {
			result.AddWarning(err)
		}
	}

	if run.Docker.Image == "" {
		result.AddError(errors.NewMissingField("run.docker.image"))
	}

	if run.Docker.Username != "" && run.Docker.Password == "" {
		result.AddError(errors.NewMissingField("run.docker.password"))
	}
	if run.Docker.Password != "" && run.Docker.Username == "" {
		result.AddError(errors.NewMissingField("run.docker.username"))
	}

	linter.lintEnvVars(run.Vars, result)
	return
}

func (linter taskLinter) lintDockerComposeTask(dc manifest.DockerCompose, result *LintResult) {
	if err := filechecker.CheckFile(linter.Fs, "docker-compose.yml", false); err != nil {
		result.AddError(err)
	}
	linter.lintEnvVars(dc.Vars, result)
	return
}

func (linter taskLinter) lintEnvVars(vars map[string]string, result *LintResult) {
	for key := range vars {
		if key != strings.ToUpper(key) {
			result.AddError(errors.NewInvalidField(key, "vars must be uppercase"))
		}
	}
	return
}
