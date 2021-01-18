package actions

import (
	"github.com/springernature/halfpipe/manifest"
)

func (a *Actions) runJob(task manifest.Run) Job {
	steps := []Step{checkoutCode}
	if task.ReadsFromArtifacts() {
		steps = append(steps, a.restoreArtifacts())
	}
	steps = append(steps, Step{
		Name: "run",
		Run:  task.Script,
	})
	if task.SavesArtifacts() {
		steps = append(steps, a.saveArtifacts(task.SaveArtifacts))
	}
	if task.SavesArtifactsOnFailure() {
		steps = append(steps, a.saveArtifactsOnFailure(task.SaveArtifactsOnFailure))
	}

	return Job{
		Name:   task.GetName(),
		RunsOn: defaultRunner,
		Container: Container{
			Image: task.Docker.Image,
			Credentials: Credentials{
				Username: task.Docker.Username,
				Password: task.Docker.Password,
			},
		},
		Steps: steps,
		Env:   Env(task.Vars),
	}
}
