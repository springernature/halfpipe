package actions

import (
	"github.com/springernature/halfpipe/manifest"
	"gopkg.in/yaml.v2"
	"strings"
)

type Renderer interface {
	Render(manifest manifest.Manifest) Actions
}

type pipeline struct {
}

func NewRenderer() pipeline {
	return pipeline{}
}

func (p pipeline) Render(manifest manifest.Manifest) (actions Actions) {
	actions.Name = manifest.Pipeline
	actions.On = p.triggers(manifest.Triggers)
	actions.Jobs = p.jobs(manifest.Tasks)
	return
}

func (p pipeline) triggers(triggers []manifest.Trigger) (on On) {
	for _, trigger := range triggers {
		switch trigger := trigger.(type) {
		case manifest.GitTrigger:
			on.Push.Branches = []string{trigger.Branch}
			on.Push.Paths = trigger.WatchedPaths
		}
	}
	return
}

func (p pipeline) jobs(tasks manifest.TaskList) (jobs yaml.MapSlice) {
	for _, task := range tasks {
		switch task := task.(type) {
		case manifest.Run:
			job := Job{
				Name:   task.Name,
				RunsOn: "ubuntu-latest",
				Container: Container{
					Image: task.Docker.Image,
				},
				Steps: []Step{
					{Name: "Checkout code", Uses: "actions/checkout@v2"},
				},
			}

			runStep := Step{
				Name: "Run",
				Run:  task.Script,
			}
			job.Steps = append(job.Steps, runStep)

			jobId := strings.Replace(strings.ToLower(task.Name), " ", "-", -1)
			jobs = append(jobs, yaml.MapItem{
				Key:   jobId,
				Value: job,
			})
		}
	}
	return jobs
}
