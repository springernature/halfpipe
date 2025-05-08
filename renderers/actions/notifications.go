package actions

import (
	"fmt"
	"github.com/springernature/halfpipe/manifest"
	"strings"
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
		if success {
			msg = "✅ "
		} else {
			msg = "❌ "
		}
		msg += "workflow ${{ job.status }} `${{ github.workflow }}` ${{ github.server_url }}/${{ github.repository }}/actions/runs/${{ github.run_id }}"
	}

	step := Step{
		Name: "Notify slack " + channel,
		Uses: "slackapi/slack-github-action@b0fa283ad8fea605de13dc3f449259339835fc52", //v2.0.0
		With: With{
			"method": "chat.postMessage",
			"token":  githubSecrets.SlackToken,
			"payload": fmt.Sprintf(`channel: "%s"
text: "%s"`, channel, strings.ReplaceAll(msg, `"`, `\"`)),
		},
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

	var name string
	var color string

	if success {
		name = "Notify teams (success)"
		color = "28a745"
		if msg == "" {
			msg = "✅ GitHub Actions workflow passed"
		}
	} else {
		name = "Notify teams (failure)"
		color = "dc3545"
		if msg == "" {
			msg = "❌ GitHub Actions workflow failed"
		}
	}

	step := Step{
		Name: name,
		Uses: "jdcargile/ms-teams-notification@28e5ca976c053d54e2b852f3f38da312f35a24fc", // v1.4
		With: With{
			"github-token":         "${{ github.token }}",
			"ms-teams-webhook-uri": webhook,
			"notification-color":   color,
			"notification-summary": msg,
		},
	}

	if !success {
		step.If = "failure()"
	}

	if count > 1 {
		step.Name = fmt.Sprintf("%s (%v)", step.Name, idx+1)
	}
	return step
}
