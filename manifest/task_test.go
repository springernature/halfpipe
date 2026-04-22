package manifest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlatten(t *testing.T) {
	t.Run("When its already flat", func(t *testing.T) {
		taskList := TaskList{
			DeployCF{},
			DockerPush{},
			DockerCompose{},
		}

		assert.Equal(t, taskList, taskList.Flatten())
	})

	t.Run("When its not flat", func(t *testing.T) {
		taskList := TaskList{
			DockerPush{Name: "Task 1"},
			DeployCF{
				Name: "Task 2",
				PrePromote: TaskList{
					DeployMLZip{Name: "Task 3"},
				}},
			DockerCompose{Name: "Task 4"},
			Sequence{
				Tasks: TaskList{
					Run{Name: "Task 5"},
				},
			},
			Parallel{
				Tasks: TaskList{
					Sequence{
						Tasks: TaskList{
							DeployCF{
								Name: "Task 6",
								PrePromote: TaskList{
									Run{Name: "Task 7"},
								},
							},
						},
					},
				},
			},
		}
		expected := TaskList{
			DockerPush{Name: "Task 1"},
			DeployCF{Name: "Task 2"},
			DeployMLZip{Name: "Task 3"},
			DockerCompose{Name: "Task 4"},
			Run{Name: "Task 5"},
			DeployCF{Name: "Task 6"},
			Run{Name: "Task 7"},
		}

		assert.Equal(t, expected, taskList.Flatten())
	})

}
