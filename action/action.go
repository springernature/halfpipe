package action

import (
	"fmt"
	"github.com/springernature/halfpipe/manifest"
	"strings"
)

type Push struct {
	Branches    []string `json:"branches,omitempty"`
	Paths       []string `json:"paths,omitempty"`
	PathsIgnore []string `json:"paths-ignore,omitempty"`
}

type Cron struct {
	Cron string `json:"cron,omitempty"`
}

type Event struct {
	Push     Push   `json:"push,omitempty"`
	Schedule []Cron `json:"schedule,omitempty"`
}

type UsesStep struct {
	Uses string `json:"uses,omitempty"`
}

func (u UsesStep) DamnGolangsLackOfGenerics() {
	panic("implement me")
}

type RunStep struct {
	Name string `json:"name,omitempty"`
	Run  string `json:"run,omitempty"`
}

func (r RunStep) DamnGolangsLackOfGenerics() {
	panic("implement me")
}

type Step interface {
	DamnGolangsLackOfGenerics()
}

type Job struct {
	Name   string `json:"name,omitempty"`
	RunsOn string `json:"runs-on,omitempty"`
	Needs  string `json:"needs,omitempty"`
	Steps  []Step `json:"steps,omitempty"`
}

type Action struct {
	Name string         `json:"name"`
	On   Event          `json:"on"`
	Jobs map[string]Job `json:"jobs,omitempty"`
}

type action struct{}

func Renderer() action {
	return action{}
}

func (p action) on(man manifest.Manifest) (on Event) {
	for _, trigger := range man.Triggers {
		switch trigger := trigger.(type) {
		case manifest.GitTrigger:
			branch := man.Triggers.GetGitTrigger().Branch
			if branch == "" {
				branch = "master"
			}

			on.Push = Push{
				Branches:    []string{branch},
				Paths:       trigger.WatchedPaths,
				PathsIgnore: trigger.IgnoredPaths,
			}
		case manifest.TimerTrigger:
			on.Schedule = append(on.Schedule, Cron{Cron: trigger.Cron})
		}
	}

	return on
}

func (p action) Render(man manifest.Manifest) (cfg Action) {

	cfg.Name = man.Pipeline
	cfg.On = p.on(man)

	cfg.Jobs = make(map[string]Job)
	var previousTask string
	for _, task := range man.Tasks {
		name := strings.Replace(task.GetName(), " ", "_", -1)
		job := Job{
			Name:   name,
			RunsOn: "ubuntu-latest",
			Steps: []Step{
				UsesStep{Uses: "actions/checkout@v1"},
			},
		}

		switch task := task.(type) {
		case manifest.Run:
			job.Steps = append(job.Steps,
				RunStep{
					Name: name,
					Run:  task.Script,
				},
			)
		case manifest.DockerCompose:
			job.Steps = append(job.Steps,
				RunStep{
					Name: name,
					Run:  fmt.Sprintf("docker-compose run %s", task.Service),
				},
			)
		}
		if previousTask != "" {
			job.Needs = previousTask
		}

		previousTask = name
		cfg.Jobs[name] = job
	}

	return cfg
}
