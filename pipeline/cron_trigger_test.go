package pipeline

import (
	"testing"

	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestCronTriggerNotSet(t *testing.T) {
	man := manifest.Manifest{
		Repo: manifest.Repo{URI: gitDir},
		Tasks: []manifest.Task{
			manifest.Run{Script: "run.sh"},
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
	assert.Equal(t, gitDir, (*plan[0].Aggregate)[0].Name())
	assert.True(t, (*plan[0].Aggregate)[0].Trigger)
	assert.Equal(t, "run run.sh", plan[1].Task)
}

func TestCronTriggerSetAddsResource(t *testing.T) {
	man := manifest.Manifest{
		Repo:        manifest.Repo{URI: gitDir},
		CronTrigger: "*/10 * * * *",
		Tasks: []manifest.Task{
			manifest.Run{Script: "run.sh"},
		},
	}

	config := testPipeline().Render(man)
	resources := config.Resources
	assert.Equal(t, "git", resources[0].Type)
	assert.Equal(t, "cron */10 * * * *", resources[1].Name)
	assert.Equal(t, "cron-resource", resources[1].Type)
	assert.Equal(t, man.CronTrigger, resources[1].Source["expression"])
}

func TestCronTriggerSetWithCorrectPassedOnSecondJob(t *testing.T) {
	man := manifest.Manifest{
		Repo:        manifest.Repo{URI: gitDir},
		CronTrigger: "*/10 * * * *",
		Tasks: []manifest.Task{
			manifest.Run{Script: "s1.sh"},
			manifest.Run{Script: "s2.sh"},
		},
	}
	config := testPipeline().Render(man)

	t1 := config.Jobs[0].Plan
	t1Aggregate := *t1[0].Aggregate

	assert.Len(t, t1, 2)
	assert.Equal(t, gitDir, t1Aggregate[0].Name())
	assert.Equal(t, "cron */10 * * * *", t1Aggregate[1].Name())
	assert.True(t, t1Aggregate[1].Trigger)

	t2 := config.Jobs[1].Plan
	t2Aggregate := *t2[0].Aggregate
	assert.Len(t, t2, 2)
	assert.Equal(t, gitDir, t2Aggregate[0].Name())
	assert.Equal(t, []string{t1[1].Task}, t2Aggregate[0].Passed)

	assert.Equal(t, "cron */10 * * * *", t2Aggregate[1].Name())
	assert.Equal(t, []string{t1[1].Task}, t2Aggregate[1].Passed)
}

func TestCronTriggerSetWithParallelTasks(t *testing.T) {
	man := manifest.Manifest{
		Repo:        manifest.Repo{URI: gitDir},
		CronTrigger: "*/10 * * * *",
		Tasks: []manifest.Task{
			manifest.Run{Script: "first.sh"},
			manifest.Run{Script: "p1.sh", Parallel: true},
			manifest.Run{Script: "p2.sh", Parallel: true},
			manifest.Run{Script: "last.sh"},
		},
	}
	config := testPipeline().Render(man)

	first := config.Jobs[0].Plan
	firstAggregate := *first[0].Aggregate

	assert.Len(t, first, 2)
	assert.Equal(t, gitDir, firstAggregate[0].Name())
	assert.Equal(t, "cron */10 * * * *", firstAggregate[1].Name())
	assert.True(t, firstAggregate[1].Trigger)

	p1 := config.Jobs[1].Plan
	p1Aggregate := *p1[0].Aggregate
	assert.Len(t, p1, 2)
	assert.Equal(t, gitDir, p1Aggregate[0].Name())
	assert.Equal(t, []string{first[1].Task}, p1Aggregate[0].Passed)

	assert.Equal(t, "cron */10 * * * *", p1Aggregate[1].Name())
	assert.Equal(t, []string{first[1].Task}, p1Aggregate[1].Passed)

	p2 := config.Jobs[2].Plan
	p2Aggregate := *p2[0].Aggregate
	assert.Len(t, p2, 2)
	assert.Equal(t, gitDir, p2Aggregate[0].Name())
	assert.Equal(t, []string{first[1].Task}, p2Aggregate[0].Passed)

	assert.Equal(t, "cron */10 * * * *", p2Aggregate[1].Name())
	assert.Equal(t, []string{first[1].Task}, p2Aggregate[1].Passed)

	last := config.Jobs[3].Plan
	lastAggregate := *last[0].Aggregate
	assert.Len(t, last, 2)
	assert.Equal(t, gitDir, lastAggregate[0].Name())
	assert.Equal(t, []string{p1[1].Task, p2[1].Task}, lastAggregate[0].Passed)

	assert.Equal(t, "cron */10 * * * *", lastAggregate[1].Name())
	assert.Equal(t, []string{p1[1].Task, p2[1].Task}, lastAggregate[1].Passed)
}

func TestCronTriggerSetWhenUsingRestoreArtifact(t *testing.T) {
	man := manifest.Manifest{
		Repo:        manifest.Repo{URI: gitDir},
		CronTrigger: "*/10 * * * *",
		Tasks: []manifest.Task{
			manifest.Run{Script: "first.sh", SaveArtifacts: []string{"something"}},
			manifest.Run{Script: "p1.sh", Parallel: true},
			manifest.Run{Script: "p2.sh", Parallel: true, RestoreArtifacts: true},
			manifest.Run{Script: "last.sh", RestoreArtifacts: true},
		},
	}

	config := testPipeline().Render(man)

	first := config.Jobs[0].Plan
	firstAggregate := *first[0].Aggregate

	assert.Len(t, first, 3)
	assert.Equal(t, gitDir, firstAggregate[0].Name())
	assert.Equal(t, "cron */10 * * * *", firstAggregate[1].Name())
	assert.True(t, firstAggregate[1].Trigger)

	p1 := config.Jobs[1].Plan
	p1Aggregate := *p1[0].Aggregate
	assert.Len(t, p1, 2)
	assert.Equal(t, gitDir, p1Aggregate[0].Name())
	assert.Equal(t, []string{first[1].Task}, p1Aggregate[0].Passed)

	assert.Equal(t, "cron */10 * * * *", p1Aggregate[1].Name())
	assert.Equal(t, []string{first[1].Task}, p1Aggregate[1].Passed)

	p2 := config.Jobs[2].Plan
	p2Aggregate := *p2[0].Aggregate
	assert.Len(t, p2, 2)
	assert.Equal(t, gitDir, p2Aggregate[0].Name())
	assert.Equal(t, []string{first[1].Task}, p2Aggregate[0].Passed)

	assert.Equal(t, "cron */10 * * * *", p2Aggregate[1].Name())
	assert.Equal(t, []string{first[1].Task}, p2Aggregate[1].Passed)

	// Artifacts should not have any passed.
	assert.Contains(t, "artifacts", p2Aggregate[2].Name())
	assert.Len(t, p2Aggregate[2].Passed, 0)

	last := config.Jobs[3].Plan
	lastAggregate := *last[0].Aggregate
	assert.Len(t, last, 2)
	assert.Equal(t, gitDir, lastAggregate[0].Name())
	assert.Equal(t, []string{p1[1].Task, p2[1].Task}, lastAggregate[0].Passed)

	assert.Equal(t, "cron */10 * * * *", lastAggregate[1].Name())
	assert.Equal(t, []string{p1[1].Task, p2[1].Task}, lastAggregate[1].Passed)

	// Artifacts should not have any passed.
	assert.Contains(t, "artifacts", lastAggregate[2].Name())
	assert.Len(t, lastAggregate[2].Passed, 0)
}
