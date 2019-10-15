package defaults

import (
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCallsOutToTaskDefaultersCorrectly(t *testing.T) {
	expectedRun := manifest.Run{
		Name: "a",
	}
	expectedDockerCompose := manifest.DockerCompose{
		Name: "b",
	}
	expectedDockerPush := manifest.DockerPush{
		Name: "c",
	}
	expectedDeployCf := manifest.DeployCF{
		Name: "d",
	}
	expectedConsumerTest := manifest.ConsumerIntegrationTest{
		Name: "e",
	}
	expectedDeployMlZip := manifest.DeployMLZip{
		Name: "f",
	}
	expectedDeployMlModules := manifest.DeployMLModules{
		Name: "g",
	}

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
		manifest.Update{},
		expectedRun,
		expectedDockerCompose,
		manifest.Parallel{
			Tasks: manifest.TaskList{
				expectedRun,
				manifest.DeployCF{
					Name: expectedDeployCf.Name,
					PrePromote: manifest.TaskList{
						expectedRun,
					},
				},
			},
		},
		expectedDockerPush,
		manifest.Parallel{
			Tasks: manifest.TaskList{
				manifest.Sequence{
					Tasks: manifest.TaskList{
						expectedConsumerTest,
						expectedDeployMlModules,
					},
				},
				expectedDeployMlModules,
			},
		},
		manifest.DeployCF{
			Name: expectedDeployCf.Name,
			PrePromote: manifest.TaskList{
				expectedDeployMlModules,
				expectedConsumerTest,
				expectedRun,
			},
		},
	}

	defaulter := tasksDefaulter{
		runDefaulter: func(original manifest.Run, defaults DefaultsNew) (updated manifest.Run) {
			return expectedRun
		},
		dockerComposeDefaulter: func(original manifest.DockerCompose, defaults DefaultsNew) (updated manifest.DockerCompose) {
			return expectedDockerCompose
		},
		dockerPushDefaulter: func(original manifest.DockerPush, defaults DefaultsNew) (updated manifest.DockerPush) {
			return expectedDockerPush
		},
		deployCfDefaulter: func(original manifest.DeployCF, defaults DefaultsNew) (updated manifest.DeployCF) {
			return expectedDeployCf
		},
		consumerIntegrationTestTaskDefaulter: func(original manifest.ConsumerIntegrationTest, defaults DefaultsNew) (updated manifest.ConsumerIntegrationTest) {
			return expectedConsumerTest
		},
		deployMlZipDefaulter: func(original manifest.DeployMLZip, defaults DefaultsNew) (updated manifest.DeployMLZip) {
			return expectedDeployMlZip
		},
		deployMlModulesDefaulter: func(original manifest.DeployMLModules, defaults DefaultsNew) (updated manifest.DeployMLModules) {
			return expectedDeployMlModules
		},
	}

	assert.Equal(t, expected, defaulter.Apply(input, DefaultValuesNew))
}
