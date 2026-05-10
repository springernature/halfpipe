package actions

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/springernature/halfpipe/config"
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
		if success {
			msg = "✅ "
		} else {
			msg = "❌ "
		}
		msg += "workflow ${{ job.status }} `${{ github.workflow }}` ${{ github.server_url }}/${{ github.repository }}/actions/runs/${{ github.run_id }}"
	}

	step := Step{
		Name: "Notify slack " + channel,
		Uses: ExternalActions.Slack.Ref,
		With: With{
			"method": "chat.postMessage",
			"token":  config.GitHubSecrets.SlackToken,
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
	var style string

	if success {
		name = "Notify teams (success)"
		style = "good"
		if msg == "" {
			msg = "✅ GitHub Actions workflow passed"
		}
	} else {
		name = "Notify teams (failure)"
		style = "attention"
		if msg == "" {
			msg = "❌ GitHub Actions workflow failed"
		}
	}

	with := With{
		"style":   style,
		"summary": msg,
	}

	// Parse platform API URLs to use native action inputs
	if parsed, err := url.Parse(webhook); err == nil && strings.HasPrefix(webhook, config.PlatformAPIMessageURL) {
		if team := parsed.Query().Get("team"); team != "" {
			with["platform-team"] = team
		} else if channelID := parsed.Query().Get("channelID"); channelID != "" {
			with["channel-id"] = channelID
		} else {
			with["webhook-url"] = webhook
		}
	} else {
		with["webhook-url"] = webhook
	}

	step := Step{
		Name: name,
		Uses: ExternalActions.Teams.Ref,
		With: with,
	}

	if !success {
		step.If = "failure()"
	}

	if count > 1 {
		step.Name = fmt.Sprintf("%s (%v)", step.Name, idx+1)
	}
	return step
}
