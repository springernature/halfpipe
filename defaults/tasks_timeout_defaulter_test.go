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
		manifest.Update{Timeout: DefaultValues.Timeout},
		manifest.Run{Timeout: DefaultValues.Timeout},
		manifest.DockerCompose{Timeout: DefaultValues.Timeout},
		manifest.Parallel{
			Tasks: manifest.TaskList{
				manifest.Run{Timeout: DefaultValues.Timeout},
				manifest.DeployCF{
					Timeout: DefaultValues.Timeout,
					PrePromote: manifest.TaskList{
						manifest.Run{Timeout: DefaultValues.Timeout},
					},
				},
			},
		},
		manifest.DockerPush{Timeout: DefaultValues.Timeout},
		manifest.Parallel{
			Tasks: manifest.TaskList{
				manifest.Sequence{
					Tasks: manifest.TaskList{
						manifest.ConsumerIntegrationTest{Timeout: DefaultValues.Timeout},
						manifest.DeployMLModules{Timeout: DefaultValues.Timeout},
					},
				},
				manifest.DeployMLModules{Timeout: DefaultValues.Timeout},
			},
		},
		manifest.DeployCF{
			Timeout: DefaultValues.Timeout,
			PrePromote: manifest.TaskList{
				manifest.DeployMLModules{Timeout: DefaultValues.Timeout},
				manifest.ConsumerIntegrationTest{Timeout: DefaultValues.Timeout},
				manifest.Run{Timeout: DefaultValues.Timeout},
			},
		},
	}

	assert.Equal(t, expected, NewTasksTimeoutDefaulter().Apply(input, DefaultValues))
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

	assert.Equal(t, input, NewTasksTimeoutDefaulter().Apply(input, DefaultValues))
}
