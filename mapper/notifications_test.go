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

	t.Run("slack_success_message", func(t *testing.T) {
		input := manifest.Manifest{
			SlackChannel:        "Blah",
			SlackSuccessMessage: "Yo",
			Tasks: manifest.TaskList{
				manifest.Run{},
				manifest.Run{NotifyOnSuccess: true},
				manifest.Run{Notifications: manifest.Notifications{
					Success: manifest.NotificationChannels{
						{Slack: "#yo", Message: "Hello"},
						{Slack: "#yo"},
					},
				}},
			},
		}

		expected := manifest.Manifest{
			Notifications: manifest.Notifications{
				Failure: manifest.NotificationChannels{{Slack: "Blah"}},
			},
			Tasks: manifest.TaskList{
				manifest.Run{
					Notifications: manifest.Notifications{
						Failure: manifest.NotificationChannels{{Slack: "Blah"}},
					},
				},
				manifest.Run{
					Notifications: manifest.Notifications{
						Failure: manifest.NotificationChannels{{Slack: "Blah"}},
						Success: manifest.NotificationChannels{{Slack: "Blah", Message: "Yo"}},
					},
				},
				manifest.Run{Notifications: manifest.Notifications{
					Success: manifest.NotificationChannels{
						{Slack: "#yo", Message: "Hello"},
						{Slack: "#yo"},
					},
				}},
			},
		}

		updated, _ := NewNotificationsMapper().Apply(input)
		assert.Equal(t, expected, updated)
	})
}

func TestMigrateTaskLevelNotifications(t *testing.T) {
	inputNotification := manifest.Notifications{
		OnFailure:        []string{"1", "2"},
		OnFailureMessage: "Failure",
		OnSuccess:        []string{"a", "b"},
		OnSuccessMessage: "Success",
	}

	expectedNotification := manifest.Notifications{
		Failure: manifest.NotificationChannels{
			{Slack: "1", Message: "Failure"},
			{Slack: "2", Message: "Failure"},
		},
		Success: manifest.NotificationChannels{
			{Slack: "a", Message: "Success"},
			{Slack: "b", Message: "Success"},
		},
	}

	input := manifest.Manifest{
		Tasks: manifest.TaskList{
			manifest.Run{Notifications: inputNotification},
			manifest.DockerPush{Notifications: inputNotification},
			manifest.Parallel{
				Tasks: manifest.TaskList{
					manifest.DeployMLZip{Notifications: inputNotification},
					manifest.Sequence{
						Tasks: manifest.TaskList{
							manifest.DeployCF{},
							manifest.ConsumerIntegrationTest{Notifications: inputNotification},
						},
					},
				},
			},
		},
	}

	expected := manifest.Manifest{
		Tasks: manifest.TaskList{
			manifest.Run{Notifications: expectedNotification},
			manifest.DockerPush{Notifications: expectedNotification},
			manifest.Parallel{
				Tasks: manifest.TaskList{
					manifest.DeployMLZip{Notifications: expectedNotification},
					manifest.Sequence{
						Tasks: manifest.TaskList{
							manifest.DeployCF{},
							manifest.ConsumerIntegrationTest{Notifications: expectedNotification},
						},
					},
				},
			},
		},
	}

	updated, _ := NewNotificationsMapper().Apply(input)
	assert.Equal(t, expected, updated)
}

func TestNotifyOnSuccess(t *testing.T) {
	slack := "yo"
	teams := "kehe"

	input := manifest.Manifest{
		SlackChannel: slack,
		TeamsWebhook: teams,
		Tasks: manifest.TaskList{
			manifest.Run{},
			manifest.Run{NotifyOnSuccess: true},
		},
	}

	expectedNotifications := manifest.Notifications{
		Failure: manifest.NotificationChannels{
			{Slack: slack},
			{Teams: teams},
		},
	}

	expected := manifest.Manifest{
		Notifications: expectedNotifications,
		Tasks: manifest.TaskList{
			manifest.Run{Notifications: expectedNotifications},
			manifest.Run{Notifications: manifest.Notifications{
				Failure: manifest.NotificationChannels{
					{Slack: slack},
					{Teams: teams},
				},
				Success: manifest.NotificationChannels{
					{Slack: slack},
					{Teams: teams},
				},
			}},
		},
	}
	updated, _ := NewNotificationsMapper().Apply(input)
	assert.Equal(t, expected, updated)
}
