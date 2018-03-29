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
	result.DocsURL = "https://docs.halfpipe.io/docs/manifest/#tasks"

	if len(man.Tasks) == 0 {
		result.AddError(errors.NewMissingField("tasks"))
		return
	}

	var lintTasks func(string, []manifest.Task)
	lintTasks = func(listName string, tasks []manifest.Task) {
		for i, t := range tasks {
			taskID := fmt.Sprintf("%s[%v]", listName, i)
			switch task := t.(type) {
			case manifest.Run:
				linter.lintRunTask(task, taskID, &result)
			case manifest.DeployCF:
				linter.lintDeployCFTask(task, taskID, &result)
				lintTasks(fmt.Sprintf("%s.pre_promote", taskID), task.PrePromote)
			case manifest.DockerPush:
				linter.lintDockerPushTask(task, taskID, &result)
			case manifest.DockerCompose:
				linter.lintDockerComposeTask(task, taskID, &result)
			default:
				result.AddError(errors.NewInvalidField("task", fmt.Sprintf("%s is not a known task", taskID)))
			}
		}
	}
	lintTasks("tasks", man.Tasks)
	return
}

func (linter taskLinter) lintDeployCFTask(cf manifest.DeployCF, taskID string, result *LintResult) {
	if cf.API == "" {
		result.AddError(errors.NewMissingField(taskID + " deploy-cf.api"))
	}
	if cf.Space == "" {
		result.AddError(errors.NewMissingField(taskID + " deploy-cf.space"))
	}
	if cf.Org == "" {
		result.AddError(errors.NewMissingField(taskID + " deploy-cf.org"))
	}
	if err := filechecker.CheckFile(linter.Fs, cf.Manifest, false); err != nil {
		result.AddError(err)
	}

	linter.lintEnvVars(cf.Vars, taskID, result)
	return
}

func (linter taskLinter) lintDockerPushTask(docker manifest.DockerPush, taskID string, result *LintResult) {
	if docker.Username == "" {
		result.AddError(errors.NewMissingField(taskID + " docker-push.username"))
	}
	if docker.Password == "" {
		result.AddError(errors.NewMissingField(taskID + " docker-push.password"))
	}
	if docker.Image == "" {
		result.AddError(errors.NewMissingField(taskID + " docker-push.image"))
	} else {
		matched, _ := regexp.Match(`^(.*)/(.*)$`, []byte(docker.Image))
		if !matched {
			result.AddError(errors.NewInvalidField(taskID+" docker-push.image", "must be specified as 'user/image' or 'registry/user/image'"))
		}
	}

	if err := filechecker.CheckFile(linter.Fs, "Dockerfile", false); err != nil {
		result.AddError(err)
	}

	linter.lintEnvVars(docker.Vars, taskID, result)
	return
}

func (linter taskLinter) lintRunTask(run manifest.Run, taskID string, result *LintResult) {
	if run.Script == "" {
		result.AddError(errors.NewMissingField(taskID + " run.script"))
	} else {
		// Possible for script to have args,
		fields := strings.Fields(strings.TrimSpace(run.Script))
		command := fields[0]
		if err := filechecker.CheckFile(linter.Fs, command, true); err != nil {
			result.AddWarning(err)
		}
	}

	if run.Docker.Image == "" {
		result.AddError(errors.NewMissingField(taskID + " run.docker.image"))
	}

	if run.Docker.Username != "" && run.Docker.Password == "" {
		result.AddError(errors.NewMissingField(taskID + " run.docker.password"))
	}
	if run.Docker.Password != "" && run.Docker.Username == "" {
		result.AddError(errors.NewMissingField(taskID + " run.docker.username"))
	}

	linter.lintEnvVars(run.Vars, taskID, result)
	return
}

func (linter taskLinter) lintDockerComposeTask(dc manifest.DockerCompose, taskID string, result *LintResult) {
	if err := filechecker.CheckFile(linter.Fs, "docker-compose.yml", false); err != nil {
		result.AddError(err)
	}
	linter.lintEnvVars(dc.Vars, taskID, result)
	return
}

func (linter taskLinter) lintEnvVars(vars map[string]string, taskID string, result *LintResult) {
	for key := range vars {
		if key != strings.ToUpper(key) {
			result.AddError(errors.NewInvalidField(taskID+" "+key, "vars must be uppercase"))
		}
	}
	return
}
