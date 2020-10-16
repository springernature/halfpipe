package pipeline

//
//import (
//	"github.com/springernature/halfpipe/manifest"
//	"github.com/stretchr/testify/assert"
//	"testing"
//)
//
//func TestShouldAddUpdateJobAsFirstJob(t *testing.T) {
//	man := manifest.Manifest{
//		FeatureToggles: manifest.FeatureToggles{
//			manifest.FeatureUpdatePipeline,
//		},
//		Tasks: manifest.TaskList{
//			manifest.Update{},
//		},
//	}
//
//	cfg := testPipeline().Render(man)
//
//	_, found := cfg.Jobs.Lookup(manifest.Update{}.GetName())
//	assert.True(t, found)
//	assert.Equal(t, manifest.Update{}.GetName(), cfg.Jobs[0].Name)
//}
//
//func TestShouldAddUpdatePipelineTask(t *testing.T) {
//	man := manifest.Manifest{
//		Triggers: manifest.TriggerList{
//			manifest.GitTrigger{},
//		},
//		FeatureToggles: manifest.FeatureToggles{
//			manifest.FeatureUpdatePipeline,
//		},
//		Tasks: manifest.TaskList{
//			manifest.Update{},
//		},
//	}
//
//	cfg := testPipeline().Render(man)
//
//	updateJob, _ := cfg.Jobs.Lookup(manifest.Update{}.GetName())
//
//	//should be 3 things in plan
//	assert.Equal(t, 3, len(updateJob.Plan))
//
//	//1. inParallel containing get "git"
//	inParallel := *updateJob.Plan[0].InParallel
//	assert.Equal(t, 1, len(inParallel.Steps))
//	assert.Equal(t, manifest.GitTrigger{}.GetTriggerName(), (inParallel.Steps)[0].Name())
//	assert.True(t, (inParallel.Steps)[0].Trigger)
//
//	//2. task "update pipeline"
//	assert.Equal(t, updateJob.Plan[1].Name(), "update")
//
//	//3. put "version"
//	assert.Equal(t, updateJob.Plan[2].Name(), versionName)
//
//}
//
//func TestShouldAddUpdatePipelineTaskWithManualGitTrigger(t *testing.T) {
//	man := manifest.Manifest{
//		Triggers: manifest.TriggerList{
//			manifest.GitTrigger{ManualTrigger: true},
//		},
//		FeatureToggles: manifest.FeatureToggles{
//			manifest.FeatureUpdatePipeline,
//		},
//		Tasks: manifest.TaskList{
//			manifest.Update{},
//		},
//	}
//
//	cfg := testPipeline().Render(man)
//
//	updateJob, _ := cfg.Jobs.Lookup(manifest.Update{}.GetName())
//
//	//1. inParallel containing get "git"
//	inParallel := *updateJob.Plan[0].InParallel
//	assert.Equal(t, 1, len(inParallel.Steps))
//	assert.Equal(t, manifest.GitTrigger{}.GetTriggerName(), (inParallel.Steps)[0].Name())
//	assert.False(t, (inParallel.Steps)[0].Trigger)
//}
//
//func TestUpdatePipelinePlan(t *testing.T) {
//	man := manifest.Manifest{
//		Pipeline: "some-name",
//		FeatureToggles: manifest.FeatureToggles{
//			manifest.FeatureUpdatePipeline,
//		},
//		Tasks: manifest.TaskList{
//			manifest.Update{},
//		},
//	}
//
//	pipeline := testPipeline()
//	cfg := pipeline.Render(man)
//	updateJob, _ := cfg.Jobs.Lookup(manifest.Update{}.GetName())
//	updatePipeline := updateJob.Plan[1]
//
//	assert.Equal(t, updatePipeline, pipeline.updatePipelineTask(man.PipelineName(), man.Triggers.GetGitTrigger().BasePath))
//}
//
//func TestUpdateThePipelineNameIsBasedOnBranch(t *testing.T) {
//	man := manifest.Manifest{
//		Pipeline: "some-pipeline",
//		FeatureToggles: manifest.FeatureToggles{
//			manifest.FeatureUpdatePipeline,
//		},
//	}
//
//	p := testPipeline()
//
//	//master
//	assert.Equal(t, man.PipelineName(), p.updatePipelineTask(man.PipelineName(), man.Triggers.GetGitTrigger().BasePath).TaskConfig.Params["PIPELINE_NAME"])
//
//	//branch
//	man = manifest.Manifest{
//		Pipeline: "some-pipeline",
//		Triggers: manifest.TriggerList{
//			manifest.GitTrigger{
//				Branch: "some-branch",
//			},
//		},
//		FeatureToggles: manifest.FeatureToggles{
//			manifest.FeatureUpdatePipeline,
//		},
//	}
//
//	assert.Equal(t, man.PipelineName(), p.updatePipelineTask(man.PipelineName(), man.Triggers.GetGitTrigger().BasePath).TaskConfig.Params["PIPELINE_NAME"])
//}
