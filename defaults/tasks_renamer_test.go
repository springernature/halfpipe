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
		manifest.Parallel{
			Tasks: manifest.TaskList{
				manifest.Sequence{
					Tasks: manifest.TaskList{
						manifest.Run{Name: "test"},
						manifest.Run{Name: "Something new"},
					},
				},
			},
		},
	}

	expected := manifest.TaskList{
		manifest.Run{Name: "run asd.sh", Script: "asd.sh"},
		manifest.Parallel{
			Tasks: manifest.TaskList{
				manifest.Run{Name: "run asd.sh (1)", Script: "asd.sh"},
				manifest.Run{Name: "test", Script: "asd.sh"},
				manifest.Run{Name: "test (1)", Script: "asd.sh"},
			},
		},
		manifest.Run{Name: "run asd.sh (2)", Script: "asd.sh"},
		manifest.Run{Name: "test (2)", Script: "asd.sh"},
		manifest.Run{Name: "test (3)", Script: "fgh.sh"},
		manifest.DeployCF{
			Name: "deploy-cf",
			PrePromote: manifest.TaskList{
				manifest.Run{Name: "test", Script: "asd.sh"},
				manifest.Run{Name: "run asd.sh", Script: "asd.sh"},
				manifest.Run{Name: "test (1)", Script: "asd.sh"},
				manifest.Run{Name: "run asd.sh (1)", Script: "asd.sh"},
			},
		},
		manifest.DeployCF{
			Name: "deploy-cf (1)",
			PrePromote: manifest.TaskList{
				manifest.Run{Name: "test", Script: "asd.sh"},
				manifest.Run{Name: "run asd.sh", Script: "asd.sh"},
				manifest.Run{Name: "test (1)", Script: "asd.sh"},
				manifest.Run{Name: "run asd.sh (1)", Script: "asd.sh"},
			},
		},
		manifest.DeployCF{Name: "deploy-cf (2)"},
		manifest.DockerPush{Name: "docker-push"},
		manifest.DockerPush{Name: "docker-push (1)"},
		manifest.DockerPush{Name: "docker-push (2)"},
		manifest.Parallel{
			Tasks: manifest.TaskList{
				manifest.Sequence{
					Tasks: manifest.TaskList{
						manifest.Run{Name: "test (4)"},
						manifest.Run{Name: "Something new"},
					},
				},
			},
		},
	}

	assert.Equal(t, expected, NewTasksRenamer().Apply(tasks))
}
