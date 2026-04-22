package mapper

import (
	"testing"

	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

var (
	team            = "my-team"
	slackWebhookURL = config.PlatformAPIMessageURL + "?team=" + team
)

func TestSlackToTeamsMapper_SetsTeamsFromSlack(t *testing.T) {
	input := manifest.Manifest{
		Team: team,
		Tasks: manifest.TaskList{
			manifest.Run{
				TaskBase: manifest.TaskBase{Notifications: manifest.Notifications{
					Failure: manifest.NotificationChannels{
						{Slack: "#alerts"},
					},
					Success: manifest.NotificationChannels{
						{Slack: "#deploys"},
					},
				}},
			},
		},
	}

	expected := manifest.Manifest{
		Team: team,
		Tasks: manifest.TaskList{
			manifest.Run{
				TaskBase: manifest.TaskBase{Notifications: manifest.Notifications{
					Failure: manifest.NotificationChannels{
						{Slack: "#alerts"},
						{Teams: slackWebhookURL},
					},
					Success: manifest.NotificationChannels{
						{Slack: "#deploys"},
						{Teams: slackWebhookURL},
					},
				}},
			},
		},
	}

	updated, err := NewSlackToTeamsMapper().Apply(input)
	assert.NoError(t, err)
	assert.Equal(t, expected, updated)
}

func TestSlackToTeamsMapper_RecursesIntoContainerTasks(t *testing.T) {
	input := manifest.Manifest{
		Team: team,
		Tasks: manifest.TaskList{
			manifest.Sequence{
				Tasks: manifest.TaskList{
					manifest.Parallel{
						Tasks: manifest.TaskList{
							manifest.Run{
								TaskBase: manifest.TaskBase{Notifications: manifest.Notifications{
									Failure: manifest.NotificationChannels{
										{Slack: "#deep"},
									},
								}},
							},
						},
					},
				},
			},
		},
	}

	updated, err := NewSlackToTeamsMapper().Apply(input)
	assert.NoError(t, err)
	runTask := updated.Tasks[0].(manifest.Sequence).Tasks[0].(manifest.Parallel).Tasks[0].(manifest.Run)
	assert.Equal(t, manifest.NotificationChannels{
		{Slack: "#deep"},
		{Teams: slackWebhookURL},
	}, runTask.Notifications.Failure)
}

func TestSlackToTeamsMapper_DoesNotOverwriteExistingTeamsURL(t *testing.T) {
	existingURL := "https://custom-webhook.example.com/my-hook"

	input := manifest.Manifest{
		Team: team,
		Tasks: manifest.TaskList{
			manifest.Run{
				TaskBase: manifest.TaskBase{Notifications: manifest.Notifications{
					Failure: manifest.NotificationChannels{
						{Slack: "#alerts"},
						{Teams: existingURL},
					},
				}},
			},
		},
	}

	updated, err := NewSlackToTeamsMapper().Apply(input)
	assert.NoError(t, err)
	channels := updated.Tasks[0].(manifest.Run).Notifications.Failure
	assert.Equal(t, existingURL, channels[1].Teams)
	assert.Len(t, channels, 3) // original slack, original teams, new generated teams
}

func TestSlackToTeamsMapper_NoDuplicateTeamsMessages(t *testing.T) {
	t.Run("multiple slack channels produce only one teams entry", func(t *testing.T) {
		input := manifest.Manifest{
			Team: team,
			Tasks: manifest.TaskList{
				manifest.Run{
					TaskBase: manifest.TaskBase{Notifications: manifest.Notifications{
						Failure: manifest.NotificationChannels{
							{Slack: "#alerts"},
							{Slack: "#errors"},
						},
					}},
				},
			},
		}

		updated, err := NewSlackToTeamsMapper().Apply(input)
		assert.NoError(t, err)
		channels := updated.Tasks[0].(manifest.Run).Notifications.Failure
		assert.Equal(t, manifest.NotificationChannels{
			{Slack: "#alerts"},
			{Slack: "#errors"},
			{Teams: slackWebhookURL},
		}, channels)
	})

	t.Run("existing teams url blocks adding a new one", func(t *testing.T) {
		input := manifest.Manifest{
			Team: team,
			Tasks: manifest.TaskList{
				manifest.Run{
					TaskBase: manifest.TaskBase{Notifications: manifest.Notifications{
						Failure: manifest.NotificationChannels{
							{Teams: slackWebhookURL},
							{Slack: "#alerts"},
						},
					}},
				},
			},
		}

		updated, err := NewSlackToTeamsMapper().Apply(input)
		assert.NoError(t, err)
		assert.Equal(t, input, updated)
	})

	t.Run("failure and success deduplicate independently", func(t *testing.T) {
		input := manifest.Manifest{
			Team: team,
			Tasks: manifest.TaskList{
				manifest.Run{
					TaskBase: manifest.TaskBase{Notifications: manifest.Notifications{
						Failure: manifest.NotificationChannels{
							{Slack: "#alerts"},
							{Slack: "#errors"},
						},
						Success: manifest.NotificationChannels{
							{Slack: "#deploys"},
							{Slack: "#releases"},
						},
					}},
				},
			},
		}

		updated, err := NewSlackToTeamsMapper().Apply(input)
		assert.NoError(t, err)
		n := updated.Tasks[0].(manifest.Run).Notifications
		assert.Equal(t, manifest.NotificationChannels{
			{Slack: "#alerts"},
			{Slack: "#errors"},
			{Teams: slackWebhookURL},
		}, n.Failure)
		assert.Equal(t, manifest.NotificationChannels{
			{Slack: "#deploys"},
			{Slack: "#releases"},
			{Teams: slackWebhookURL},
		}, n.Success)
	})
}
