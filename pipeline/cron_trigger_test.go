package pipeline

import (
	"testing"

	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestCronTriggerResourceTypeSet(t *testing.T) {
	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.Cron{
				Trigger: "*/10 * * * *",
			},
		},
		Tasks: []manifest.Task{
			manifest.Run{Script: "run.sh"},
		},
	}

	config := testPipeline().Render(man)
	_, found := config.ResourceTypes.Lookup("cron-resource")
	assert.True(t, found)
}

func TestCronTriggerNotSet(t *testing.T) {
	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.Git{},
		},
		Tasks: []manifest.Task{
			manifest.Run{Name: "blah", Script: "run.sh"},
		},
	}
	config := testPipeline().Render(man)
	resources := config.Resources
	plan := config.Jobs[0].Plan

	//should be 1 resource: git
	assert.Len(t, resources, 1)
	assert.Equal(t, "git", resources[0].Type)

	//should be 2 items in the plan: get git + task
	assert.Len(t, plan, 2)
	assert.Equal(t, gitName, (plan[0].InParallel.Steps)[0].Name())
	assert.True(t, (plan[0].InParallel.Steps)[0].Trigger)
	assert.Equal(t, "blah", plan[1].Task)
}

func TestCronTriggerSetAddsResource(t *testing.T) {
	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.Cron{
				Trigger: "*/10 * * * *",
			},
		},
		Tasks: []manifest.Task{
			manifest.Run{Script: "run.sh"},
		},
	}

	config := testPipeline().Render(man)
	resource, found := config.Resources.Lookup(manifest.Cron{}.GetTriggerName())
	assert.True(t, found)
	assert.Equal(t, cronName, resource.Name)
	assert.Equal(t, "cron-resource", resource.Type)
	assert.Equal(t, man.Triggers[0].(manifest.Cron).Trigger, resource.Source["expression"])
	assert.Equal(t, "1m", resource.CheckEvery)
}

func TestCronTriggerSetWithCorrectPassedOnSecondJob(t *testing.T) {
	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.Cron{
				Trigger: "*/10 * * * *",
			},
		},
		Tasks: []manifest.Task{
			manifest.Run{Script: "s1.sh"},
			manifest.Run{Script: "s2.sh"},
		},
	}
	config := testPipeline().Render(man)

	t1 := config.Jobs[0].Plan
	t1InParallel := t1[0].InParallel.Steps

	assert.Len(t, t1, 2)
	assert.Equal(t, cronName, t1InParallel[0].Name())
	assert.True(t, t1InParallel[0].Trigger)

	t2 := config.Jobs[1].Plan
	t2InParallel := t2[0].InParallel.Steps
	assert.Len(t, t2, 2)

	assert.Equal(t, cronName, t2InParallel[0].Name())
	assert.Equal(t, cronGetAttempts, t2InParallel[0].Attempts)
	assert.Equal(t, []string{t1[1].Task}, t2InParallel[0].Passed)
}

func TestCronTriggerSetWithParallelTasks(t *testing.T) {
	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.Cron{
				Trigger: "*/10 * * * *",
			},
		},
		Tasks: []manifest.Task{
			manifest.Run{Script: "first.sh"},
			manifest.Parallel{
				Tasks: manifest.TaskList{
					manifest.Run{Script: "p1.sh"},
					manifest.Run{Script: "p2.sh"},
				},
			},
			manifest.Run{Script: "last.sh"},
		},
	}
	config := testPipeline().Render(man)

	first := config.Jobs[0].Plan
	firstInParallel := first[0].InParallel.Steps

	assert.Len(t, first, 2)
	assert.Equal(t, cronName, firstInParallel[0].Name())
	assert.True(t, firstInParallel[0].Trigger)

	p1 := config.Jobs[1].Plan
	p1InParallel := p1[0].InParallel.Steps
	assert.Len(t, p1, 2)

	assert.Equal(t, cronName, p1InParallel[0].Name())
	assert.Equal(t, []string{first[1].Task}, p1InParallel[0].Passed)

	p2 := config.Jobs[2].Plan
	p2InParallel := p2[0].InParallel.Steps
	assert.Len(t, p2, 2)

	assert.Equal(t, cronName, p2InParallel[0].Name())
	assert.Equal(t, []string{first[1].Task}, p2InParallel[0].Passed)

	last := config.Jobs[3].Plan
	lastInParallel := last[0].InParallel.Steps
	assert.Len(t, last, 2)

	assert.Equal(t, cronName, lastInParallel[0].Name())
	assert.Equal(t, []string{p1[1].Task, p2[1].Task}, lastInParallel[0].Passed)
}

func TestCronTriggerSetWhenUsingRestoreArtifact(t *testing.T) {
	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.Cron{
				Trigger: "*/10 * * * *",
			},
		},
		Tasks: []manifest.Task{
			manifest.Run{Script: "first.sh", SaveArtifacts: []string{"something"}},
			manifest.Parallel{
				Tasks: manifest.TaskList{
					manifest.Run{Script: "p1.sh"},
					manifest.Run{Script: "p2.sh", RestoreArtifacts: true},
				},
			},
			manifest.Run{Script: "last.sh", RestoreArtifacts: true},
		},
	}

	config := testPipeline().Render(man)

	first := config.Jobs[0].Plan
	firstInParallel := first[0].InParallel.Steps

	assert.Len(t, first, 3)
	assert.Equal(t, cronName, firstInParallel[0].Name())
	assert.True(t, firstInParallel[0].Trigger)

	p1 := config.Jobs[1].Plan
	p1InParallel := p1[0].InParallel.Steps
	assert.Len(t, p1, 2)

	assert.Equal(t, cronName, p1InParallel[0].Name())
	assert.Equal(t, []string{first[1].Task}, p1InParallel[0].Passed)

	p2 := config.Jobs[2].Plan
	p2InParallel := p2[0].InParallel.Steps
	assert.Len(t, p2, 3)

	assert.Equal(t, cronName, p2InParallel[0].Name())
	assert.Equal(t, []string{first[1].Task}, p2InParallel[0].Passed)

	assert.Equal(t, cronName, p2InParallel[0].Name())
	assert.Equal(t, []string{first[1].Task}, p2InParallel[0].Passed)
	assert.Equal(t, restoreArtifactTask(man), p2[1])

	last := config.Jobs[3].Plan
	lastInParallel := last[0].InParallel.Steps
	assert.Len(t, last, 3)

	assert.Equal(t, cronName, lastInParallel[0].Name())
	assert.Equal(t, []string{p1[1].Task, p2[2].Task}, lastInParallel[0].Passed)

	// Artifacts should not have any passed.
	assert.Equal(t, restoreArtifactTask(man), last[1])
}
