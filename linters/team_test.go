package linters

import (
	"testing"

	"github.com/springernature/halfpipe/model"
	"github.com/stretchr/testify/assert"
)

var teamLinter = TeamLinter{}

func TestTeamIsEmpty(t *testing.T) {
	man := model.Manifest{}
	errs := teamLinter.Lint(man)
	assert.Len(t, errs, 1)
	assertMissingField(t, "team", errs[0])
}

func TestTeamIsValid(t *testing.T) {
	man := model.Manifest{
		Team: "yolo",
	}

	errs := teamLinter.Lint(man)
	assert.Empty(t, errs)
}
