package mapper

import (
	"github.com/springernature/halfpipe/manifest"
)

type notificationsMapper struct {
}

func (n notificationsMapper) topLevelNotifications(man manifest.Manifest) manifest.Notifications {
	notifications := man.Notifications

	if notifications.Equal(manifest.Notifications{}) {
		if man.SlackChannel != "" {
			notifications.Slack.OnFailure = []string{man.SlackChannel}
		}

		if man.SlackFailureMessage != "" {
			notifications.Slack.OnFailureMessage = man.SlackFailureMessage
		}

		if man.SlackSuccessMessage != "" {
			notifications.Slack.OnSuccessMessage = man.SlackSuccessMessage
		}

		if len(notifications.Slack.OnFailure) == 1 {
			not := manifest.NotificationChannel{"slack": notifications.Slack.OnFailure[0]}
			if notifications.Slack.OnFailureMessage != "" {
				not["message"] = notifications.Slack.OnFailureMessage
			}
			notifications.Failure = append(notifications.Failure, not)
		}

		if man.TeamsWebhook != "" {
			notifications.Teams.OnFailure = []string{man.TeamsWebhook}
			notifications.Failure = append(notifications.Failure, manifest.NotificationChannel{
				"teams": man.TeamsWebhook,
			})
		}

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

			if taskNotifications.Slack.Equal(manifest.Channels{}) {
				taskNotifications.Slack = manifest.Channels{
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
			if task.GetNotifications().Slack.Equal(manifest.Channels{}) {
				if topLevelNotifications.NotificationsDefined() {
					task = task.SetNotifications(topLevelNotifications)
				}
				if task.NotifiesOnSuccess() {
					taskNotifications := task.GetNotifications()
					if len(taskNotifications.Slack.OnFailure) > 0 {
						taskNotifications.Slack.OnSuccess = []string{taskNotifications.Slack.OnFailure[0]}
						task = task.SetNotifications(taskNotifications)
					}
					if len(taskNotifications.Failure.Slack()) > 0 {
						taskNotifications.Success = append(taskNotifications.Success, taskNotifications.Failure.Slack()[0])
						task = task.SetNotifications(taskNotifications)
					}

					if len(taskNotifications.Teams.OnFailure) > 0 {
						taskNotifications.Teams.OnSuccess = []string{taskNotifications.Teams.OnFailure[0]}
						task = task.SetNotifications(taskNotifications)
					}

					if len(taskNotifications.Failure.Teams()) > 0 {
						taskNotifications.Success = append(taskNotifications.Success, taskNotifications.Failure.Teams()[0])
						task = task.SetNotifications(taskNotifications)
					}

				}
			}

			updated = append(updated, task.SetNotifyOnSuccess(false))
		}
	}
	return updated
}

func (n notificationsMapper) Apply(man manifest.Manifest) (manifest.Manifest, error) {
	man.Notifications = n.topLevelNotifications(man)
	man.Tasks = n.migrateTaskNotifications(man.Tasks)
	man.Tasks = n.updateTasks(man.Tasks, man.SlackChannel, man.SlackSuccessMessage, man.SlackFailureMessage, man.Notifications)
	man.SlackChannel = ""
	man.SlackSuccessMessage = ""
	man.SlackFailureMessage = ""
	man.TeamsWebhook = ""
	return man, nil
}

func NewNotificationsMapper() Mapper {
	return notificationsMapper{}
}
