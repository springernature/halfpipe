package pipeline

//
//import (
//	"testing"
//
//	"github.com/springernature/halfpipe/manifest"
//	"github.com/stretchr/testify/assert"
//)
//
//func TestPipelineTriggerResourceTypeSet(t *testing.T) {
//	man := manifest.Manifest{
//		Triggers: manifest.TriggerList{
//			manifest.PipelineTrigger{},
//		},
//		Tasks: []manifest.Task{
//			manifest.Run{Script: "run.sh"},
//		},
//	}
//
//	config := testPipeline().Render(man)
//	_, found := config.ResourceTypes.Lookup("halfpipe-pipeline-trigger")
//	assert.True(t, found)
//}
//
//func TestPipelineTriggerSetAddsResource(t *testing.T) {
//	trigger := manifest.PipelineTrigger{
//		Pipeline: "a",
//		Job:      "b",
//	}
//	man := manifest.Manifest{
//		Triggers: manifest.TriggerList{
//			trigger,
//		},
//		Tasks: []manifest.Task{
//			manifest.Run{Script: "run.sh"},
//		},
//	}
//
//	config := testPipeline().Render(man)
//	resource, found := config.Resources.Lookup(trigger.GetTriggerName())
//	assert.True(t, found)
//	assert.Equal(t, "halfpipe-pipeline-trigger", resource.Type)
//	assert.Equal(t, trigger.Pipeline, resource.Source["pipeline"])
//	assert.Equal(t, trigger.Job, resource.Source["job"])
//}
//
//func TestPipelineTriggerSetWithCorrectPassedOnSecondJob(t *testing.T) {
//	trigger1 := manifest.PipelineTrigger{
//		Pipeline: "a",
//		Job:      "b",
//	}
//	trigger2 := manifest.PipelineTrigger{
//		Pipeline: "aa",
//		Job:      "bb",
//	}
//
//	man := manifest.Manifest{
//		Triggers: manifest.TriggerList{
//			trigger1,
//			trigger2,
//		},
//		Tasks: []manifest.Task{
//			manifest.Run{Script: "s1.sh"},
//			manifest.Run{Script: "s2.sh"},
//		},
//	}
//	config := testPipeline().Render(man)
//
//	t1 := config.Jobs[0]
//	t1InParallel := t1.Plan[0].InParallel.Steps
//
//	assert.Len(t, t1.Plan, 2)
//	assert.Equal(t, trigger1.GetTriggerName(), t1InParallel[0].Name())
//	assert.Equal(t, trigger2.GetTriggerName(), t1InParallel[1].Name())
//	assert.True(t, t1InParallel[0].Trigger)
//
//	t2 := config.Jobs[1].Plan
//	t2InParallel := t2[0].InParallel.Steps
//	assert.Len(t, t2, 2)
//
//	assert.Equal(t, trigger1.GetTriggerName(), t2InParallel[0].Name())
//	assert.Equal(t, trigger2.GetTriggerName(), t2InParallel[1].Name())
//	assert.Equal(t, trigger2.GetTriggerAttempts(), t2InParallel[0].Attempts)
//	assert.Equal(t, []string{t1.Name}, t2InParallel[0].Passed)
//}
//
//func TestPipelineTriggerSetWithParallelTasks(t *testing.T) {
//	trigger := manifest.PipelineTrigger{
//		Pipeline: "a",
//		Job:      "b",
//	}
//
//	man := manifest.Manifest{
//		Triggers: manifest.TriggerList{
//			trigger,
//		},
//		Tasks: []manifest.Task{
//			manifest.Run{Script: "first.sh"},
//			manifest.Parallel{
//				Tasks: manifest.TaskList{
//					manifest.Run{Script: "p1.sh"},
//					manifest.Run{Script: "p2.sh"},
//				},
//			},
//			manifest.Run{Script: "last.sh"},
//		},
//	}
//	config := testPipeline().Render(man)
//
//	first := config.Jobs[0]
//	firstInParallel := first.Plan[0].InParallel.Steps
//
//	assert.Len(t, first.Plan, 2)
//	assert.Equal(t, trigger.GetTriggerName(), firstInParallel[0].Name())
//	assert.True(t, firstInParallel[0].Trigger)
//
//	p1 := config.Jobs[1]
//	p1InParallel := p1.Plan[0].InParallel.Steps
//	assert.Len(t, p1.Plan, 2)
//
//	assert.Equal(t, trigger.GetTriggerName(), p1InParallel[0].Name())
//	assert.Equal(t, []string{first.Name}, p1InParallel[0].Passed)
//
//	p2 := config.Jobs[2]
//	p2InParallel := p2.Plan[0].InParallel.Steps
//	assert.Len(t, p2.Plan, 2)
//
//	assert.Equal(t, trigger.GetTriggerName(), p2InParallel[0].Name())
//	assert.Equal(t, []string{first.Name}, p2InParallel[0].Passed)
//
//	last := config.Jobs[3].Plan
//	lastInParallel := last[0].InParallel.Steps
//	assert.Len(t, last, 2)
//
//	assert.Equal(t, trigger.GetTriggerName(), lastInParallel[0].Name())
//	assert.Equal(t, []string{p1.Name, p2.Name}, lastInParallel[0].Passed)
//}
