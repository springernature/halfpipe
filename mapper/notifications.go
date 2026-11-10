package mapper

import (
	"github.com/springernature/halfpipe/manifest"
)

type notificationsMapper struct {
}

func (n notificationsMapper) updateTasks(tasks manifest.TaskList, topLevelNotifications manifest.Notifications) (updated manifest.TaskList) {
	for _, task := range tasks {
		switch task := task.(type) {
		case manifest.Parallel:
			task.Tasks = n.updateTasks(task.Tasks, topLevelNotifications)
			updated = append(updated, task)
		case manifest.Sequence:
			task.Tasks = n.updateTasks(task.Tasks, topLevelNotifications)
			updated = append(updated, task)
		default:
			if task.GetBase().Notifications.Equal(manifest.Notifications{}) {
				if topLevelNotifications.NotificationsDefined() {
					task = task.SetNotifications(topLevelNotifications)
				}
				if task.GetBase().NotifyOnSuccess {
					taskNotifications := task.GetBase().Notifications
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
	man.Tasks = n.updateTasks(man.Tasks, man.Notifications)
	man.TeamsWebhook = ""
	return man, nil
}

func NewNotificationsMapper() Mapper {
	return notificationsMapper{}
}
