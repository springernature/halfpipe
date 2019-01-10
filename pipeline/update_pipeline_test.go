package pipeline

import (
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestShouldAddUpdateJobAsFirstJob(t *testing.T) {
	man := manifest.Manifest{
		FeatureToggles: manifest.FeatureToggles{
			manifest.FeatureUpdatePipeline,
		},
	}

	cfg := testPipeline().Render(man)

	_, found := cfg.Jobs.Lookup(updateJobName)
	assert.True(t, found)
	assert.Equal(t, updateJobName, cfg.Jobs[0].Name)
}

func TestShouldAddUpdatePipelineTask(t *testing.T) {
	man := manifest.Manifest{
		FeatureToggles: manifest.FeatureToggles{
			manifest.FeatureUpdatePipeline,
		},
	}

	cfg := testPipeline().Render(man)

	updateJob, _ := cfg.Jobs.Lookup(updateJobName)

	//should be 3 things in plan
	assert.Equal(t, 3, len(updateJob.Plan))

	//1. aggregate containing get "git"
	aggregate := updateJob.Plan[0].Aggregate
	assert.Equal(t, 1, len(*aggregate))
	assert.Equal(t, gitName, (*aggregate)[0].Name())

	//2. task "update pipeline"
	assert.Equal(t, updateJob.Plan[1].Name(), updatePipelineName)

	//3. put "version"
	assert.Equal(t, updateJob.Plan[2].Name(), versionName)
}

func TestUpdatePipelinePlan(t *testing.T) {
	man := manifest.Manifest{
		Pipeline: "some-name",
		FeatureToggles: manifest.FeatureToggles{
			manifest.FeatureUpdatePipeline,
		},
	}

	pipeline := testPipeline()
	cfg := pipeline.Render(man)
	updateJob, _ := cfg.Jobs.Lookup(updateJobName)
	updatePipeline := updateJob.Plan[1]

	assert.Equal(t, updatePipeline, pipeline.updatePipelineTask(man))
}