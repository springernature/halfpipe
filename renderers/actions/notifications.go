package actions

import (
	"fmt"
	"github.com/springernature/halfpipe/manifest"
)

func notify(notifications manifest.Notifications) (steps Steps) {

	for _, channel := range notifications.Slack.OnFailure {
		steps = append(steps, notifySlack(channel, notifications.Slack.OnFailureMessage, false))
	}

	for idx, webhook := range notifications.Teams.OnFailure {
		steps = append(steps, notifyTeams(webhook, notifications.Teams.OnFailureMessage, false, idx, len(notifications.Teams.OnFailure)))
	}

	for _, channel := range notifications.Slack.OnSuccess {
		steps = append(steps, notifySlack(channel, notifications.Slack.OnSuccessMessage, true))
	}

	for idx, webhook := range notifications.Teams.OnSuccess {
		steps = append(steps, notifyTeams(webhook, notifications.Teams.OnSuccessMessage, true, idx, len(notifications.Teams.OnSuccess)))
	}

	return steps
}

func notifySlack(channel string, msg string, success bool) Step {
	if msg == "" {
		msg = "${{ job.status }} for pipeline ${{ github.workflow }} - link to the pipeline: ${{ github.server_url }}/${{ github.repository }}/actions/runs/${{ github.run_id }}"
	}

	step := Step{
		Name: "Notify slack " + channel,
		Uses: "slackapi/slack-github-action@v1.25.0",
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
