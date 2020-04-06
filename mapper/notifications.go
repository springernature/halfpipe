package mapper

import "github.com/springernature/halfpipe/manifest"

type notificationsMapper struct {
}

func (n notificationsMapper) updateTasks(tasks manifest.TaskList, slackChannel string) (updated manifest.TaskList) {
	for _, task := range tasks {
		switch task := task.(type) {
		case manifest.Parallel:
			task.Tasks = n.updateTasks(task.Tasks, slackChannel)
			updated = append(updated, task)
		case manifest.Sequence:
			task.Tasks = n.updateTasks(task.Tasks, slackChannel)
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

			updated = append(updated, task)
		}
	}
	return updated
}

func (n notificationsMapper) Apply(original manifest.Manifest) (updated manifest.Manifest, err error) {
	updated = original

	if updated.SlackChannel != "" {
		updated.Tasks = n.updateTasks(updated.Tasks, updated.SlackChannel)
		updated.SlackChannel = ""
	}

	return updated, nil
}

func NewNotificationsMapper() Mapper {
	return notificationsMapper{}
}
