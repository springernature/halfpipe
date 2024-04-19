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
			assert.Equal(t, manifest.Manifest{Notifications: manifest.Notifications{Slack: manifest.Slack{OnFailure: []string{channel}}}}, updated)
		})
		t.Run("set and notification set, should not override", func(t *testing.T) {
			not := manifest.Notifications{Slack: manifest.Slack{OnFailure: []string{"#Howdie!"}}}

			updated, _ := NewNotificationsMapper().Apply(manifest.Manifest{SlackChannel: channel, Notifications: not})
			assert.Equal(t, manifest.Manifest{
				Notifications: not,
			}, updated)

		})
	})

	t.Run("slack_failure_message", func(t *testing.T) {
		failureMessage := "Oh noes"

		t.Run("set and notification not set", func(t *testing.T) {
			updated, _ := NewNotificationsMapper().Apply(manifest.Manifest{SlackFailureMessage: failureMessage})
			assert.Equal(t, manifest.Manifest{
				Notifications: manifest.Notifications{Slack: manifest.Slack{OnFailureMessage: failureMessage}}}, updated)
		})

		t.Run("set and notification set, should not override", func(t *testing.T) {
			not := manifest.Notifications{Slack: manifest.Slack{OnFailureMessage: "Wryyyyy"}}

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
				Notifications: manifest.Notifications{Slack: manifest.Slack{OnSuccessMessage: successMessage}}}, updated)
		})

		t.Run("set and notification set, should not override", func(t *testing.T) {
			not := manifest.Notifications{Slack: manifest.Slack{OnSuccessMessage: "Wryyyyy"}}

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
		Slack: manifest.Slack{
			OnFailure:        []string{"1", "2"},
			OnFailureMessage: "Failure",
			OnSuccess:        []string{"a", "b"},
			OnSuccessMessage: "Success",
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

			notifications := manifest.Notifications{Slack: manifest.Slack{OnFailure: []string{input.SlackChannel}}}
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
				Slack: manifest.Slack{
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

			notifications := manifest.Notifications{Slack: manifest.Slack{OnFailure: []string{input.SlackChannel}}}
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
				Slack: manifest.Slack{
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

			notifications := manifest.Notifications{Slack: manifest.Slack{OnFailure: []string{input.SlackChannel}}}
			notificationsWithSuccess := manifest.Notifications{Slack: manifest.Slack{OnFailure: []string{input.SlackChannel}, OnSuccess: []string{input.SlackChannel}}}
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
			notifications := manifest.Notifications{Slack: manifest.Slack{OnFailure: []string{"#OhNoes", "#AnotherOne"}}}
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

			notificationsWithSuccess := manifest.Notifications{Slack: manifest.Slack{OnFailure: notifications.Slack.OnFailure, OnSuccess: []string{notifications.Slack.OnFailure[0]}}}
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
			notifications := manifest.Notifications{Slack: manifest.Slack{OnFailure: []string{"#test"}}}

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

			notificationsWithSuccess := manifest.Notifications{Slack: manifest.Slack{OnFailure: notifications.Slack.OnFailure, OnSuccess: notifications.Slack.OnFailure}}

			expected := manifest.Manifest{
				Notifications: notifications,
				Tasks: manifest.TaskList{
					manifest.Run{
						Notifications: manifest.Notifications{
							Slack: manifest.Slack{
								OnSuccess: []string{"1"},
								OnFailure: []string{"2"},
							},
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
											Slack: manifest.Slack{
												OnSuccess: []string{"a", "b"},
												OnFailure: []string{"x", "y", "z"},
											},
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

//func TestDefaultNotificationMessages(t *testing.T) {
//	defaultFailureMessage := "failure msg"
//	defaultSuccessMessage := "success msg"
//
//	input := manifest.Manifest{
//		SlackChannel:        "#test",
//		SlackFailureMessage: defaultFailureMessage,
//		SlackSuccessMessage: defaultSuccessMessage,
//		Tasks: manifest.TaskList{
//			manifest.Run{},
//			manifest.DockerPush{},
//			manifest.DeployCF{
//				Notifications: manifest.Notifications{
//					OnSuccess:        []string{"#foo"},
//					OnSuccessMessage: "custom",
//					OnFailureMessage: "custom",
//				},
//			},
//		},
//	}
//
//	updated, _ := NewNotificationsMapper().Apply(input)
//	assert.Equal(t, defaultFailureMessage, updated.Tasks[0].GetNotifications().OnFailureMessage)
//	assert.Equal(t, defaultFailureMessage, updated.Tasks[1].GetNotifications().OnFailureMessage)
//	assert.Equal(t, "custom", updated.Tasks[2].GetNotifications().OnFailureMessage)
//
//	assert.Equal(t, defaultSuccessMessage, updated.Tasks[0].GetNotifications().OnSuccessMessage)
//	assert.Equal(t, defaultSuccessMessage, updated.Tasks[1].GetNotifications().OnSuccessMessage)
//	assert.Equal(t, "custom", updated.Tasks[2].GetNotifications().OnFailureMessage)
//
//}
