package defaults

import (
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBuildHistoryDefaulterDoesTheNeedful(t *testing.T) {
	defaults := Defaults{BuildHistory: 1337}

	taskList := manifest.TaskList{
		manifest.Run{},
		manifest.DockerCompose{},
		manifest.DeployCF{},
		manifest.DockerPush{BuildHistory: 9000},
		manifest.ConsumerIntegrationTest{},
		manifest.DeployMLZip{},
		manifest.DeployMLModules{},
		manifest.Parallel{
			Tasks: manifest.TaskList{
				manifest.Run{},
				manifest.Sequence{
					Tasks: manifest.TaskList{
						manifest.Run{},
						manifest.DockerPush{BuildHistory: 123},
					},
				},
			},
		},
		manifest.Sequence{
			Tasks: manifest.TaskList{
				manifest.Run{},
				manifest.Run{BuildHistory: 987},
				manifest.Parallel{
					Tasks: manifest.TaskList{
						manifest.Run{BuildHistory: 444},
						manifest.Run{},
					},
				},
			},
		},
	}

	expected := manifest.TaskList{
		manifest.Run{BuildHistory: 1337},
		manifest.DockerCompose{BuildHistory: 1337},
		manifest.DeployCF{BuildHistory: 1337},
		manifest.DockerPush{BuildHistory: 9000},
		manifest.ConsumerIntegrationTest{BuildHistory: 1337},
		manifest.DeployMLZip{BuildHistory: 1337},
		manifest.DeployMLModules{BuildHistory: 1337},
		manifest.Parallel{
			Tasks: manifest.TaskList{
				manifest.Run{BuildHistory: 1337},
				manifest.Sequence{
					Tasks: manifest.TaskList{
						manifest.Run{BuildHistory: 1337},
						manifest.DockerPush{BuildHistory: 123},
					},
				},
			},
		},
		manifest.Sequence{
			Tasks: manifest.TaskList{
				manifest.Run{BuildHistory: 1337},
				manifest.Run{BuildHistory: 987},
				manifest.Parallel{
					Tasks: manifest.TaskList{
						manifest.Run{BuildHistory: 444},
						manifest.Run{BuildHistory: 1337},
					},
				},
			},
		},
	}

	assert.Equal(t, expected, NewBuildHistoryDefaulter().Apply(taskList, defaults))
}
