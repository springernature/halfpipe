package actions

import "github.com/springernature/halfpipe/manifest"

func (a Actions) runJob(task manifest.Run, man manifest.Manifest) Job {
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
		Steps: []Step{
			{
				Name: "run",
				Run:  "echo not implemented yet",
			},
		},
		Env: Env(task.Vars),
	}
}
