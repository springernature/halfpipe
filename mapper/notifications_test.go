package mapper

import (
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDoesNothingWhenSlackChannelIsNotDefined(t *testing.T) {
	updated, _ := NewNotificationsMapper().Apply(manifest.Manifest{})
	assert.Equal(t, manifest.Manifest{}, updated)
}

func TestTopLevelNotification(t *testing.T) {

	t.Run("slack_channel", func(t *testing.T) {
		channel := "#myCoolChannel"

		t.Run("set and notification not set", func(t *testing.T) {
			updated, _ := NewNotificationsMapper().Apply(manifest.Manifest{SlackChannel: channel})
			assert.Equal(t, manifest.Manifest{Notifications: manifest.Notifications{
				Slack: manifest.Channels{OnFailure: []string{channel}},
				Failure: manifest.NotificationChannels{
					{"slack": channel},
				},
			}}, updated)
		})

		t.Run("set and notification set, should not override", func(t *testing.T) {
			not := manifest.Notifications{
				Slack: manifest.Channels{OnFailure: []string{"#Howdie!"}},
				Failure: manifest.NotificationChannels{
					{"slack": "#Howdie!"},
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
				Teams: manifest.Channels{OnFailure: []string{webhook}},
				Failure: manifest.NotificationChannels{
					{"teams": webhook},
				},
			}}, updated)
		})

		t.Run("set and notification set, should not override", func(t *testing.T) {
			not := manifest.Notifications{Teams: manifest.Channels{OnFailure: []string{"kjlsfdajklfdsklfds"}}}

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
					Slack: manifest.Channels{OnFailure: []string{channel}},
					Teams: manifest.Channels{OnFailure: []string{webhook}},
					Failure: manifest.NotificationChannels{
						{"slack": channel},
						{"teams": webhook},
					},
				}}, updated)
			})

			t.Run("set and notification set, should not override", func(t *testing.T) {
				not := manifest.Notifications{Teams: manifest.Channels{OnSuccessMessage: "Howdie"}}

				updated, _ := NewNotificationsMapper().Apply(manifest.Manifest{SlackChannel: channel, TeamsWebhook: webhook, Notifications: not})
				assert.Equal(t, manifest.Manifest{
					Notifications: not,
				}, updated)
			})
		})
	})

	t.Run("slack_failure_message", func(t *testing.T) {
		failureMessage := "Oh noes"

		t.Run("set and notification not set", func(t *testing.T) {
			updated, _ := NewNotificationsMapper().Apply(manifest.Manifest{SlackFailureMessage: failureMessage})
			assert.Equal(t, manifest.Manifest{
				Notifications: manifest.Notifications{Slack: manifest.Channels{OnFailureMessage: failureMessage}}}, updated)
		})

		t.Run("set and notification set, should not override", func(t *testing.T) {
			not := manifest.Notifications{Slack: manifest.Channels{OnFailureMessage: "Wryyyyy"}}

			updated, _ := NewNotificationsMapper().Apply(manifest.Manifest{SlackFailureMessage: failureMessage, Notifications: not})
			assert.Equal(t, manifest.Manifest{
				Notifications: not,
			}, updated)

		})
	})

	t.Run("slack_success_message", func(t *testing.T) {
		successMessage := "Yay"
		t.Run("set and notification not set", func(t *testing.T) {
			updated, _ := NewNotificationsMapper().Apply(manifest.Manifest{SlackSuccessMessage: successMessage})
			assert.Equal(t, manifest.Manifest{
				Notifications: manifest.Notifications{Slack: manifest.Channels{OnSuccessMessage: successMessage}}}, updated)
		})

		t.Run("set and notification set, should not override", func(t *testing.T) {
			not := manifest.Notifications{Slack: manifest.Channels{OnSuccessMessage: "Wryyyyy"}}

			updated, _ := NewNotificationsMapper().Apply(manifest.Manifest{SlackSuccessMessage: successMessage, Notifications: not})
			assert.Equal(t, manifest.Manifest{
				Notifications: not,
			}, updated)

		})
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
		Slack: manifest.Channels{
			OnFailure:        []string{"1", "2"},
			OnFailureMessage: "Failure",
			OnSuccess:        []string{"a", "b"},
			OnSuccessMessage: "Success",
		},
		Failure: manifest.NotificationChannels{
			{"slack": "1", "message": "Failure"},
			{"slack": "2", "message": "Failure"},
		},
		Success: manifest.NotificationChannels{
			{"slack": "a", "message": "Success"},
			{"slack": "b", "message": "Success"},
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

func TestUpdatesNotificationsWhenSlackChannelIsDefined(t *testing.T) {
	t.Run("Only failure", func(t *testing.T) {
		t.Run("slack_channel", func(t *testing.T) {
			input := manifest.Manifest{
				SlackChannel: "#test",
				Tasks: manifest.TaskList{
					manifest.Run{},
					manifest.DockerPush{},
					manifest.Parallel{
						Tasks: manifest.TaskList{
							manifest.DeployMLZip{},
							manifest.Sequence{
								Tasks: manifest.TaskList{
									manifest.DeployCF{},
									manifest.ConsumerIntegrationTest{},
								},
							},
						},
					},
				},
			}

			notifications := manifest.Notifications{
				Slack: manifest.Channels{OnFailure: []string{input.SlackChannel}},
				Failure: manifest.NotificationChannels{
					{"slack": input.SlackChannel},
				},
			}

			expected := manifest.Manifest{
				Notifications: notifications,
				Tasks: manifest.TaskList{
					manifest.Run{Notifications: notifications},
					manifest.DockerPush{Notifications: notifications},
					manifest.Parallel{
						Tasks: manifest.TaskList{
							manifest.DeployMLZip{Notifications: notifications},
							manifest.Sequence{
								Tasks: manifest.TaskList{
									manifest.DeployCF{Notifications: notifications},
									manifest.ConsumerIntegrationTest{Notifications: notifications},
								},
							},
						},
					},
				},
			}

			updated, _ := NewNotificationsMapper().Apply(input)
			assert.Equal(t, expected, updated)

			// Make sure we don't update the old manifest in place, cus that leads to horrible bugs.
			assert.NotEqual(t, updated, input)
		})
		t.Run("top level notifications", func(t *testing.T) {
			notifications := manifest.Notifications{
				Slack: manifest.Channels{
					OnFailure: []string{"#oh-noes"},
				},
			}

			input := manifest.Manifest{
				SlackChannel:  "#test",
				Notifications: notifications,
				Tasks: manifest.TaskList{
					manifest.Run{},
					manifest.DockerPush{},
					manifest.Parallel{
						Tasks: manifest.TaskList{
							manifest.DeployMLZip{},
							manifest.Sequence{
								Tasks: manifest.TaskList{
									manifest.DeployCF{},
									manifest.ConsumerIntegrationTest{},
								},
							},
						},
					},
				},
			}

			expected := manifest.Manifest{
				Notifications: notifications,
				Tasks: manifest.TaskList{
					manifest.Run{Notifications: notifications},
					manifest.DockerPush{Notifications: notifications},
					manifest.Parallel{
						Tasks: manifest.TaskList{
							manifest.DeployMLZip{Notifications: notifications},
							manifest.Sequence{
								Tasks: manifest.TaskList{
									manifest.DeployCF{Notifications: notifications},
									manifest.ConsumerIntegrationTest{Notifications: notifications},
								},
							},
						},
					},
				},
			}

			updated, _ := NewNotificationsMapper().Apply(input)
			assert.Equal(t, expected, updated)

			// Make sure we don't update the old manifest in place, cus that leads to horrible bugs.
			assert.NotEqual(t, updated, input)
		})
	})

	t.Run("Doesn't update the cf push pre-promotes", func(t *testing.T) {
		t.Run("slack_channel", func(t *testing.T) {
			input := manifest.Manifest{
				SlackChannel: "#test",
				Tasks: manifest.TaskList{
					manifest.DeployCF{},
					manifest.DeployCF{
						PrePromote: manifest.TaskList{
							manifest.Run{},
							manifest.Run{},
						},
					},
					manifest.DeployCF{},
				},
			}

			notifications := manifest.Notifications{
				Slack: manifest.Channels{OnFailure: []string{input.SlackChannel}},
				Failure: manifest.NotificationChannels{
					{"slack": input.SlackChannel},
				},
			}

			expected := manifest.Manifest{
				Notifications: notifications,
				Tasks: manifest.TaskList{
					manifest.DeployCF{Notifications: notifications},
					manifest.DeployCF{
						Notifications: notifications,
						PrePromote: manifest.TaskList{
							manifest.Run{},
							manifest.Run{},
						},
					},
					manifest.DeployCF{
						Notifications: notifications,
					},
				},
			}

			updated, _ := NewNotificationsMapper().Apply(input)
			assert.Equal(t, expected, updated)
		})

		t.Run("top level notifications", func(t *testing.T) {
			notifications := manifest.Notifications{
				Slack: manifest.Channels{
					OnFailure: []string{"#oh-noes"},
				},
			}

			input := manifest.Manifest{
				Notifications: notifications,
				Tasks: manifest.TaskList{
					manifest.DeployCF{},
					manifest.DeployCF{
						PrePromote: manifest.TaskList{
							manifest.Run{},
							manifest.Run{},
						},
					},
					manifest.DeployCF{},
				},
			}

			expected := manifest.Manifest{
				Notifications: notifications,
				Tasks: manifest.TaskList{
					manifest.DeployCF{Notifications: notifications},
					manifest.DeployCF{
						Notifications: notifications,
						PrePromote: manifest.TaskList{
							manifest.Run{},
							manifest.Run{},
						},
					},
					manifest.DeployCF{
						Notifications: notifications,
					},
				},
			}

			updated, _ := NewNotificationsMapper().Apply(input)
			assert.Equal(t, expected, updated)
		})
	})

	t.Run("NotifyOnSuccess", func(t *testing.T) {
		t.Run("slack_channel", func(t *testing.T) {
			input := manifest.Manifest{
				SlackChannel: "#test",
				Tasks: manifest.TaskList{
					manifest.Run{},
					manifest.DockerPush{NotifyOnSuccess: true},
					manifest.Parallel{
						Tasks: manifest.TaskList{
							manifest.DeployMLZip{NotifyOnSuccess: true},
							manifest.Sequence{
								Tasks: manifest.TaskList{
									manifest.DeployCF{NotifyOnSuccess: true},
									manifest.ConsumerIntegrationTest{},
								},
							},
						},
					},
				},
			}

			notifications := manifest.Notifications{
				Slack: manifest.Channels{OnFailure: []string{input.SlackChannel}},
				Failure: manifest.NotificationChannels{
					{"slack": input.SlackChannel},
				},
			}

			notificationsWithSuccess := manifest.Notifications{
				Slack: manifest.Channels{OnFailure: []string{input.SlackChannel}, OnSuccess: []string{input.SlackChannel}},
				Failure: manifest.NotificationChannels{
					{"slack": input.SlackChannel},
				},
				Success: manifest.NotificationChannels{
					{"slack": input.SlackChannel},
				},
			}

			expected := manifest.Manifest{
				Notifications: notifications,
				Tasks: manifest.TaskList{
					manifest.Run{Notifications: notifications},
					manifest.DockerPush{Notifications: notificationsWithSuccess},
					manifest.Parallel{
						Tasks: manifest.TaskList{
							manifest.DeployMLZip{Notifications: notificationsWithSuccess},
							manifest.Sequence{
								Tasks: manifest.TaskList{
									manifest.DeployCF{Notifications: notificationsWithSuccess},
									manifest.ConsumerIntegrationTest{Notifications: notifications},
								},
							},
						},
					},
				},
			}

			updated, _ := NewNotificationsMapper().Apply(input)
			assert.Equal(t, expected.Tasks, updated.Tasks)
		})

		t.Run("top level notifications", func(t *testing.T) {
			notifications := manifest.Notifications{Slack: manifest.Channels{OnFailure: []string{"#OhNoes", "#AnotherOne"}}}
			input := manifest.Manifest{
				Notifications: notifications,
				Tasks: manifest.TaskList{
					manifest.Run{},
					manifest.DockerPush{NotifyOnSuccess: true},
					manifest.Parallel{
						Tasks: manifest.TaskList{
							manifest.DeployMLZip{NotifyOnSuccess: true},
							manifest.Sequence{
								Tasks: manifest.TaskList{
									manifest.DeployCF{NotifyOnSuccess: true},
									manifest.ConsumerIntegrationTest{},
								},
							},
						},
					},
				},
			}

			notificationsWithSuccess := manifest.Notifications{Slack: manifest.Channels{OnFailure: notifications.Slack.OnFailure, OnSuccess: []string{notifications.Slack.OnFailure[0]}}}
			expected := manifest.Manifest{
				Notifications: notifications,
				Tasks: manifest.TaskList{
					manifest.Run{Notifications: notifications},
					manifest.DockerPush{Notifications: notificationsWithSuccess},
					manifest.Parallel{
						Tasks: manifest.TaskList{
							manifest.DeployMLZip{Notifications: notificationsWithSuccess},
							manifest.Sequence{
								Tasks: manifest.TaskList{
									manifest.DeployCF{Notifications: notificationsWithSuccess},
									manifest.ConsumerIntegrationTest{Notifications: notifications},
								},
							},
						},
					},
				},
			}

			updated, _ := NewNotificationsMapper().Apply(input)
			assert.Equal(t, expected, updated)
		})
	})

	t.Run("Doesnt map if Notifications is already defined", func(t *testing.T) {
		t.Run("Old format", func(t *testing.T) {
			notifications := manifest.Notifications{
				Slack: manifest.Channels{OnFailure: []string{"#test"}},
				Failure: manifest.NotificationChannels{
					manifest.NotificationChannel{"slack": "#test"},
				},
			}

			input := manifest.Manifest{
				SlackChannel: "#test",
				Tasks: manifest.TaskList{
					manifest.Run{Notifications: manifest.Notifications{
						OnSuccess: []string{"1"},
						OnFailure: []string{"2"},
					}},
					manifest.DockerPush{},
					manifest.Parallel{
						Tasks: manifest.TaskList{
							manifest.DeployMLZip{NotifyOnSuccess: true},
							manifest.Sequence{
								Tasks: manifest.TaskList{
									manifest.DeployCF{NotifyOnSuccess: true},
									manifest.ConsumerIntegrationTest{Notifications: manifest.Notifications{
										OnSuccess: []string{"a", "b"},
										OnFailure: []string{"x", "y", "z"},
									}},
								},
							},
						},
					},
				},
			}

			notificationsWithSuccess := manifest.Notifications{
				Slack:   manifest.Channels{OnFailure: notifications.Slack.OnFailure, OnSuccess: notifications.Slack.OnFailure},
				Failure: manifest.NotificationChannels{manifest.NotificationChannel{"slack": input.SlackChannel}},
				Success: manifest.NotificationChannels{manifest.NotificationChannel{"slack": input.SlackChannel}},
			}

			expected := manifest.Manifest{
				Notifications: notifications,
				Tasks: manifest.TaskList{
					manifest.Run{
						Notifications: manifest.Notifications{
							Slack: manifest.Channels{
								OnSuccess: []string{"1"},
								OnFailure: []string{"2"},
							},
							Failure: manifest.NotificationChannels{{"slack": "2"}},
							Success: manifest.NotificationChannels{{"slack": "1"}},
						},
					},
					manifest.DockerPush{Notifications: notifications},
					manifest.Parallel{
						Tasks: manifest.TaskList{
							manifest.DeployMLZip{Notifications: notificationsWithSuccess},
							manifest.Sequence{
								Tasks: manifest.TaskList{
									manifest.DeployCF{Notifications: notificationsWithSuccess},
									manifest.ConsumerIntegrationTest{
										Notifications: manifest.Notifications{
											Slack: manifest.Channels{
												OnSuccess: []string{"a", "b"},
												OnFailure: []string{"x", "y", "z"},
											},
											Failure: manifest.NotificationChannels{{"slack": "x"}, {"slack": "y"}, {"slack": "z"}},
											Success: manifest.NotificationChannels{{"slack": "a"}, {"slack": "b"}},
										},
									},
								},
							},
						},
					},
				},
			}

			updated, _ := NewNotificationsMapper().Apply(input)
			assert.Equal(t, expected, updated)
		})
	})
}

func TestUpdatesNotificationsWhenTeamsWebhookIsDefined(t *testing.T) {
	t.Run("Only failure", func(t *testing.T) {
		t.Run("teams_webhook", func(t *testing.T) {
			input := manifest.Manifest{
				TeamsWebhook: "https://",
				Tasks: manifest.TaskList{
					manifest.Run{},
					manifest.DockerPush{},
					manifest.Parallel{
						Tasks: manifest.TaskList{
							manifest.DeployMLZip{},
							manifest.Sequence{
								Tasks: manifest.TaskList{
									manifest.DeployCF{},
									manifest.ConsumerIntegrationTest{},
								},
							},
						},
					},
				},
			}

			notifications := manifest.Notifications{
				Teams: manifest.Channels{OnFailure: []string{input.TeamsWebhook}},
				Failure: manifest.NotificationChannels{
					manifest.NotificationChannel{"teams": input.TeamsWebhook},
				},
			}
			expected := manifest.Manifest{
				Notifications: notifications,
				Tasks: manifest.TaskList{
					manifest.Run{Notifications: notifications},
					manifest.DockerPush{Notifications: notifications},
					manifest.Parallel{
						Tasks: manifest.TaskList{
							manifest.DeployMLZip{Notifications: notifications},
							manifest.Sequence{
								Tasks: manifest.TaskList{
									manifest.DeployCF{Notifications: notifications},
									manifest.ConsumerIntegrationTest{Notifications: notifications},
								},
							},
						},
					},
				},
			}

			updated, _ := NewNotificationsMapper().Apply(input)
			assert.Equal(t, expected, updated)

			// Make sure we don't update the old manifest in place, cus that leads to horrible bugs.
			assert.NotEqual(t, updated, input)
		})
	})

	t.Run("Doesn't update the cf push pre-promotes", func(t *testing.T) {
		t.Run("teams_webhook", func(t *testing.T) {
			input := manifest.Manifest{
				TeamsWebhook: "https://",
				Tasks: manifest.TaskList{
					manifest.DeployCF{},
					manifest.DeployCF{
						PrePromote: manifest.TaskList{
							manifest.Run{},
							manifest.Run{},
						},
					},
					manifest.DeployCF{},
				},
			}

			notifications := manifest.Notifications{
				Teams: manifest.Channels{OnFailure: []string{input.TeamsWebhook}},
				Failure: manifest.NotificationChannels{
					{"teams": input.TeamsWebhook},
				},
			}

			expected := manifest.Manifest{
				Notifications: notifications,
				Tasks: manifest.TaskList{
					manifest.DeployCF{Notifications: notifications},
					manifest.DeployCF{
						Notifications: notifications,
						PrePromote: manifest.TaskList{
							manifest.Run{},
							manifest.Run{},
						},
					},
					manifest.DeployCF{
						Notifications: notifications,
					},
				},
			}

			updated, _ := NewNotificationsMapper().Apply(input)
			assert.Equal(t, expected, updated)
		})

		t.Run("top level notifications", func(t *testing.T) {
			notifications := manifest.Notifications{
				Teams: manifest.Channels{
					OnFailure: []string{"https://sdfa"},
				},
			}

			input := manifest.Manifest{
				Notifications: notifications,
				Tasks: manifest.TaskList{
					manifest.DeployCF{},
					manifest.DeployCF{
						PrePromote: manifest.TaskList{
							manifest.Run{},
							manifest.Run{},
						},
					},
					manifest.DeployCF{},
				},
			}

			expected := manifest.Manifest{
				Notifications: notifications,
				Tasks: manifest.TaskList{
					manifest.DeployCF{Notifications: notifications},
					manifest.DeployCF{
						Notifications: notifications,
						PrePromote: manifest.TaskList{
							manifest.Run{},
							manifest.Run{},
						},
					},
					manifest.DeployCF{
						Notifications: notifications,
					},
				},
			}

			updated, _ := NewNotificationsMapper().Apply(input)
			assert.Equal(t, expected, updated)
		})
	})

	t.Run("NotifyOnSuccess", func(t *testing.T) {
		t.Run("teams_webhook", func(t *testing.T) {
			input := manifest.Manifest{
				TeamsWebhook: "https://",
				Tasks: manifest.TaskList{
					manifest.Run{},
					manifest.DockerPush{NotifyOnSuccess: true},
					manifest.Parallel{
						Tasks: manifest.TaskList{
							manifest.DeployMLZip{NotifyOnSuccess: true},
							manifest.Sequence{
								Tasks: manifest.TaskList{
									manifest.DeployCF{NotifyOnSuccess: true},
									manifest.ConsumerIntegrationTest{},
								},
							},
						},
					},
				},
			}

			notifications := manifest.Notifications{
				Teams: manifest.Channels{OnFailure: []string{input.TeamsWebhook}},
				Failure: manifest.NotificationChannels{
					{"teams": input.TeamsWebhook},
				},
			}
			notificationsWithSuccess := manifest.Notifications{
				Teams:   manifest.Channels{OnFailure: []string{input.TeamsWebhook}, OnSuccess: []string{input.TeamsWebhook}},
				Failure: manifest.NotificationChannels{{"teams": input.TeamsWebhook}},
				Success: manifest.NotificationChannels{{"teams": input.TeamsWebhook}},
			}
			expected := manifest.Manifest{
				Notifications: notifications,
				Tasks: manifest.TaskList{
					manifest.Run{Notifications: notifications},
					manifest.DockerPush{Notifications: notificationsWithSuccess},
					manifest.Parallel{
						Tasks: manifest.TaskList{
							manifest.DeployMLZip{Notifications: notificationsWithSuccess},
							manifest.Sequence{
								Tasks: manifest.TaskList{
									manifest.DeployCF{Notifications: notificationsWithSuccess},
									manifest.ConsumerIntegrationTest{Notifications: notifications},
								},
							},
						},
					},
				},
			}

			updated, _ := NewNotificationsMapper().Apply(input)
			assert.Equal(t, expected, updated)
		})
		t.Run("top level notifications", func(t *testing.T) {
			notifications := manifest.Notifications{Teams: manifest.Channels{OnFailure: []string{"#OhNoes", "#AnotherOne"}}}
			input := manifest.Manifest{
				Notifications: notifications,
				Tasks: manifest.TaskList{
					manifest.Run{},
					manifest.DockerPush{NotifyOnSuccess: true},
					manifest.Parallel{
						Tasks: manifest.TaskList{
							manifest.DeployMLZip{NotifyOnSuccess: true},
							manifest.Sequence{
								Tasks: manifest.TaskList{
									manifest.DeployCF{NotifyOnSuccess: true},
									manifest.ConsumerIntegrationTest{},
								},
							},
						},
					},
				},
			}

			notificationsWithSuccess := manifest.Notifications{Teams: manifest.Channels{OnFailure: notifications.Teams.OnFailure, OnSuccess: []string{notifications.Teams.OnFailure[0]}}}
			expected := manifest.Manifest{
				Notifications: notifications,
				Tasks: manifest.TaskList{
					manifest.Run{Notifications: notifications},
					manifest.DockerPush{Notifications: notificationsWithSuccess},
					manifest.Parallel{
						Tasks: manifest.TaskList{
							manifest.DeployMLZip{Notifications: notificationsWithSuccess},
							manifest.Sequence{
								Tasks: manifest.TaskList{
									manifest.DeployCF{Notifications: notificationsWithSuccess},
									manifest.ConsumerIntegrationTest{Notifications: notifications},
								},
							},
						},
					},
				},
			}

			updated, _ := NewNotificationsMapper().Apply(input)
			assert.Equal(t, expected, updated)
		})
	})
}

func TestUpdatesNotificationsWhenSlackChannelAndTeamsWebhookIsDefined(t *testing.T) {
	t.Run("Only failure", func(t *testing.T) {
		t.Run("both", func(t *testing.T) {
			input := manifest.Manifest{
				SlackChannel: "#blah",
				TeamsWebhook: "https://",
				Tasks: manifest.TaskList{
					manifest.Run{},
					manifest.DockerPush{},
					manifest.Parallel{
						Tasks: manifest.TaskList{
							manifest.DeployMLZip{},
							manifest.Sequence{
								Tasks: manifest.TaskList{
									manifest.DeployCF{},
									manifest.ConsumerIntegrationTest{},
								},
							},
						},
					},
				},
			}

			notifications := manifest.Notifications{
				Slack:   manifest.Channels{OnFailure: []string{input.SlackChannel}},
				Teams:   manifest.Channels{OnFailure: []string{input.TeamsWebhook}},
				Failure: manifest.NotificationChannels{{"slack": input.SlackChannel}, {"teams": input.TeamsWebhook}},
			}
			expected := manifest.Manifest{
				Notifications: notifications,
				Tasks: manifest.TaskList{
					manifest.Run{Notifications: notifications},
					manifest.DockerPush{Notifications: notifications},
					manifest.Parallel{
						Tasks: manifest.TaskList{
							manifest.DeployMLZip{Notifications: notifications},
							manifest.Sequence{
								Tasks: manifest.TaskList{
									manifest.DeployCF{Notifications: notifications},
									manifest.ConsumerIntegrationTest{Notifications: notifications},
								},
							},
						},
					},
				},
			}

			updated, _ := NewNotificationsMapper().Apply(input)
			assert.Equal(t, expected, updated)

			// Make sure we don't update the old manifest in place, cus that leads to horrible bugs.
			assert.NotEqual(t, updated, input)
		})
	})

	t.Run("Doesn't update the cf push pre-promotes", func(t *testing.T) {
		t.Run("both", func(t *testing.T) {
			input := manifest.Manifest{
				SlackChannel: "#blah",
				TeamsWebhook: "https://",
				Tasks: manifest.TaskList{
					manifest.DeployCF{},
					manifest.DeployCF{
						PrePromote: manifest.TaskList{
							manifest.Run{},
							manifest.Run{},
						},
					},
					manifest.DeployCF{},
				},
			}

			notifications := manifest.Notifications{
				Slack:   manifest.Channels{OnFailure: []string{input.SlackChannel}},
				Teams:   manifest.Channels{OnFailure: []string{input.TeamsWebhook}},
				Failure: manifest.NotificationChannels{{"slack": input.SlackChannel}, {"teams": input.TeamsWebhook}},
			}
			expected := manifest.Manifest{
				Notifications: notifications,
				Tasks: manifest.TaskList{
					manifest.DeployCF{Notifications: notifications},
					manifest.DeployCF{
						Notifications: notifications,
						PrePromote: manifest.TaskList{
							manifest.Run{},
							manifest.Run{},
						},
					},
					manifest.DeployCF{
						Notifications: notifications,
					},
				},
			}

			updated, _ := NewNotificationsMapper().Apply(input)
			assert.Equal(t, expected, updated)
		})

		t.Run("top level notifications", func(t *testing.T) {
			notifications := manifest.Notifications{
				Slack: manifest.Channels{
					OnFailure: []string{"#blah"},
				},
				Teams: manifest.Channels{
					OnFailure: []string{"https://sdfa"},
				},
			}

			input := manifest.Manifest{
				Notifications: notifications,
				Tasks: manifest.TaskList{
					manifest.DeployCF{},
					manifest.DeployCF{
						PrePromote: manifest.TaskList{
							manifest.Run{},
							manifest.Run{},
						},
					},
					manifest.DeployCF{},
				},
			}

			expected := manifest.Manifest{
				Notifications: notifications,
				Tasks: manifest.TaskList{
					manifest.DeployCF{Notifications: notifications},
					manifest.DeployCF{
						Notifications: notifications,
						PrePromote: manifest.TaskList{
							manifest.Run{},
							manifest.Run{},
						},
					},
					manifest.DeployCF{
						Notifications: notifications,
					},
				},
			}

			updated, _ := NewNotificationsMapper().Apply(input)
			assert.Equal(t, expected, updated)
		})
	})

	t.Run("NotifyOnSuccess", func(t *testing.T) {
		t.Run("both", func(t *testing.T) {
			input := manifest.Manifest{
				SlackChannel: "#blah",
				TeamsWebhook: "https://",
				Tasks: manifest.TaskList{
					manifest.Run{},
					manifest.DockerPush{NotifyOnSuccess: true},
					manifest.Parallel{
						Tasks: manifest.TaskList{
							manifest.DeployMLZip{NotifyOnSuccess: true},
							manifest.Sequence{
								Tasks: manifest.TaskList{
									manifest.DeployCF{NotifyOnSuccess: true},
									manifest.ConsumerIntegrationTest{},
								},
							},
						},
					},
				},
			}

			notifications := manifest.Notifications{
				Slack:   manifest.Channels{OnFailure: []string{input.SlackChannel}},
				Teams:   manifest.Channels{OnFailure: []string{input.TeamsWebhook}},
				Failure: manifest.NotificationChannels{{"slack": input.SlackChannel}, {"teams": input.TeamsWebhook}},
			}
			notificationsWithSuccess := manifest.Notifications{
				Slack:   manifest.Channels{OnFailure: []string{input.SlackChannel}, OnSuccess: []string{input.SlackChannel}},
				Teams:   manifest.Channels{OnFailure: []string{input.TeamsWebhook}, OnSuccess: []string{input.TeamsWebhook}},
				Failure: manifest.NotificationChannels{{"slack": input.SlackChannel}, {"teams": input.TeamsWebhook}},
				Success: manifest.NotificationChannels{{"slack": input.SlackChannel}, {"teams": input.TeamsWebhook}},
			}

			expected := manifest.Manifest{
				Notifications: notifications,
				Tasks: manifest.TaskList{
					manifest.Run{Notifications: notifications},
					manifest.DockerPush{Notifications: notificationsWithSuccess},
					manifest.Parallel{
						Tasks: manifest.TaskList{
							manifest.DeployMLZip{Notifications: notificationsWithSuccess},
							manifest.Sequence{
								Tasks: manifest.TaskList{
									manifest.DeployCF{Notifications: notificationsWithSuccess},
									manifest.ConsumerIntegrationTest{Notifications: notifications},
								},
							},
						},
					},
				},
			}

			updated, _ := NewNotificationsMapper().Apply(input)
			assert.Equal(t, expected, updated)
		})

		t.Run("top level notifications", func(t *testing.T) {
			notifications := manifest.Notifications{Slack: manifest.Channels{OnFailure: []string{"a", "b"}}, Teams: manifest.Channels{OnFailure: []string{"1", "2"}}}
			input := manifest.Manifest{
				Notifications: notifications,
				Tasks: manifest.TaskList{
					manifest.Run{},
					manifest.DockerPush{NotifyOnSuccess: true},
					manifest.Parallel{
						Tasks: manifest.TaskList{
							manifest.DeployMLZip{NotifyOnSuccess: true},
							manifest.Sequence{
								Tasks: manifest.TaskList{
									manifest.DeployCF{NotifyOnSuccess: true},
									manifest.ConsumerIntegrationTest{},
								},
							},
						},
					},
				},
			}

			notificationsWithSuccess := manifest.Notifications{
				Slack: manifest.Channels{OnFailure: notifications.Slack.OnFailure, OnSuccess: []string{notifications.Slack.OnFailure[0]}},
				Teams: manifest.Channels{OnFailure: notifications.Teams.OnFailure, OnSuccess: []string{notifications.Teams.OnFailure[0]}},
			}
			expected := manifest.Manifest{
				Notifications: notifications,
				Tasks: manifest.TaskList{
					manifest.Run{Notifications: notifications},
					manifest.DockerPush{Notifications: notificationsWithSuccess},
					manifest.Parallel{
						Tasks: manifest.TaskList{
							manifest.DeployMLZip{Notifications: notificationsWithSuccess},
							manifest.Sequence{
								Tasks: manifest.TaskList{
									manifest.DeployCF{Notifications: notificationsWithSuccess},
									manifest.ConsumerIntegrationTest{Notifications: notifications},
								},
							},
						},
					},
				},
			}

			updated, _ := NewNotificationsMapper().Apply(input)
			assert.Equal(t, expected, updated)
		})
	})
}

func TestDefaultNotificationMessages(t *testing.T) {
	defaultFailureMessage := "failure msg"
	defaultSuccessMessage := "success msg"

	input := manifest.Manifest{
		SlackChannel:        "#test",
		SlackFailureMessage: defaultFailureMessage,
		SlackSuccessMessage: defaultSuccessMessage,
		Tasks: manifest.TaskList{
			manifest.Run{},
			manifest.DockerPush{},
			manifest.DeployCF{
				Notifications: manifest.Notifications{
					OnSuccess:        []string{"#foo"},
					OnSuccessMessage: "custom",
					OnFailureMessage: "custom",
				},
			},
		},
	}

	updated, _ := NewNotificationsMapper().Apply(input)
	assert.Equal(t, defaultFailureMessage, updated.Tasks[0].GetNotifications().Slack.OnFailureMessage)
	assert.Equal(t, defaultFailureMessage, updated.Tasks[1].GetNotifications().Slack.OnFailureMessage)
	assert.Equal(t, "custom", updated.Tasks[2].GetNotifications().Slack.OnFailureMessage)

	assert.Equal(t, defaultSuccessMessage, updated.Tasks[0].GetNotifications().Slack.OnSuccessMessage)
	assert.Equal(t, defaultSuccessMessage, updated.Tasks[1].GetNotifications().Slack.OnSuccessMessage)
	assert.Equal(t, "custom", updated.Tasks[2].GetNotifications().Slack.OnFailureMessage)
}
