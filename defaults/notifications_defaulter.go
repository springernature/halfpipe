package defaults

import (
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
	"slices"
)

type notificationsDefaulter struct {
}

func (n notificationsDefaulter) Apply(original manifest.Manifest) (updated manifest.Manifest) {
	updated = original
	if !slices.ContainsFunc(updated.Notifications.Failure, func(channel manifest.NotificationChannel) bool {
		return channel.Teams != ""
	}) && updated.FeatureToggles.AutomaticNotifications() {
		channel := original.TeamsWebhook
		if channel == "" {
			channel = config.PlatformAPIMessageURL + "?team=" + updated.Team
		}
		updated.Notifications.Failure = manifest.NotificationChannels{
			{Teams: channel},
		}
	}
	return updated
}

func NewNotificationsDefaulter() NotificationsDefaulter {
	return notificationsDefaulter{}
}
