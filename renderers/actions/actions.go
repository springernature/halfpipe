package actions

import (
	"time"

	"github.com/springernature/halfpipe/manifest"
)

const repoAccessToken = "${{ secrets.EE_REPO_ACCESS_TOKEN }}"
const slackToken = "${{ secrets.EE_SLACK_TOKEN }}"
const defaultRunner = "ubuntu-18.04"

type Actions struct{}

func NewActions() Actions {
	return Actions{}
}

func (a Actions) Render(man manifest.Manifest) (string, error) {
	w := Workflow{}
	w.Name = man.Pipeline
	w.On = a.triggers(man.Triggers)
	w.Jobs = a.jobs(man.Tasks, man)
	return w.asYAML()
}

func (a Actions) jobs(tasks manifest.TaskList, man manifest.Manifest) (jobs Jobs) {
	appendJob := func(job Job, notifications manifest.Notifications) {
		if notifications.NotificationsDefined() {
			job.Steps = append(job.Steps, notify(notifications)...)
		}
		jobs = append(jobs, Jobs{{Key: job.ID(), Value: job}}[0])
	}

	for _, t := range tasks {
		switch task := t.(type) {
		case manifest.DockerPush:
			appendJob(a.dockerPushJob(task, man), task.Notifications)
		case manifest.Run:
			appendJob(a.runJob(task, man), task.Notifications)
		}
	}
	return jobs
}

var checkoutCode = Step{
	Name: "Checkout code",
	Uses: "actions/checkout@v2",
}

func timeoutInMinutes(timeout string) int {
	d, err := time.ParseDuration(timeout)
	if err != nil {
		return 60
	}
	return int(d.Minutes())
}

func notify(notifications manifest.Notifications) []Step {
	var steps []Step

	s := func(channel string, text string) Step {
		return Step{
			Name: "Notify slack " + channel,
			Uses: "yukin01/slack-bot-action@v0.0.4",
			With: With{
				{Key: "status", Value: "${{ job.status }}"},
				{Key: "oauth_token", Value: slackToken},
				{Key: "channel", Value: channel},
				{Key: "text", Value: text},
			},
		}
	}

	for _, channel := range notifications.OnFailure {
		step := s(channel, notifications.OnFailureMessage)
		step.If = "failure()"
		step.Name += " (failure)"
		steps = append(steps, step)
	}

	for _, channel := range notifications.OnSuccess {
		step := s(channel, notifications.OnSuccessMessage)
		step.Name += " (success)"
		steps = append(steps, step)
	}

	return steps
}
