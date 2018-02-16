package linters

import (
	"testing"

	"github.com/springernature/halfpipe/model"
	"github.com/stretchr/testify/assert"
)

func testTeamLinter() TeamLinter {
	return TeamLinter{}
}

func TestTeamIsEmpty(t *testing.T) {
	man := model.Manifest{}
	result := testTeamLinter().Lint(man)
	assert.Len(t, result.Errors, 1)
	assertMissingField(t, "team", result.Errors[0])
}

func TestTeamIsValid(t *testing.T) {
	man := model.Manifest{
		Team: "yolo",
	}

	result := testTeamLinter().Lint(man)
	assert.False(t, result.HasErrors())
}
