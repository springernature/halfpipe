package actions

import (
	"fmt"
	"github.com/springernature/halfpipe/manifest"
)

func notify(notifications manifest.Notifications) (steps Steps) {

	for _, channel := range notifications.Failure.Slack() {
		steps = append(steps, notifySlack(channel.Slack, channel.Message, false))
	}

	for idx, channel := range notifications.Failure.Teams() {
		steps = append(steps, notifyTeams(channel.Teams, channel.Message, false, idx, len(notifications.Failure.Teams())))
	}

	for _, channel := range notifications.Success.Slack() {
		steps = append(steps, notifySlack(channel.Slack, channel.Message, true))
	}

	for idx, channel := range notifications.Success.Teams() {
		steps = append(steps, notifyTeams(channel.Teams, channel.Message, true, idx, len(notifications.Success.Teams())))
	}

	return steps
}

func notifySlack(channel string, msg string, success bool) Step {
	if msg == "" {
		msg = "${{ job.status }} for pipeline ${{ github.workflow }} - link to the pipeline: ${{ github.server_url }}/${{ github.repository }}/actions/runs/${{ github.run_id }}"
	}

	step := Step{
		Name: "Notify slack " + channel,
		Uses: "slackapi/slack-github-action@v1.26.0",
		With: With{
			"channel-id":    channel,
			"slack-message": msg,
		},
		Env: Env{"SLACK_BOT_TOKEN": githubSecrets.SlackToken},
	}

	if success {
		step.Name += " (success)"
	} else {
		step.If = "failure()"
		step.Name += " (failure)"
	}

	return step
}

func notifyTeams(webhook string, msg string, success bool, idx int, count int) Step {
	var step Step
	if success {
		if msg == "" {
			msg = "✅ GitHub Actions workflow passed"
		}
		step = Step{
			Name: "Notify teams (success)",
			Uses: "jdcargile/ms-teams-notification@v1.4",
			With: With{
				"github-token":         "${{ github.token }}",
				"ms-teams-webhook-uri": webhook,
				"notification-color":   "28a745",
				"notification-summary": msg,
			},
		}
	} else {
		if msg == "" {
			msg = "❌ GitHub Actions workflow failed"
		}
		step = Step{
			Name: "Notify teams (failure)",
			Uses: "jdcargile/ms-teams-notification@v1.4",
			If:   "failure()",
			With: With{
				"github-token":         "${{ github.token }}",
				"ms-teams-webhook-uri": webhook,
				"notification-color":   "dc3545",
				"notification-summary": msg,
			},
		}
	}

	if count > 1 {
		step.Name = fmt.Sprintf("%s (%v)", step.Name, idx+1)
	}
	return step
}
