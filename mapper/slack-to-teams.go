package mapper

import (
	"slices"

	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
)

type slackToTeamsMapper struct{}

func NewSlackToTeamsMapper() Mapper {
	return slackToTeamsMapper{}
}

func (s slackToTeamsMapper) Apply(man manifest.Manifest) (manifest.Manifest, error) {
	webhookURL := config.PlatformAPIMessageURL + "?team=" + man.Team

	man.Tasks = convertTasks(man.Tasks, webhookURL)

	return man, nil
}

func convertNotifications(n manifest.Notifications, webhookURL string) manifest.Notifications {
	n.Failure = convertChannels(n.Failure, webhookURL)
	n.Success = convertChannels(n.Success, webhookURL)
	return n
}

func convertChannels(channels manifest.NotificationChannels, webhookURL string) manifest.NotificationChannels {
	hasSlack := slices.ContainsFunc(channels, func(ch manifest.NotificationChannel) bool { return ch.Slack != "" })
	hasTeams := slices.ContainsFunc(channels, func(ch manifest.NotificationChannel) bool { return ch.Teams == webhookURL })

	if hasSlack && !hasTeams {
		return append(channels, manifest.NotificationChannel{Teams: webhookURL})
	}
	return channels
}

func convertTasks(tasks manifest.TaskList, webhookURL string) manifest.TaskList {
	if tasks == nil {
		return nil
	}
	updated := make(manifest.TaskList, len(tasks))
	for i, task := range tasks {
		switch t := task.(type) {
		case manifest.Parallel:
			t.Tasks = convertTasks(t.Tasks, webhookURL)
			updated[i] = t
		case manifest.Sequence:
			t.Tasks = convertTasks(t.Tasks, webhookURL)
			updated[i] = t
		default:
			updated[i] = task.SetNotifications(convertNotifications(task.GetBase().Notifications, webhookURL))
		}
	}
	return updated
}
