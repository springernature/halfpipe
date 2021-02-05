package mapper

import (
	"github.com/springernature/halfpipe/manifest"
)

type notificationsMapper struct {
}

func (n notificationsMapper) updateTasks(tasks manifest.TaskList, slackChannel string, slackSuccessMessage string, slackFailureMessage string) (updated manifest.TaskList) {
	for _, task := range tasks {
		switch task := task.(type) {
		case manifest.Parallel:
			task.Tasks = n.updateTasks(task.Tasks, slackChannel, slackSuccessMessage, slackFailureMessage)
			updated = append(updated, task)
		case manifest.Sequence:
			task.Tasks = n.updateTasks(task.Tasks, slackChannel, slackSuccessMessage, slackFailureMessage)
			updated = append(updated, task)
		default:
			if slackChannel != "" && !task.GetNotifications().NotificationsDefined() {
				notifications := manifest.Notifications{
					OnFailure: []string{slackChannel},
				}
				if task.NotifiesOnSuccess() {
					notifications.OnSuccess = []string{slackChannel}
				}
				task = task.SetNotifications(notifications)
			}
			task = updateMessages(task, slackSuccessMessage, slackFailureMessage)
			updated = append(updated, task)
		}
	}
	return updated
}

func updateMessages(task manifest.Task, slackSuccessMessage string, slackFailureMessage string) manifest.Task {
	notifications := task.GetNotifications()
	if notifications.OnSuccessMessage == "" {
		notifications.OnSuccessMessage = slackSuccessMessage
	}
	if notifications.OnFailureMessage == "" {
		notifications.OnFailureMessage = slackFailureMessage
	}
	return task.SetNotifications(notifications)
}

func (n notificationsMapper) Apply(man manifest.Manifest) (manifest.Manifest, error) {
	man.Tasks = n.updateTasks(man.Tasks, man.SlackChannel, man.SlackSuccessMessage, man.SlackFailureMessage)
	man.SlackChannel = ""
	man.SlackSuccessMessage = ""
	man.SlackFailureMessage = ""
	return man, nil
}

func NewNotificationsMapper() Mapper {
	return notificationsMapper{}
}
