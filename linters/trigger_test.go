package linters

import (
	"testing"

	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func testTriggerLinter() triggerLinter {
	return triggerLinter{}
}

func TestCronTriggerWithIntervalTrigger(t *testing.T) {
	man := manifest.Manifest{}
	man.Team = "yolo"
	man.Pipeline = "alles-gut"
	man.CronTrigger = "*/10 * * * *"
	man.TriggerInterval = "10m"

	result := testTriggerLinter().Lint(man)
	assert.True(t, result.HasErrors())
}

func TestCronTriggerOnly(t *testing.T) {
	man := manifest.Manifest{}
	man.Team = "yolo"
	man.Pipeline = "alles-gut"
	man.CronTrigger = "*/10 * * * *"

	result := testTriggerLinter().Lint(man)
	assert.False(t, result.HasErrors())
}

func TestInvalidCronTrigger(t *testing.T) {
	man := manifest.Manifest{}
	man.Team = "yolo"
	man.Pipeline = "alles-gut"
	man.CronTrigger = "*/99 * * * *"

	result := testTriggerLinter().Lint(man)
	assert.True(t, result.HasErrors())
}

func TestIntervalTriggerTriggerOnly(t *testing.T) {
	man := manifest.Manifest{}
	man.Team = "yolo"
	man.Pipeline = "alles-gut"
	man.TriggerInterval = "10m"

	result := testTriggerLinter().Lint(man)
	assert.True(t, result.HasWarnings())
}

