package actions

import (
	"github.com/springernature/halfpipe/manifest"
)

func (a Actions) runJob(task manifest.Run) Job {
	steps := []Step{checkoutCode}
	if task.ReadsFromArtifacts() {
		steps = append(steps, restoreArtifacts)
	}
	steps = append(steps, Step{
		Name: "run",
		Run:  task.Script,
	})
	if task.SavesArtifacts() {
		steps = append(steps, saveArtifacts(task.SaveArtifacts))
	}
	if task.SavesArtifactsOnFailure() {
		steps = append(steps, saveArtifactsOnFailure(task.SaveArtifactsOnFailure))
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
