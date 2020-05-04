package result

import (
	"strings"
	"testing"

	"github.com/pkg/errors"
	ourErrors "github.com/springernature/halfpipe/linters/linterrors"
	"github.com/stretchr/testify/assert"
)

func TestHasErrors(t *testing.T) {

	lintResult := NewLintResult("blah", "url", []error{}, nil)
	lintResults := LintResults{
		lintResult,
	}

	assert.False(t, lintResults.HasErrors())

	lintResult.AddError(errors.New("Blurg"))
	lintResults = LintResults{
		lintResult,
	}
	assert.True(t, lintResults.HasErrors())
}

func TestError(t *testing.T) {
	e1 := errors.New("error1")
	e2 := errors.New("error2")
	documentedError := ourErrors.NewMissingField("blurgh")
	hasErrors := NewLintResult("blah", "url", []error{e1, e2, documentedError}, nil)

	assert.Contains(t, hasErrors.Error(), e1.Error())
	assert.Contains(t, hasErrors.Error(), e2.Error())
}

func TestDeduplicatesErrors(t *testing.T) {
	err := errors.New("This should only appear once in the output")
	lintResult := NewLintResult(
		"linter",
		"url",
		[]error{
			err,
			err,
			err,
		},
		nil,
	)

	numTimesErrInLintResultStr := strings.Count(lintResult.Error(), err.Error())
	assert.Equal(t, 1, numTimesErrInLintResultStr)
}

func TestDifferenceBetweenErrorsAndWarnings(t *testing.T) {
	err := errors.New("some error")
	warn := errors.New("some warning")

	lintResult := NewLintResult("linter", "url", nil, nil)
	assert.False(t, lintResult.HasErrors())
	assert.False(t, lintResult.HasWarnings())

	lintResult.AddWarning(warn)
	assert.False(t, lintResult.HasErrors())
	assert.True(t, lintResult.HasWarnings())

	lintResult.AddError(err)
	assert.True(t, lintResult.HasErrors())
	assert.True(t, lintResult.HasWarnings())
}

func TestDeduplicateErrorsAndWarnings(t *testing.T) {
	err1 := errors.New("some error1")
	err2 := errors.New("some error2")
	warn1 := errors.New("some warning1")
	warn2 := errors.New("some warning2")
	lintResult := NewLintResult("linter", "url", []error{err1, err2, err1}, []error{warn1, warn2, warn1})

	assert.Equal(t, 1, strings.Count(lintResult.Error(), err1.Error()))
	assert.Equal(t, 1, strings.Count(lintResult.Error(), err2.Error()))
	assert.Equal(t, 1, strings.Count(lintResult.Error(), warn1.Error()))
	assert.Equal(t, 1, strings.Count(lintResult.Error(), warn2.Error()))

}
