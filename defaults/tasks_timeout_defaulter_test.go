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

	tb := func(timeout string) manifest.TaskBase { return manifest.TaskBase{Timeout: timeout} }

	expected := manifest.TaskList{
		manifest.DockerCompose{TaskBase: tb(Concourse.Timeout)},
		manifest.Update{TaskBase: tb(Concourse.Timeout)},
		manifest.Run{TaskBase: tb(Concourse.Timeout)},
		manifest.DeployKatee{TaskBase: tb(Concourse.Timeout)},
		manifest.Parallel{
			Tasks: manifest.TaskList{
				manifest.Run{TaskBase: tb(Concourse.Timeout)},
				manifest.DeployCF{
					TaskBase: tb(Concourse.Timeout),
					PrePromote: manifest.TaskList{
						manifest.Run{TaskBase: tb(Concourse.Timeout)},
					},
				},
			},
		},
		manifest.DockerPush{TaskBase: tb(Concourse.Timeout)},
		manifest.Parallel{
			Tasks: manifest.TaskList{
				manifest.Sequence{
					Tasks: manifest.TaskList{
						manifest.ConsumerIntegrationTest{TaskBase: tb(Concourse.Timeout)},
						manifest.DeployMLModules{TaskBase: tb(Concourse.Timeout)},
					},
				},
				manifest.DeployMLModules{TaskBase: tb(Concourse.Timeout)},
			},
		},
		manifest.DeployCF{
			TaskBase: tb(Concourse.Timeout),
			PrePromote: manifest.TaskList{
				manifest.DeployMLModules{TaskBase: tb(Concourse.Timeout)},
				manifest.ConsumerIntegrationTest{TaskBase: tb(Concourse.Timeout)},
				manifest.Run{TaskBase: tb(Concourse.Timeout)},
			},
		},
	}

	assert.Equal(t, expected, NewTasksTimeoutDefaulter().Apply(input, Concourse))
}

func TestDoesntOverrideTimeout(t *testing.T) {
	expectedTimeout := "1337h"
	tb := func() manifest.TaskBase { return manifest.TaskBase{Timeout: expectedTimeout} }

	input := manifest.TaskList{
		manifest.Update{TaskBase: tb()},
		manifest.Run{TaskBase: tb()},
		manifest.DockerCompose{TaskBase: tb()},
		manifest.Parallel{
			Tasks: manifest.TaskList{
				manifest.Run{TaskBase: tb()},
				manifest.DeployCF{
					TaskBase: tb(),
					PrePromote: manifest.TaskList{
						manifest.Run{TaskBase: tb()},
					},
				},
			},
		},
		manifest.DockerPush{TaskBase: tb()},
		manifest.Parallel{
			Tasks: manifest.TaskList{
				manifest.Sequence{
					Tasks: manifest.TaskList{
						manifest.ConsumerIntegrationTest{TaskBase: tb()},
						manifest.DeployMLModules{TaskBase: tb()},
					},
				},
				manifest.DeployMLModules{TaskBase: tb()},
			},
		},
		manifest.DeployCF{
			TaskBase: tb(),
			PrePromote: manifest.TaskList{
				manifest.DeployMLModules{TaskBase: tb()},
				manifest.ConsumerIntegrationTest{TaskBase: tb()},
				manifest.Run{TaskBase: tb()},
			},
		},
	}

	assert.Equal(t, input, NewTasksTimeoutDefaulter().Apply(input, Concourse))
}
