package linters

import (
	"testing"

	"github.com/springernature/halfpipe/model"
	"github.com/stretchr/testify/assert"
)

var taskLinter = TaskLinter{}

func TestAtLeastOneTaskExists(t *testing.T) {
	man := model.Manifest{}

	errs := taskLinter.Lint(man)
	assert.Len(t, errs, 1)
	assert.IsType(t, model.MissingField{}, errs[0])
}
