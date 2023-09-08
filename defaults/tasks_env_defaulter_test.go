package defaults

import (
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetsCorrectEnvVarsToEmptyVars(t *testing.T) {
	expectedVars := map[string]string{
		"ARTIFACTORY_URL":      Concourse.Artifactory.URL,
		"ARTIFACTORY_USERNAME": Concourse.Artifactory.Username,
		"ARTIFACTORY_PASSWORD": Concourse.Artifactory.Password,
		"RUNNING_IN_CI":        "true",
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
		manifest.Run{Vars: expectedVars},
		manifest.DockerCompose{Vars: expectedVars},
		manifest.Parallel{
			Tasks: manifest.TaskList{
				manifest.Run{Vars: expectedVars},
				manifest.DeployCF{
					PrePromote: manifest.TaskList{
						manifest.Run{Vars: expectedVars},
					},
				},
			},
		},
		manifest.DockerPush{Vars: expectedVars},
		manifest.Parallel{
			Tasks: manifest.TaskList{
				manifest.Sequence{
					Tasks: manifest.TaskList{
						manifest.ConsumerIntegrationTest{Vars: expectedVars},
						manifest.DeployMLModules{},
					},
				},
				manifest.DeployMLModules{},
			},
		},
		manifest.DeployCF{
			PrePromote: manifest.TaskList{
				manifest.DeployMLZip{},
				manifest.ConsumerIntegrationTest{Vars: expectedVars},
				manifest.Run{Vars: expectedVars},
			},
		},
	}

	assert.Equal(t, expected, NewTasksEnvVarsDefaulter().Apply(input, Concourse))
}
