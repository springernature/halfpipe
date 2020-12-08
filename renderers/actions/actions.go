package actions

import (
	"regexp"
	"strings"
	"time"

	"github.com/springernature/halfpipe/manifest"
)

const repoAccessToken = "${{ secrets.EE_REPO_ACCESS_TOKEN }}"
const slackToken = "${{ secrets.EE_SLACK_TOKEN }}"
const defaultRunner = "ubuntu-20.04"

var globalEnv = Env{
	"ARTIFACTORY_PASSWORD": "${{ secrets.EE_ARTIFACTORY_PASSWORD }}",
	"ARTIFACTORY_URL":      "${{ secrets.EE_ARTIFACTORY_URL }}",
	"ARTIFACTORY_USERNAME": "${{ secrets.EE_ARTIFACTORY_USERNAME }}",
}

type Actions struct{}

func NewActions() Actions {
	return Actions{}
}

func (a Actions) Render(man manifest.Manifest) (string, error) {
	w := Workflow{}
	w.Name = man.Pipeline
	w.On = a.triggers(man.Triggers)
	if len(man.Tasks) > 0 {
		w.Env = globalEnv
		w.Jobs = a.jobs(man.Tasks, man)
	}
	return w.asYAML()
}

func (a Actions) jobs(tasks manifest.TaskList, man manifest.Manifest) (jobs Jobs) {
	appendJob := func(job Job, task manifest.Task, needs []string) {
		if task.GetNotifications().NotificationsDefined() {
			job.Steps = append(job.Steps, notify(task.GetNotifications())...)
		}
		job.TimeoutMinutes = timeoutInMinutes(task.GetTimeout())
		job.Needs = needs
		jobs = append(jobs, Jobs{{Key: idFromName(job.Name), Value: job}}[0])
	}

	for i, t := range tasks {
		needs := idsFromNames(tasks.PreviousTaskNames(i))
		switch task := t.(type) {
		case manifest.DockerPush:
			appendJob(a.dockerPushJob(task, man), task, needs)
		case manifest.Run:
			appendJob(a.runJob(task, man), task, needs)
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

func idFromName(name string) string {
	re := regexp.MustCompile(`[^a-z_0-9\-]`)
	return re.ReplaceAllString(strings.ToLower(name), "_")
}

func idsFromNames(names []string) []string {
	for i, n := range names {
		names[i] = idFromName(n)
	}
	return names
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
