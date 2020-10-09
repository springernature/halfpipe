package concourse

import (
	"fmt"
	"testing"

	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestDockerTriggerSetAddsResource(t *testing.T) {
	trigger := manifest.DockerTrigger{
		Image: "myUser/ubuntu-with-somedeps",
	}
	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			trigger,
		},
		Tasks: []manifest.Task{
			manifest.Run{Script: "run.sh"},
		},
	}

	config := testPipeline().Render(man)
	resource, found := config.Resources.Lookup(trigger.GetTriggerName())
	assert.True(t, found)
	assert.Equal(t, trigger.GetTriggerName(), resource.Name)
	assert.Equal(t, "docker-image", resource.Type)
	assert.Equal(t, man.Triggers[0].(manifest.DockerTrigger).Image, resource.Source["repository"])
}

func TestDockerTriggerSetWithCorrectPassedOnSecondJob(t *testing.T) {
	trigger := manifest.DockerTrigger{
		Image: "myUser/ubuntu-with-somedeps",
	}
	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			trigger,
		},
		Tasks: []manifest.Task{
			manifest.Run{Script: "s1.sh"},
			manifest.Run{Script: "s2.sh"},
		},
	}
	config := testPipeline().Render(man)

	fmt.Println(ToString(config))

	t1 := config.Jobs[0].Plan
	t1InParallel := t1[0].InParallel.Steps

	assert.Len(t, t1, 2)
	assert.Equal(t, trigger.GetTriggerName(), t1InParallel[0].Name())
	assert.True(t, t1InParallel[0].Trigger)

	t2 := config.Jobs[1].Plan
	t2InParallel := t2[0].InParallel.Steps
	assert.Len(t, t2, 2)

	assert.Equal(t, trigger.GetTriggerName(), t2InParallel[0].Name())
	assert.Equal(t, []string{t1[1].Task}, t2InParallel[0].Passed)
}

//
//func TestCronTriggerSetWithParallelTasks(t *testing.T) {
//	man := manifest.Manifest{
//		Triggers: manifest.TriggerList{
//			manifest.TimerTrigger{
//				Cron: "*/10 * * * *",
//			},
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
//	first := config.Jobs[0].Plan
//	firstInParallel := first[0].InParallel.Steps
//
//	assert.Len(t, first, 2)
//	assert.Equal(t, cronName, firstInParallel[0].Name())
//	assert.True(t, firstInParallel[0].Cron)
//
//	p1 := config.Jobs[1].Plan
//	p1InParallel := p1[0].InParallel.Steps
//	assert.Len(t, p1, 2)
//
//	assert.Equal(t, cronName, p1InParallel[0].Name())
//	assert.Equal(t, []string{first[1].Task}, p1InParallel[0].Passed)
//
//	p2 := config.Jobs[2].Plan
//	p2InParallel := p2[0].InParallel.Steps
//	assert.Len(t, p2, 2)
//
//	assert.Equal(t, cronName, p2InParallel[0].Name())
//	assert.Equal(t, []string{first[1].Task}, p2InParallel[0].Passed)
//
//	last := config.Jobs[3].Plan
//	lastInParallel := last[0].InParallel.Steps
//	assert.Len(t, last, 2)
//
//	assert.Equal(t, cronName, lastInParallel[0].Name())
//	assert.Equal(t, []string{p1[1].Task, p2[1].Task}, lastInParallel[0].Passed)
//}
//
//func TestCronTriggerSetWhenUsingRestoreArtifact(t *testing.T) {
//	man := manifest.Manifest{
//		Triggers: manifest.TriggerList{
//			manifest.TimerTrigger{
//				Cron: "*/10 * * * *",
//			},
//		},
//		Tasks: []manifest.Task{
//			manifest.Run{Script: "first.sh", SaveArtifacts: []string{"something"}},
//			manifest.Parallel{
//				Tasks: manifest.TaskList{
//					manifest.Run{Script: "p1.sh"},
//					manifest.Run{Script: "p2.sh", RestoreArtifacts: true},
//				},
//			},
//			manifest.Run{Script: "last.sh", RestoreArtifacts: true},
//		},
//	}
//
//	config := testPipeline().Render(man)
//
//	first := config.Jobs[0].Plan
//	firstInParallel := first[0].InParallel.Steps
//
//	assert.Len(t, first, 3)
//	assert.Equal(t, cronName, firstInParallel[0].Name())
//	assert.True(t, firstInParallel[0].Cron)
//
//	p1 := config.Jobs[1].Plan
//	p1InParallel := p1[0].InParallel.Steps
//	assert.Len(t, p1, 2)
//
//	assert.Equal(t, cronName, p1InParallel[0].Name())
//	assert.Equal(t, []string{first[1].Task}, p1InParallel[0].Passed)
//
//	p2 := config.Jobs[2].Plan
//	p2InParallel := p2[0].InParallel.Steps
//	assert.Len(t, p2, 3)
//
//	assert.Equal(t, cronName, p2InParallel[0].Name())
//	assert.Equal(t, []string{first[1].Task}, p2InParallel[0].Passed)
//
//	assert.Equal(t, cronName, p2InParallel[0].Name())
//	assert.Equal(t, []string{first[1].Task}, p2InParallel[0].Passed)
//	assert.Equal(t, restoreArtifactTask(man), p2[1])
//
//	last := config.Jobs[3].Plan
//	lastInParallel := last[0].InParallel.Steps
//	assert.Len(t, last, 3)
//
//	assert.Equal(t, cronName, lastInParallel[0].Name())
//	assert.Equal(t, []string{p1[1].Task, p2[2].Task}, lastInParallel[0].Passed)
//
//	// Artifacts should not have any passed.
//	assert.Equal(t, restoreArtifactTask(man), last[1])
//}
