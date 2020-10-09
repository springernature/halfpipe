package concourse

import (
	"github.com/concourse/concourse/atc"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRenderWithParallelTasks(t *testing.T) {
	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.GitTrigger{},
		},

		Tasks: []manifest.Task{
			manifest.Run{Name: "Build", Script: "asd.sh"},

			manifest.Parallel{
				Tasks: manifest.TaskList{
					manifest.DeployCF{Name: "Deploy"},
					manifest.DockerPush{Name: "Push"},
				},
			},
			manifest.Run{Name: "Smoke Test", Script: "asd.sh"},
			manifest.Parallel{
				Tasks: manifest.TaskList{
					manifest.DeployCF{Name: "Deploy 2"},
					manifest.DockerPush{Name: "Push 2"},
				},
			},
			manifest.Run{Name: "Smoke Test 2", Script: "asd.sh"},
		},
	}
	config := testPipeline().Render(man)

	assert.Nil(t, (config.Jobs[0].Plan[0].InParallel.Steps)[0].Passed)

	assert.Equal(t, "Build", (config.Jobs[1].Plan[0].InParallel.Steps)[0].Passed[0])
	assert.Equal(t, "Build", (config.Jobs[2].Plan[0].InParallel.Steps)[0].Passed[0])

	assert.Equal(t, []string{"Deploy", "Push"}, (config.Jobs[3].Plan[0].InParallel.Steps)[0].Passed)

	assert.Equal(t, "Smoke Test", (config.Jobs[4].Plan[0].InParallel.Steps)[0].Passed[0])
	assert.Equal(t, "Smoke Test", (config.Jobs[5].Plan[0].InParallel.Steps)[0].Passed[0])

	assert.Equal(t, []string{"Deploy 2", "Push 2"}, (config.Jobs[6].Plan[0].InParallel.Steps)[0].Passed)

}

func TestRenderWithParallelAndSeqTasks(t *testing.T) {
	getPassed := func(jobConfig atc.JobConfig) []string {
		return (jobConfig.Plan[0].InParallel.Steps)[0].Passed
	}

	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.GitTrigger{},
		},

		Tasks: []manifest.Task{
			manifest.Run{Name: "a"},
			manifest.Parallel{
				Tasks: manifest.TaskList{
					manifest.DeployCF{Name: "b1"},
					manifest.DockerPush{Name: "b2"},
				},
			},
			manifest.Parallel{
				Tasks: manifest.TaskList{
					manifest.Sequence{
						Tasks: manifest.TaskList{
							manifest.DeployCF{Name: "c-a-1"},
							manifest.DockerPush{Name: "c-a-2"},
						},
					},
					manifest.Sequence{
						Tasks: manifest.TaskList{
							manifest.DeployCF{Name: "c-b-1"},
							manifest.DockerPush{Name: "c-b-2"},
							manifest.DockerPush{Name: "c-b-3"},
						},
					},
					manifest.Run{Name: "c-c"},
				},
			},
			manifest.Run{Name: "d"},
		},
	}
	config := testPipeline().Render(man)

	t.Run("first task", func(t *testing.T) {
		assert.Empty(t, getPassed(config.Jobs[0]))
	})

	t.Run("first parallel", func(t *testing.T) {
		assert.Equal(t, []string{config.Jobs[0].Name}, getPassed(config.Jobs[1]))
		assert.Equal(t, []string{config.Jobs[0].Name}, getPassed(config.Jobs[2]))
	})

	t.Run("second parallel", func(t *testing.T) {
		t.Run("first sequence", func(t *testing.T) {
			assert.Equal(t, []string{config.Jobs[1].Name, config.Jobs[2].Name}, getPassed(config.Jobs[3]))
			assert.Equal(t, []string{config.Jobs[3].Name}, getPassed(config.Jobs[4]))
		})
		t.Run("second sequence", func(t *testing.T) {
			assert.Equal(t, []string{config.Jobs[1].Name, config.Jobs[2].Name}, getPassed(config.Jobs[5]))
			assert.Equal(t, []string{config.Jobs[5].Name}, getPassed(config.Jobs[6]))
			assert.Equal(t, []string{config.Jobs[6].Name}, getPassed(config.Jobs[7]))
		})
		t.Run("normal run", func(t *testing.T) {
			assert.Equal(t, []string{config.Jobs[1].Name, config.Jobs[2].Name}, getPassed(config.Jobs[8]))
		})
	})

	t.Run("last task", func(t *testing.T) {
		assert.Equal(t, []string{config.Jobs[4].Name, config.Jobs[7].Name, config.Jobs[8].Name}, getPassed(config.Jobs[9]))
	})
}

func TestRenderWithParallelOnFirstTasks(t *testing.T) {
	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.GitTrigger{},
		},

		Tasks: []manifest.Task{
			manifest.Parallel{
				Tasks: manifest.TaskList{
					manifest.Run{Name: "Build", Script: "asd.sh"},
					manifest.DeployCF{Name: "Deploy"},
				},
			},

			manifest.DockerPush{Name: "Push"},
		},
	}
	config := testPipeline().Render(man)

	assert.Nil(t, config.Jobs[0].Plan[0].Passed)
	assert.Nil(t, config.Jobs[1].Plan[0].Passed)

	assert.Equal(t, []string{"Build", "Deploy"}, (config.Jobs[2].Plan[0].InParallel.Steps)[0].Passed)
}
