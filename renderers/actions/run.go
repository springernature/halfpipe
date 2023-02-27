package actions

import (
	"fmt"
	"github.com/springernature/halfpipe/manifest"
	"strings"
)

func (a *Actions) runSteps(task manifest.Run) (steps Steps) {
	run := Step{
		Name: task.GetName(),
		Env:  Env(task.Vars),
	}

	if task.Docker.Image != "" {
		prefix := ""
		if a.workingDir != "" {
			prefix = fmt.Sprintf("cd %s; ", a.workingDir)
		}
		if !strings.HasPrefix(task.Script, "./") && !strings.HasPrefix(task.Script, "/") && !strings.HasPrefix(task.Script, `\`) {
			task.Script = "./" + task.Script
		}
		task.Script = strings.Replace(task.Script, `"`, `\"`, -1)
		run.Uses = "docker://" + task.Docker.Image
		run.With = With{
			"entrypoint": WithOneLine{"/bin/sh"},
			"args":       WithOneLine{fmt.Sprintf(`-c "%s%s"`, prefix, task.Script)},
		}
	} else {
		run.Run = task.Script
	}

	steps = append(steps, dockerLogin(task.Docker.Image, task.Docker.Username, task.Docker.Password)...)
	steps = append(steps, run)

	if task.SavesArtifacts() {
		steps = append(steps, a.saveArtifacts(task.SaveArtifacts)...)
	}
	if task.SavesArtifactsOnFailure() {
		steps = append(steps, a.saveArtifactsOnFailure(task.SaveArtifactsOnFailure)...)
	}
	return steps
}
