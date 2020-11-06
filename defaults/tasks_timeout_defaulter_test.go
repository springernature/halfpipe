package defaults

import (
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetsCorrectTimeout(t *testing.T) {
	input := manifest.TaskList{
		manifest.Update{},
		manifest.Run{},
		manifest.DockerCompose{},
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
		manifest.Update{Timeout: Concourse.Aux.Timeout},
		manifest.Run{Timeout: Concourse.Aux.Timeout},
		manifest.DockerCompose{Timeout: Concourse.Aux.Timeout},
		manifest.Parallel{
			Tasks: manifest.TaskList{
				manifest.Run{Timeout: Concourse.Aux.Timeout},
				manifest.DeployCF{
					Timeout: Concourse.Aux.Timeout,
					PrePromote: manifest.TaskList{
						manifest.Run{Timeout: Concourse.Aux.Timeout},
					},
				},
			},
		},
		manifest.DockerPush{Timeout: Concourse.Aux.Timeout},
		manifest.Parallel{
			Tasks: manifest.TaskList{
				manifest.Sequence{
					Tasks: manifest.TaskList{
						manifest.ConsumerIntegrationTest{Timeout: Concourse.Aux.Timeout},
						manifest.DeployMLModules{Timeout: Concourse.Aux.Timeout},
					},
				},
				manifest.DeployMLModules{Timeout: Concourse.Aux.Timeout},
			},
		},
		manifest.DeployCF{
			Timeout: Concourse.Aux.Timeout,
			PrePromote: manifest.TaskList{
				manifest.DeployMLModules{Timeout: Concourse.Aux.Timeout},
				manifest.ConsumerIntegrationTest{Timeout: Concourse.Aux.Timeout},
				manifest.Run{Timeout: Concourse.Aux.Timeout},
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
