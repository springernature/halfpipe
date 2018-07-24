package pipeline

import (
	"testing"

	cfManifest "code.cloudfoundry.org/cli/util/manifest"

	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func testPipeline() pipeline {
	return pipeline{
		rManifest: func(s string) ([]cfManifest.Application, error) {
			return []cfManifest.Application{
				{
					Name:   "test-name",
					Routes: []string{"test-route"},
				},
			}, nil
		},
	}
}

func TestRenderWithTriggerTrueAndPassedOnPreviousTask(t *testing.T) {
	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.Run{Script: "asd.sh"},
			manifest.DeployCF{ManualTrigger: true},
			manifest.DockerPush{},
		},
	}
	config := testPipeline().Render(man)

	assert.Nil(t, config.Jobs[0].Plan[0].Passed)
	assert.Equal(t, config.Jobs[0].Plan[0].Trigger, true)

	assert.Equal(t, config.Jobs[1].Plan[0].Passed[0], config.Jobs[0].Name)
	assert.Equal(t, config.Jobs[1].Plan[0].Trigger, false)

	assert.Equal(t, config.Jobs[2].Plan[0].Passed[0], config.Jobs[1].Name)
	assert.Equal(t, config.Jobs[2].Plan[0].Trigger, true)
}

func TestPassedOnPreviousTaskWithAutoUpdate(t *testing.T) {
	man := manifest.Manifest{
		AutoUpdate: true,
		Tasks: []manifest.Task{
			manifest.Run{Script: "asd.sh"},
			manifest.DeployCF{ManualTrigger: true},
			manifest.DockerPush{},
		},
	}
	config := testPipeline().Render(man)

	assert.Len(t, config.Jobs, 4)

	assert.Nil(t, config.Jobs[0].Plan[0].Passed)
	assert.Equal(t, config.Jobs[0].Plan[0].Trigger, true)

	assert.Equal(t, config.Jobs[1].Plan[0].Passed[0], config.Jobs[0].Name)
	assert.Equal(t, config.Jobs[1].Plan[0].Trigger, true)

	assert.Equal(t, config.Jobs[2].Plan[0].Passed[0], config.Jobs[1].Name)
	assert.Equal(t, config.Jobs[2].Plan[0].Trigger, false)

	assert.Equal(t, config.Jobs[3].Plan[0].Passed[0], config.Jobs[2].Name)
	assert.Equal(t, config.Jobs[3].Plan[0].Trigger, true)
}

func TestRenderWithParallelTasks(t *testing.T) {
	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.Run{Name: "Build", Script: "asd.sh"},

			manifest.DeployCF{Name: "Deploy", Parallel: true},
			manifest.DockerPush{Name: "Push", Parallel: true},

			manifest.Run{Name: "Smoke Test", Script: "asd.sh"},

			manifest.DeployCF{Name: "Deploy 2", Parallel: true},
			manifest.DockerPush{Name: "Push 2", Parallel: true},

			manifest.Run{Name: "Smoke Test 2", Script: "asd.sh"},
		},
	}
	config := testPipeline().Render(man)

	assert.Nil(t, config.Jobs[0].Plan[0].Passed)

	assert.Equal(t, "Build", config.Jobs[1].Plan[0].Passed[0])
	assert.Equal(t, "Build", config.Jobs[2].Plan[0].Passed[0])

	assert.Equal(t, []string{"Deploy", "Push"}, config.Jobs[3].Plan[0].Passed)

	assert.Equal(t, "Smoke Test", config.Jobs[4].Plan[0].Passed[0])
	assert.Equal(t, "Smoke Test", config.Jobs[5].Plan[0].Passed[0])

	assert.Equal(t, []string{"Deploy 2", "Push 2"}, config.Jobs[6].Plan[0].Passed)

}

func TestRenderWithParallelOnFirstTasks(t *testing.T) {
	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.Run{Name: "Build", Script: "asd.sh", Parallel: true},
			manifest.DeployCF{Name: "Deploy", Parallel: true},

			manifest.DockerPush{Name: "Push"},
		},
	}
	config := testPipeline().Render(man)

	assert.Nil(t, config.Jobs[0].Plan[0].Passed)
	assert.Nil(t, config.Jobs[1].Plan[0].Passed)

	assert.Equal(t, []string{"Build", "Deploy"}, config.Jobs[2].Plan[0].Passed)

}

func TestRenderWithParallelOnFirstTasksWithAutoUpdate(t *testing.T) {
	man := manifest.Manifest{
		AutoUpdate: true,
		Tasks: []manifest.Task{
			manifest.Run{Name: "Build", Script: "asd.sh", Parallel: true},
			manifest.DeployCF{Name: "Deploy", Parallel: true},

			manifest.DockerPush{Name: "Push"},
		},
	}
	config := testPipeline().Render(man)

	assert.Equal(t, "Update Pipeline", config.Jobs[1].Plan[0].Passed[0])
	assert.Equal(t, "Update Pipeline", config.Jobs[2].Plan[0].Passed[0])

	assert.Equal(t, []string{"Build", "Deploy"}, config.Jobs[3].Plan[0].Passed)

}

func TestRenderTwoGetJobsAsAggregate(t *testing.T) {
	man := manifest.Manifest{
		AutoUpdate: true,
		Tasks: []manifest.Task{
			manifest.Run{Name: "Build", Script: "asd.sh", Parallel: true},
			manifest.DeployCF{Name: "Deploy", Parallel: true},

			manifest.DockerPush{Name: "Push"},
		},
	}
	config := testPipeline().Render(man)

	assert.Equal(t, "Update Pipeline", config.Jobs[1].Plan[0].Passed[0])
	assert.Equal(t, "Update Pipeline", config.Jobs[2].Plan[0].Passed[0])

	assert.Equal(t, []string{"Build", "Deploy"}, config.Jobs[3].Plan[0].Passed)

}
