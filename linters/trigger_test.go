package linters

import (
	"testing"

	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func testTriggerLinter() triggerLinter {
	return triggerLinter{}
}

func TestCronTrigger(t *testing.T) {
	man := manifest.Manifest{}
	man.Team = "yolo"
	man.Pipeline = "alles-gut"
	man.CronTrigger = "*/10 * * * *"

	assert.False(t, testTriggerLinter().Lint(man).HasErrors())
}

func TestInvalidCronTrigger(t *testing.T) {
	man := manifest.Manifest{}
	man.Team = "yolo"
	man.Pipeline = "alles-gut"
	man.CronTrigger = "*/99 * * * *"

	assert.True(t, testTriggerLinter().Lint(man).HasErrors())
}

func TestCronTriggerWithSecondsShouldHaveError(t *testing.T) {
	man := manifest.Manifest{}
	man.Team = "yolo"
	man.Pipeline = "alles-gut"
	// 6 parts means there is seconds.
	man.CronTrigger = "* * * * * *"

	assert.True(t, testTriggerLinter().Lint(man).HasErrors())
}
