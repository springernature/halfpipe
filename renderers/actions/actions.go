package actions

import (
	"github.com/springernature/halfpipe/manifest"
	"gopkg.in/yaml.v2"
	"path"
	"time"
)

type Actions struct{}

func NewActions() Actions {
	return Actions{}
}

func (a Actions) Render(man manifest.Manifest) (string, error) {
	basePath := man.Triggers.GetGitTrigger().BasePath
	w := Workflow{}
	w.Name = man.Pipeline
	w.On = a.onTriggers(man.Triggers)
	w.Jobs = a.jobs(man.Tasks, basePath)
	return w.asYAML()
}

func (a Actions) onTriggers(triggers manifest.TriggerList) (on On) {
	for _, t := range triggers {
		switch trigger := t.(type) {
		case manifest.GitTrigger:
			on.Push = a.onPush(trigger)
		case manifest.TimerTrigger:
			on.Schedule = a.onSchedule(trigger)
		}
	}
	return on
}

func (a Actions) onPush(git manifest.GitTrigger) (push Push) {
	if git.ManualTrigger {
		return push
	}

	push.Branches = Branches{git.Branch}

	for _, p := range git.WatchedPaths {
		push.Paths = append(push.Paths, p+"**")
	}
	for _, p := range git.IgnoredPaths {
		push.Paths = append(push.Paths, "!"+p+"**")
	}

	return push
}

func (a Actions) onSchedule(timer manifest.TimerTrigger) []Cron {
	return []Cron{{timer.Cron}}
}

func (a Actions) jobs(tasks manifest.TaskList, basePath string) (jobs Jobs) {
	appendJob := func(job Job) {
		jobs = append(jobs, yaml.MapItem{Key: job.ID(), Value: job})
	}

	for _, t := range tasks {
		switch task := t.(type) {
		case manifest.DockerPush:
			appendJob(a.dockerPushJob(task, basePath))
		}
	}
	return jobs
}

func (a Actions) dockerPushJob(task manifest.DockerPush, basePath string) Job {
	return Job{
		Name:           task.GetName(),
		RunsOn:         "ubuntu-18.04",
		TimeoutMinutes: timeoutMinutes(task.GetTimeout()),
		Steps: []Step{
			checkoutCode,
			{
				Name: "Set up Docker Buildx",
				Uses: "docker/setup-buildx-action@v1",
			},
			{
				Name: "Login to registry",
				Uses: "docker/login-action@v1",
				With: []yaml.MapItem{
					{Key: "registry", Value: "eu.gcr.io"},
					{Key: "username", Value: task.Username},
					{Key: "password", Value: secretMapper(task.Password)},
				},
			},
			{
				Name: "Build and push",
				Uses: "docker/build-push-action@v2",
				With: []yaml.MapItem{
					{Key: "context", Value: path.Join(basePath, task.BuildPath)},
					{Key: "file", Value: path.Join(basePath, task.DockerfilePath)},
					{Key: "push", Value: true},
					{Key: "tags", Value: task.Image},
					{Key: "outputs", Value: "type=image,oci-mediatypes=true,push=true"},
				},
			},
		},
	}
}

var checkoutCode = Step{
	Name: "Checkout code",
	Uses: "actions/checkout@v2",
}

func timeoutMinutes(timeout string) int {
	d, err := time.ParseDuration(timeout)
	if err != nil {
		return 60
	}
	return int(d.Minutes())
}
