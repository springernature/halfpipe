package mapper

import (
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDoesNothingWhenNoNotificationsIsDefined(t *testing.T) {
	updated, _ := NewNotificationsMapper().Apply(manifest.Manifest{})
	assert.Equal(t, manifest.Manifest{}, updated)
}

func TestTopLevelNotification(t *testing.T) {
	t.Run("teams_channel", func(t *testing.T) {
		webhook := "https://blabla"

		t.Run("set and notification not set", func(t *testing.T) {
			updated, _ := NewNotificationsMapper().Apply(manifest.Manifest{TeamsWebhook: webhook})
			assert.Equal(t, manifest.Manifest{Notifications: manifest.Notifications{
				Failure: manifest.NotificationChannels{
					{Teams: webhook},
				},
			}}, updated)
		})

		t.Run("set and notification set, should not override", func(t *testing.T) {
			not := manifest.Notifications{Success: manifest.NotificationChannels{{Teams: "some-random-webhook"}}}

			updated, _ := NewNotificationsMapper().Apply(manifest.Manifest{TeamsWebhook: webhook, Notifications: not})
			assert.Equal(t, manifest.Manifest{
				Notifications: not,
			}, updated)
		})
	})
}
