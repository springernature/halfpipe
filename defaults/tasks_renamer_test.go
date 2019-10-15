package defaults

import (
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetsNames(t *testing.T) {
	tasks := manifest.TaskList{
		manifest.Run{Script: "asd.sh"},
		manifest.Parallel{
			Tasks: manifest.TaskList{
				manifest.Run{Script: "asd.sh"},
				manifest.Run{Name: "test", Script: "asd.sh"},
				manifest.Run{Name: "test", Script: "asd.sh"},
			},
		},
		manifest.Run{Script: "asd.sh"},
		manifest.Run{Name: "test", Script: "asd.sh"},
		manifest.Run{Name: "test", Script: "fgh.sh"},
		manifest.DeployCF{
			Name: "deploy-cf",
			PrePromote: manifest.TaskList{
				manifest.Run{Name: "test", Script: "asd.sh"},
				manifest.Run{Script: "asd.sh"},
				manifest.Run{Name: "test", Script: "asd.sh"},
				manifest.Run{Script: "asd.sh"},
			},
		},
		manifest.DeployCF{
			Name: "deploy-cf",
			PrePromote: manifest.TaskList{
				manifest.Run{Name: "test", Script: "asd.sh"},
				manifest.Run{Script: "asd.sh"},
				manifest.Run{Name: "test", Script: "asd.sh"},
				manifest.Run{Script: "asd.sh"},
			},
		},
		manifest.DeployCF{},
		manifest.DockerPush{},
		manifest.DockerPush{},
		manifest.DockerPush{},
		manifest.DeployCF{Name: "deploy to dev"},
		manifest.DeployCF{Name: "deploy to dev"},
		manifest.DockerPush{Name: "push to docker hub"},
		manifest.DockerPush{Name: "push to docker hub"},
	}

	expectedWithoutAllTheOtherFields := manifest.TaskList{
			manifest.Run{Name: "run asd.sh"},
			manifest.Parallel{
				Tasks: manifest.TaskList{
					manifest.Run{Name: "run asd.sh (1)"},
					manifest.Run{Name: "test"},
					manifest.Run{Name: "test (1)"},
				},
			},
			manifest.Run{Name: "run asd.sh (2)"},
			manifest.Run{Name: "test (2)"},
			manifest.Run{Name: "test (3)"},
			manifest.DeployCF{
				Name: "deploy-cf",
				PrePromote: manifest.TaskList{
					manifest.Run{Name: "test"},
					manifest.Run{Name: "run asd.sh"},
					manifest.Run{Name: "test (1)"},
					manifest.Run{Name: "run asd.sh (1)"},
				},
			},
			manifest.DeployCF{
				Name: "deploy-cf (1)",
				PrePromote: manifest.TaskList{
					manifest.Run{Name: "test"},
					manifest.Run{Name: "run asd.sh"},
					manifest.Run{Name: "test (1)"},
					manifest.Run{Name: "run asd.sh (1)"},
				},
			},
			manifest.DeployCF{Name: "deploy-cf (2)"},
			manifest.DockerPush{Name: "docker-push"},
			manifest.DockerPush{Name: "docker-push (1)"},
			manifest.DockerPush{Name: "docker-push (2)"},
			manifest.DeployCF{Name: "deploy to dev"},
			manifest.DeployCF{Name: "deploy to dev (1)"},
			manifest.DockerPush{Name: "push to docker hub"},
			manifest.DockerPush{Name: "push to docker hub (1)"},
	}

	updated := NewTasksRenamer().Apply(tasks)

	assert.Len(t, expectedWithoutAllTheOtherFields, len(updated))
	for i, updatedTask := range updated {
		if updateParallelTask, isParallelTask := updatedTask.(manifest.Parallel); isParallelTask {
			expectedParallelTask := expectedWithoutAllTheOtherFields[i].(manifest.Parallel)
			for pi, pTask := range updateParallelTask.Tasks {
				assert.Equal(t, expectedParallelTask.Tasks[pi].GetName(), pTask.GetName())
			}
		} else {
			assert.Equal(t, expectedWithoutAllTheOtherFields[i].GetName(), updatedTask.GetName())
			if updatedDeployCf, isDeployCf := updatedTask.(manifest.DeployCF); isDeployCf {
				expectedDeployCf := expectedWithoutAllTheOtherFields[i].(manifest.DeployCF)
				for ppi, ppTask := range updatedDeployCf.PrePromote {
					assert.Equal(t, expectedDeployCf.PrePromote[ppi].GetName(), ppTask.GetName())
				}
			}
		}
	}
}
