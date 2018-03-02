package linters

import (
	"testing"

	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func testTeamLinter() teamlinter {
	return teamlinter{}
}

func TestTeamIsEmpty(t *testing.T) {
	man := manifest.Manifest{}
	result := testTeamLinter().Lint(man)
	assert.Len(t, result.Errors, 1)
	assertMissingField(t, "team", result.Errors[0])
}

func TestTeamIsValid(t *testing.T) {
	man := manifest.Manifest{
		Team: "yolo",
	}

	result := testTeamLinter().Lint(man)
	assert.False(t, result.HasErrors())
}
