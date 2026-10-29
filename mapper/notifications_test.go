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

	t.Run("slack_channel", func(t *testing.T) {
		channel := "#myCoolChannel"

		t.Run("set and notification not set", func(t *testing.T) {
			updated, _ := NewNotificationsMapper().Apply(manifest.Manifest{SlackChannel: channel})
			assert.Equal(t, manifest.Manifest{Notifications: manifest.Notifications{
				Failure: manifest.NotificationChannels{
					{Slack: channel},
				},
			}}, updated)
		})

		t.Run("set and notification set, should not override", func(t *testing.T) {
			not := manifest.Notifications{
				Failure: manifest.NotificationChannels{
					{Slack: "#Howdie!"},
				},
			}

			updated, _ := NewNotificationsMapper().Apply(manifest.Manifest{SlackChannel: channel, Notifications: not})
			assert.Equal(t, manifest.Manifest{
				Notifications: not,
			}, updated)

		})
	})

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

		t.Run("slack_channel and teams_channel", func(t *testing.T) {
			channel := "#blah"
			webhook := "https://blabla"

			t.Run("set and notification not set", func(t *testing.T) {
				updated, _ := NewNotificationsMapper().Apply(manifest.Manifest{SlackChannel: channel, TeamsWebhook: webhook})
				assert.Equal(t, manifest.Manifest{Notifications: manifest.Notifications{
					Failure: manifest.NotificationChannels{
						{Slack: channel},
						{Teams: webhook},
					},
				}}, updated)
			})

			t.Run("set and notification set, should not override", func(t *testing.T) {
				not := manifest.Notifications{Success: manifest.NotificationChannels{{Teams: "yo"}}}

				updated, _ := NewNotificationsMapper().Apply(manifest.Manifest{SlackChannel: channel, TeamsWebhook: webhook, Notifications: not})
				assert.Equal(t, manifest.Manifest{
					Notifications: not,
				}, updated)
			})
		})
	})

	t.Run("slack_failure_message", func(t *testing.T) {
		failureMessage := "Oh noes"

		t.Run("set and no slack channel set", func(t *testing.T) {
			updated, _ := NewNotificationsMapper().Apply(manifest.Manifest{SlackFailureMessage: failureMessage})
			assert.Equal(t, manifest.Manifest{}, updated)
		})

		t.Run("set and slack channel set", func(t *testing.T) {
			updated, _ := NewNotificationsMapper().Apply(manifest.Manifest{SlackFailureMessage: failureMessage, SlackChannel: "#yo"})
			assert.Equal(t, manifest.Manifest{
				Notifications: manifest.Notifications{
					Failure: manifest.NotificationChannels{
						{Slack: "#yo", Message: failureMessage},
					},
				},
			}, updated)
		})

		t.Run("set and notification set, should not override", func(t *testing.T) {
			not := manifest.Notifications{Failure: manifest.NotificationChannels{{Slack: "#Somethin"}}}

			updated, _ := NewNotificationsMapper().Apply(manifest.Manifest{SlackFailureMessage: failureMessage, SlackChannel: "#yo", Notifications: not})
			assert.Equal(t, manifest.Manifest{
				Notifications: not,
			}, updated)

		})
	})
}
