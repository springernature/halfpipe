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
	w := Workflow{}
	w.Name = man.Pipeline
	w.On = a.onTriggers(man.Triggers)
	w.Jobs = a.jobs(man.Tasks, man)
	return w.asYAML()
}

func (a Actions) onTriggers(triggers manifest.TriggerList) (on On) {
	for _, t := range triggers {
		switch trigger := t.(type) {
		case manifest.GitTrigger:
			on.Push = a.onPush(trigger)
		case manifest.TimerTrigger:
			on.Schedule = a.onSchedule(trigger)
		case manifest.DockerTrigger:
			on.RepositoryDispatch = a.onRepositoryDispatch(trigger.Image)
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

func (a Actions) onRepositoryDispatch(name string) RepositoryDispatch {
	return RepositoryDispatch{
		Types: []string{"docker-push:" + name},
	}
}

func (a Actions) jobs(tasks manifest.TaskList, man manifest.Manifest) (jobs Jobs) {
	appendJob := func(job Job) {
		jobs = append(jobs, yaml.MapItem{Key: job.ID(), Value: job})
	}

	for _, t := range tasks {
		switch task := t.(type) {
		case manifest.DockerPush:
			appendJob(a.dockerPushJob(task, man))
		}
	}
	return jobs
}

func (a Actions) dockerPushJob(task manifest.DockerPush, man manifest.Manifest) Job {
	basePath := man.Triggers.GetGitTrigger().BasePath
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
			repositoryDispatch(man.PipelineName()),
		},
	}
}

var checkoutCode = Step{
	Name: "Checkout code",
	Uses: "actions/checkout@v2",
}

func repositoryDispatch(name string) Step {
	return Step{
		Name: "Repository dispatch",
		Uses: "peter-evans/repository-dispatch@v1",
		With: []yaml.MapItem{
			{Key: "token", Value: repoAccessToken},
			{Key: "event-type", Value: "docker-push:" + name},
		},
	}

}

func timeoutMinutes(timeout string) int {
	d, err := time.ParseDuration(timeout)
	if err != nil {
		return 60
	}
	return int(d.Minutes())
}
