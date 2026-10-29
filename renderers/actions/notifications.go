package actions

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
)

func notify(notifications manifest.Notifications) (steps Steps) {
	for idx, channel := range notifications.Failure.Teams() {
		steps = append(steps, notifyTeams(channel.Teams, channel.Message, false, idx, len(notifications.Failure.Teams())))
	}

	for idx, channel := range notifications.Success.Teams() {
		steps = append(steps, notifyTeams(channel.Teams, channel.Message, true, idx, len(notifications.Success.Teams())))
	}

	return steps
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
