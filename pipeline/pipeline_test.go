package pipeline

import (
	"testing"

	cfManifest "code.cloudfoundry.org/cli/util/manifest"

	"github.com/cloudfoundry/bosh-cli/director/template"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func testPipeline() pipeline {
	cfManifestReader := func(pathToManifest string, pathsToVarsFiles []string, vars []template.VarKV) ([]cfManifest.Application, error) {
		return []cfManifest.Application{
			{
				Name:   "test-name",
				Routes: []string{"test-route"},
			},
		}, nil
	}

	return NewPipeline(cfManifestReader, afero.Afero{Fs: afero.NewMemMapFs()})
}

func TestRenderWithGitTriggerTrueAndPassedOnPreviousTask(t *testing.T) {
	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.GitTrigger{},
		},
		Tasks: []manifest.Task{
			manifest.Run{Name: "t1", Script: "asd.sh"},
			manifest.DeployCF{Name: "t2", ManualTrigger: true},
			manifest.DockerPush{Name: "t3"},
		},
	}
	config := testPipeline().Render(man)

	assert.Nil(t, config.Jobs[0].Plan[0].Passed)
	getGitStep := (config.Jobs[0].Plan[0].InParallel.Steps)[0]
	assert.Equal(t, gitName, getGitStep.Name())
	assert.True(t, getGitStep.Trigger)

	getGitStep = (config.Jobs[1].Plan[0].InParallel.Steps)[0]
	assert.Equal(t, config.Jobs[0].Name, getGitStep.Passed[0])
	assert.False(t, getGitStep.Trigger)

	getGitStep = (config.Jobs[2].Plan[0].InParallel.Steps)[0]
	assert.Equal(t, config.Jobs[1].Name, getGitStep.Passed[0])
	assert.True(t, getGitStep.Trigger)
}

func TestRenderWithGitManualTrigger(t *testing.T) {
	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.GitTrigger{
				ManualTrigger: true,
			},
		},
		Tasks: []manifest.Task{
			manifest.Run{Name: "t1", Script: "asd.sh"},
			manifest.DeployCF{Name: "t2", ManualTrigger: true},
			manifest.DockerPush{Name: "t3"},
		},
	}
	config := testPipeline().Render(man)

	assert.Nil(t, config.Jobs[0].Plan[0].Passed)
	getGitStep := (config.Jobs[0].Plan[0].InParallel.Steps)[0]
	assert.Equal(t, gitName, getGitStep.Name())
	assert.False(t, getGitStep.Trigger)

	getGitStep = (config.Jobs[1].Plan[0].InParallel.Steps)[0]
	assert.Equal(t, config.Jobs[0].Name, getGitStep.Passed[0])
	assert.False(t, getGitStep.Trigger)

	getGitStep = (config.Jobs[2].Plan[0].InParallel.Steps)[0]
	assert.Equal(t, config.Jobs[1].Name, getGitStep.Passed[0])
	assert.False(t, getGitStep.Trigger)
}

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

func TestRenderDeployMLTasksAsRunTask(t *testing.T) {
	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DeployMLZip{Name: "foobar 1"},
			manifest.DeployMLModules{Name: "foobar 2"},
		},
	}
	config := testPipeline().Render(man)
	assert.Equal(t, "foobar 1", config.Jobs[0].Plan[2].Task)
	assert.Equal(t, "foobar 2", config.Jobs[1].Plan[1].Task)
}
