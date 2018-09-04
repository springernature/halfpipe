package linters

import (
	"testing"

	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func testTeamLinter() teamLinter {
	return teamLinter{}
}

func TestAllMissing(t *testing.T) {
	man := manifest.Manifest{}
	result := testTeamLinter().Lint(man)
	assert.Len(t, result.Errors, 2)
}

func TestTeamIsMissing(t *testing.T) {
	man := manifest.Manifest{}
	man.Pipeline = "yolo"

	result := testTeamLinter().Lint(man)
	assert.Len(t, result.Errors, 1)
	assertMissingField(t, "team", result.Errors[0])
}

func TestTeamIsUpperCase(t *testing.T) {
	man := manifest.Manifest{}
	man.Pipeline = "yolo"
	man.Team = "yoLo"

	result := testTeamLinter().Lint(man)
	assert.Len(t, result.Warnings, 1)
	assert.Len(t, result.Errors, 0)
	assertInvalidField(t, "team", result.Warnings[0])
}

func TestPipelineIsMissing(t *testing.T) {
	man := manifest.Manifest{}
	man.Team = "yolo"

	result := testTeamLinter().Lint(man)
	assert.Len(t, result.Errors, 1)
	assertMissingField(t, "pipeline", result.Errors[0])
}

func TestPipelineIsValid(t *testing.T) {
	man := manifest.Manifest{}
	man.Team = "yolo"
	man.Pipeline = "Something with spaces"

	result := testTeamLinter().Lint(man)
	assert.True(t, result.HasErrors())
}

func TestHappyPath(t *testing.T) {
	man := manifest.Manifest{}
	man.Team = "yolo"
	man.Pipeline = "alles-gut"

	result := testTeamLinter().Lint(man)
	assert.False(t, result.HasErrors())
}
