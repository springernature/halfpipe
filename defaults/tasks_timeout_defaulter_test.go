package defaults

import (
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetsCorrectTimeout(t *testing.T) {
	input := manifest.TaskList{
		manifest.DockerCompose{},
		manifest.Update{},
		manifest.Run{},
		manifest.DeployKatee{},
		manifest.Parallel{
			Tasks: manifest.TaskList{
				manifest.Run{},
				manifest.DeployCF{
					PrePromote: manifest.TaskList{
						manifest.Run{},
					},
				},
			},
		},
		manifest.DockerPush{},
		manifest.Parallel{
			Tasks: manifest.TaskList{
				manifest.Sequence{
					Tasks: manifest.TaskList{
						manifest.ConsumerIntegrationTest{},
						manifest.DeployMLModules{},
					},
				},
				manifest.DeployMLModules{},
			},
		},
		manifest.DeployCF{
			PrePromote: manifest.TaskList{
				manifest.DeployMLModules{},
				manifest.ConsumerIntegrationTest{},
				manifest.Run{},
			},
		},
	}

	expected := manifest.TaskList{
		manifest.DockerCompose{Timeout: Concourse.Timeout},
		manifest.Update{Timeout: Concourse.Timeout},
		manifest.Run{Timeout: Concourse.Timeout},
		manifest.DeployKatee{Timeout: Concourse.Timeout},
		manifest.Parallel{
			Tasks: manifest.TaskList{
				manifest.Run{Timeout: Concourse.Timeout},
				manifest.DeployCF{
					Timeout: Concourse.Timeout,
					PrePromote: manifest.TaskList{
						manifest.Run{Timeout: Concourse.Timeout},
					},
				},
			},
		},
		manifest.DockerPush{Timeout: Concourse.Timeout},
		manifest.Parallel{
			Tasks: manifest.TaskList{
				manifest.Sequence{
					Tasks: manifest.TaskList{
						manifest.ConsumerIntegrationTest{Timeout: Concourse.Timeout},
						manifest.DeployMLModules{Timeout: Concourse.Timeout},
					},
				},
				manifest.DeployMLModules{Timeout: Concourse.Timeout},
			},
		},
		manifest.DeployCF{
			Timeout: Concourse.Timeout,
			PrePromote: manifest.TaskList{
				manifest.DeployMLModules{Timeout: Concourse.Timeout},
				manifest.ConsumerIntegrationTest{Timeout: Concourse.Timeout},
				manifest.Run{Timeout: Concourse.Timeout},
			},
		},
	}

	assert.Equal(t, expected, NewTasksTimeoutDefaulter().Apply(input, Concourse))
}

func TestDoesntOverrideTimeout(t *testing.T) {
	expectedTimeout := "1337h"
	input := manifest.TaskList{
		manifest.Update{Timeout: expectedTimeout},
		manifest.Run{Timeout: expectedTimeout},
		manifest.DockerCompose{Timeout: expectedTimeout},
		manifest.Parallel{
			Tasks: manifest.TaskList{
				manifest.Run{Timeout: expectedTimeout},
				manifest.DeployCF{
					Timeout: expectedTimeout,
					PrePromote: manifest.TaskList{
						manifest.Run{Timeout: expectedTimeout},
					},
				},
			},
		},
		manifest.DockerPush{Timeout: expectedTimeout},
		manifest.Parallel{
			Tasks: manifest.TaskList{
				manifest.Sequence{
					Tasks: manifest.TaskList{
						manifest.ConsumerIntegrationTest{Timeout: expectedTimeout},
						manifest.DeployMLModules{Timeout: expectedTimeout},
					},
				},
				manifest.DeployMLModules{Timeout: expectedTimeout},
			},
		},
		manifest.DeployCF{
			Timeout: expectedTimeout,
			PrePromote: manifest.TaskList{
				manifest.DeployMLModules{Timeout: expectedTimeout},
				manifest.ConsumerIntegrationTest{Timeout: expectedTimeout},
				manifest.Run{Timeout: expectedTimeout},
			},
		},
	}

	assert.Equal(t, input, NewTasksTimeoutDefaulter().Apply(input, Concourse))
}
