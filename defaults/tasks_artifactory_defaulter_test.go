package defaults

import (
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetsCorrectArtifactoryVarsToEmptyVars(t *testing.T) {
	expectedVars := map[string]string{
		"ARTIFACTORY_URL":      DefaultValues.Artifactory.URL,
		"ARTIFACTORY_USERNAME": DefaultValues.Artifactory.Username,
		"ARTIFACTORY_PASSWORD": DefaultValues.Artifactory.Password,
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

	assert.Equal(t, expected, NewTasksArtifactoryVarsDefaulter().Apply(input, DefaultValues))

}

func TestSetsCorrectArtifactoryVarsToAlreadyPresentVars(t *testing.T) {
	//expectedNumberOfTasksThatSupportTimeout := 14
	//expectedTimeout := "1337h"
	//
	//input := manifest.TaskList{
	//	manifest.Update{Timeout: expectedTimeout},
	//	manifest.Run{Timeout: expectedTimeout},
	//	manifest.DockerCompose{Timeout: expectedTimeout},
	//	manifest.Parallel{
	//		Tasks: manifest.TaskList{
	//			manifest.Run{Timeout: expectedTimeout},
	//			manifest.DeployCF{
	//				Timeout: expectedTimeout,
	//				PrePromote: manifest.TaskList{
	//					manifest.Run{Timeout: expectedTimeout},
	//				},
	//			},
	//		},
	//	},
	//	manifest.DockerPush{Timeout: expectedTimeout},
	//	manifest.Parallel{
	//		Tasks: manifest.TaskList{
	//			manifest.Sequence{
	//				Tasks: manifest.TaskList{
	//					manifest.ConsumerIntegrationTest{Timeout: expectedTimeout},
	//					manifest.DeployMLModules{Timeout: expectedTimeout},
	//				},
	//			},
	//			manifest.DeployMLModules{Timeout: expectedTimeout},
	//		},
	//	},
	//	manifest.DeployCF{
	//		Timeout: expectedTimeout,
	//		PrePromote: manifest.TaskList{
	//			manifest.DeployMLModules{Timeout: expectedTimeout},
	//			manifest.ConsumerIntegrationTest{Timeout: expectedTimeout},
	//			manifest.Run{Timeout: expectedTimeout},
	//		},
	//	},
	//}
	//
	//updated := NewTasksTimeoutDefaulter().Apply(input, DefaultValues)
	//
	//var numberOfChecked int
	//var check func(taskList manifest.TaskList)
	//check = func(taskList manifest.TaskList) {
	//	for _, task := range taskList {
	//		switch task := task.(type) {
	//		case manifest.Parallel:
	//			check(task.Tasks)
	//		case manifest.Sequence:
	//			check(task.Tasks)
	//		default:
	//			numberOfChecked++
	//			assert.Equal(t, expectedTimeout, task.GetTimeout())
	//			if deployTask, isDeployTask := task.(manifest.DeployCF); isDeployTask {
	//				check(deployTask.PrePromote)
	//			}
	//		}
	//	}
	//}
	//check(updated)
	//
	//assert.Equal(t, expectedNumberOfTasksThatSupportTimeout, numberOfChecked)
}
