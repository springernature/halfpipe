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

	//1. inParallel containing get "git"
	inParallel := *updateJob.Plan[0].InParallel
	assert.Equal(t, 1, len(inParallel.Steps))
	assert.Equal(t, gitName, (inParallel.Steps)[0].Name())

	//2. put "version"
	assert.Equal(t, updateJob.Plan[1].Name(), versionName)

	//3. task "update pipeline"
	assert.Equal(t, updateJob.Plan[2].Name(), updatePipelineName)

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
	updatePipeline := updateJob.Plan[2]

	assert.Equal(t, updatePipeline, pipeline.updatePipelineTask(man))
}

func TestUpdateThePipelineNameIsBasedOnBranch(t *testing.T) {
	man := manifest.Manifest{
		Pipeline: "some-pipeline",
		FeatureToggles: manifest.FeatureToggles{
			manifest.FeatureUpdatePipeline,
		},
	}

	p := testPipeline()

	//master
	assert.Equal(t, man.PipelineName(), p.updatePipelineTask(man).TaskConfig.Params["PIPELINE_NAME"])

	//branch
	man.Repo.Branch = "some-branch"
	assert.Equal(t, man.PipelineName(), p.updatePipelineTask(man).TaskConfig.Params["PIPELINE_NAME"])
}
