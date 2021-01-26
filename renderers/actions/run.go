package actions

import (
	"fmt"
	"github.com/springernature/halfpipe/manifest"
	"strings"
)

func (a *Actions) runJob(task manifest.Run) Job {
	steps := []Step{checkoutCode}
	if task.ReadsFromArtifacts() {
		steps = append(steps, a.restoreArtifacts()...)
	}
	run := Step{
		Name: "run",
		Env:  Env(task.Vars),
	}

	if task.Docker.Image != "" {
		run.Uses = "docker://" + task.Docker.Image
		run.With = With{
			{"entrypoint", "/bin/sh"},
			{"args", fmt.Sprintf(`-c "%s"`, strings.Replace(task.Script, `"`, `\"`, -1))},
		}
	} else {
		run.Run = task.Script
	}

	if task.Docker.Username != "" {
		login := dockerLogin(task.Docker.Image, task.Docker.Username, task.Docker.Password)
		steps = append(steps, login)
	}
	steps = append(steps, run)

	if task.SavesArtifacts() {
		steps = append(steps, a.saveArtifacts(task.SaveArtifacts)...)
	}
	if task.SavesArtifactsOnFailure() {
		steps = append(steps, a.saveArtifactsOnFailure(task.SaveArtifactsOnFailure)...)
	}

	return Job{
		Name:   task.GetName(),
		RunsOn: defaultRunner,
		Steps:  steps,
	}
}
