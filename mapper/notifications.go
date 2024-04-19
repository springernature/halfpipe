package mapper

import (
	"github.com/springernature/halfpipe/manifest"
)

type notificationsMapper struct {
}

func (n notificationsMapper) topLevelNotifications(man manifest.Manifest) manifest.Notifications {
	notifications := man.Notifications
	if man.SlackChannel != "" && len(notifications.Slack.OnFailure) == 0 {
		notifications.Slack.OnFailure = []string{man.SlackChannel}
	}

	if man.SlackFailureMessage != "" && notifications.Slack.OnFailureMessage == "" {
		notifications.Slack.OnFailureMessage = man.SlackFailureMessage
	}

	if man.SlackSuccessMessage != "" && notifications.Slack.OnSuccessMessage == "" {
		notifications.Slack.OnSuccessMessage = man.SlackSuccessMessage
	}

	return notifications
}

func (n notificationsMapper) migrateTaskNotifications(tasks manifest.TaskList) (updated manifest.TaskList) {
	for _, task := range tasks {
		switch task := task.(type) {
		case manifest.Parallel:
			task.Tasks = n.migrateTaskNotifications(task.Tasks)
			updated = append(updated, task)
		case manifest.Sequence:
			task.Tasks = n.migrateTaskNotifications(task.Tasks)
			updated = append(updated, task)
		default:
			taskNotifications := task.GetNotifications()

			if taskNotifications.Slack.Equal(manifest.Slack{}) {
				taskNotifications.Slack = manifest.Slack{
					OnFailure:        taskNotifications.OnFailure,
					OnFailureMessage: taskNotifications.OnFailureMessage,
					OnSuccess:        taskNotifications.OnSuccess,
					OnSuccessMessage: taskNotifications.OnSuccessMessage,
				}
				taskNotifications.OnFailure = nil
				taskNotifications.OnFailureMessage = ""
				taskNotifications.OnSuccess = nil
				taskNotifications.OnSuccessMessage = ""
			}

			updated = append(updated, task.SetNotifications(taskNotifications))
		}
	}

	return updated
}

func (n notificationsMapper) updateTasks(tasks manifest.TaskList, slackChannel string, slackSuccessMessage string, slackFailureMessage string, topLevelNotifications manifest.Notifications) (updated manifest.TaskList) {
	for _, task := range tasks {
		switch task := task.(type) {
		case manifest.Parallel:
			task.Tasks = n.updateTasks(task.Tasks, slackChannel, slackSuccessMessage, slackFailureMessage, topLevelNotifications)
			updated = append(updated, task)
		case manifest.Sequence:
			task.Tasks = n.updateTasks(task.Tasks, slackChannel, slackSuccessMessage, slackFailureMessage, topLevelNotifications)
			updated = append(updated, task)
		default:
			if task.GetNotifications().Slack.Equal(manifest.Slack{}) {
				if topLevelNotifications.NotificationsDefined() {
					task = task.SetNotifications(topLevelNotifications)
				}
				if task.NotifiesOnSuccess() {
					taskNotifications := task.GetNotifications()
					if len(taskNotifications.Slack.OnFailure) > 0 {
						taskNotifications.Slack.OnSuccess = []string{taskNotifications.Slack.OnFailure[0]}
						task = task.SetNotifications(taskNotifications)
					}
				}
			}

			updated = append(updated, task.SetNotifyOnSuccess(false))
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
	man.Notifications = n.topLevelNotifications(man)
	man.Tasks = n.migrateTaskNotifications(man.Tasks)
	man.Tasks = n.updateTasks(man.Tasks, man.SlackChannel, man.SlackSuccessMessage, man.SlackFailureMessage, man.Notifications)
	man.SlackChannel = ""
	man.SlackSuccessMessage = ""
	man.SlackFailureMessage = ""
	return man, nil
}

func NewNotificationsMapper() Mapper {
	return notificationsMapper{}
}
