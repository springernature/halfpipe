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
			not := manifest.NotificationChannel{Slack: man.SlackChannel}
			if man.SlackFailureMessage != "" {
				not.Message = man.SlackFailureMessage
			}
			notifications.Failure = append(notifications.Failure, not)
		}

		if man.TeamsWebhook != "" {
			notifications.Failure = append(notifications.Failure, manifest.NotificationChannel{
				Teams: man.TeamsWebhook,
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
			if len(taskNotifications.Success) == 0 && len(taskNotifications.Failure) == 0 {
				for _, n := range taskNotifications.OnFailure {
					not := manifest.NotificationChannel{Slack: n}
					if taskNotifications.OnFailureMessage != "" {
						not.Message = taskNotifications.OnFailureMessage
					}
					taskNotifications.Failure = append(taskNotifications.Failure, not)
				}

				for _, n := range taskNotifications.OnSuccess {
					not := manifest.NotificationChannel{Slack: n}
					if taskNotifications.OnSuccessMessage != "" {
						not.Message = taskNotifications.OnSuccessMessage
					}
					taskNotifications.Success = append(taskNotifications.Success, not)
				}
			}

			taskNotifications.OnFailure = nil
			taskNotifications.OnFailureMessage = ""
			taskNotifications.OnSuccess = nil
			taskNotifications.OnSuccessMessage = ""

			updated = append(updated, task.SetNotifications(taskNotifications))
		}
	}

	return updated
}

func (n notificationsMapper) updateTasks(tasks manifest.TaskList, slackSuccessMessage string, topLevelNotifications manifest.Notifications) (updated manifest.TaskList) {
	for _, task := range tasks {
		switch task := task.(type) {
		case manifest.Parallel:
			task.Tasks = n.updateTasks(task.Tasks, slackSuccessMessage, topLevelNotifications)
			updated = append(updated, task)
		case manifest.Sequence:
			task.Tasks = n.updateTasks(task.Tasks, slackSuccessMessage, topLevelNotifications)
			updated = append(updated, task)
		default:
			if task.GetNotifications().Equal(manifest.Notifications{}) {
				if topLevelNotifications.NotificationsDefined() {
					task = task.SetNotifications(topLevelNotifications)
				}
				if task.NotifiesOnSuccess() {
					taskNotifications := task.GetNotifications()
					if len(taskNotifications.Failure.Slack()) > 0 {
						not := taskNotifications.Failure.Slack()[0]
						not.Message = slackSuccessMessage
						taskNotifications.Success = append(taskNotifications.Success, not)
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
	man.Tasks = n.updateTasks(man.Tasks, man.SlackSuccessMessage, man.Notifications)
	man.SlackChannel = ""
	man.SlackSuccessMessage = ""
	man.SlackFailureMessage = ""
	man.TeamsWebhook = ""
	return man, nil
}

func NewNotificationsMapper() Mapper {
	return notificationsMapper{}
}
