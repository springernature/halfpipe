package mapper

import "github.com/springernature/halfpipe/manifest"

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
			if !task.GetNotifications().NotificationsDefined() {
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
	if len(notifications.OnSuccess) > 0 && notifications.OnSuccessMessage == "" && slackSuccessMessage != "" {
		notifications.OnSuccessMessage = slackSuccessMessage
	}
	if notifications.OnFailureMessage == "" && slackFailureMessage != "" {
		notifications.OnFailureMessage = slackFailureMessage
	}
	task = task.SetNotifications(notifications)
	return task
}

func (n notificationsMapper) Apply(original manifest.Manifest) (updated manifest.Manifest, err error) {
	updated = original

	if updated.SlackChannel != "" {
		updated.Tasks = n.updateTasks(updated.Tasks, updated.SlackChannel, updated.SlackSuccessMessage, updated.SlackFailureMessage)
		updated.SlackChannel = ""
	}

	return updated, nil
}

func NewNotificationsMapper() Mapper {
	return notificationsMapper{}
}
