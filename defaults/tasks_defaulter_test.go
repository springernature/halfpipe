package defaults

import (
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

type testTasksRenamer struct {
	apply func(original manifest.TaskList) (updated manifest.TaskList)
}

func (t testTasksRenamer) Apply(original manifest.TaskList) (updated manifest.TaskList) {
	return t.apply(original)
}

type testTasksTimeoutDefaulter struct {
	apply func(original manifest.TaskList, defaults Defaults) (updated manifest.TaskList)
}

func (t testTasksTimeoutDefaulter) Apply(original manifest.TaskList, defaults Defaults) (updated manifest.TaskList) {
	return t.apply(original, defaults)
}

type testTasksArtifactoryVarsDefaulter struct {
	apply func(original manifest.TaskList, defaults Defaults) (updated manifest.TaskList)
}

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
				manifest.DeployMLZip{},
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
				expectedDeployMlZip,
				expectedConsumerTest,
				expectedRun,
			},
		},
	}

	var tasksRenamerCalled bool
	var tasksTimeoutDefaulterCalled bool
	var tasksArtifactoryVarsDefaulterCalled bool
	defaulter := tasksDefaulter{
		tasksRenamer: testTasksRenamer{apply: func(original manifest.TaskList) (updated manifest.TaskList) {
			tasksRenamerCalled = true
			return original
		}},
		tasksTimeoutDefaulter: testTasksTimeoutDefaulter{apply: func(original manifest.TaskList, defaults Defaults) (updated manifest.TaskList) {
			tasksTimeoutDefaulterCalled = true
			return original
		}},
		tasksArtifactoryVarsDefaulter: testTasksArtifactoryVarsDefaulter{apply: func(original manifest.TaskList, defaults Defaults) (updated manifest.TaskList) {
			tasksArtifactoryVarsDefaulterCalled = true
			return original
		}},
		runDefaulter: func(original manifest.Run, defaults Defaults) (updated manifest.Run) {
			return expectedRun
		},
		dockerComposeDefaulter: func(original manifest.DockerCompose, defaults Defaults) (updated manifest.DockerCompose) {
			return expectedDockerCompose
		},
		dockerPushDefaulter: func(original manifest.DockerPush, defaults Defaults) (updated manifest.DockerPush) {
			return expectedDockerPush
		},
		deployCfDefaulter: func(original manifest.DeployCF, defaults Defaults, man manifest.Manifest) (updated manifest.DeployCF) {
			return expectedDeployCf
		},
		consumerIntegrationTestTaskDefaulter: func(original manifest.ConsumerIntegrationTest, defaults Defaults) (updated manifest.ConsumerIntegrationTest) {
			return expectedConsumerTest
		},
		deployMlZipDefaulter: func(original manifest.DeployMLZip, defaults Defaults) (updated manifest.DeployMLZip) {
			return expectedDeployMlZip
		},
		deployMlModulesDefaulter: func(original manifest.DeployMLModules, defaults Defaults) (updated manifest.DeployMLModules) {
			return expectedDeployMlModules
		},
	}

	assert.Equal(t, expected, defaulter.Apply(input, DefaultValues, manifest.Manifest{}))
	assert.True(t, tasksRenamerCalled)
	assert.True(t, tasksTimeoutDefaulterCalled)
	assert.True(t, tasksArtifactoryVarsDefaulterCalled)
}
