package linters

import (
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestHasErrors(t *testing.T) {

	lintResult := NewLintResult("blah", "url", []error{})
	lintResults := LintResults{
		lintResult,
	}

	assert.False(t, lintResults.HasErrors())

	lintResult.Add(errors.New("Blurg"))
	lintResults = LintResults{
		lintResult,
	}
	assert.True(t, lintResults.HasErrors())
}

func TestError(t *testing.T) {
	e1 := errors.New("error1")
	e2 := errors.New("error2")
	documentedError := NewErrMissingField("blurgh")
	hasErrors := NewLintResult("blah", "url", []error{e1, e2, documentedError})

	assert.Contains(t, hasErrors.Error(), e1.Error())
	assert.Contains(t, hasErrors.Error(), e2.Error())
}

func TestDeduplicatesErrors(t *testing.T) {
	err := errors.New("This should only appear once in the output")
	lintResult := NewLintResult("linter", "url", []error{err, err, err})

	numTimesErrInLintResultStr := strings.Count(lintResult.Error(), err.Error())
	assert.Equal(t, 1, numTimesErrInLintResultStr)
}

func TestDifferenceBetweenErrorsAndWarnings(t *testing.T) {
	err := newError("some error")
	warn := newError("some warning").AsWarning()

	lintResult := NewLintResult("linter", "url", nil)
	assert.False(t, lintResult.HasErrors())
	assert.False(t, lintResult.HasWarnings())

	lintResult.Add(warn)
	assert.False(t, lintResult.HasErrors())
	assert.True(t, lintResult.HasWarnings())

	lintResult.Add(err)
	assert.True(t, lintResult.HasErrors())
	assert.True(t, lintResult.HasWarnings())
}

func TestDeduplicateErrorsAndWarnings(t *testing.T) {
	error1 := newError("some error1")
	error2 := newError("some error2")
	warn1 := newError("some warning1").AsWarning()
	warn2 := newError("some warning2").AsWarning()
	lintResult := NewLintResult("linter", "url", []error{error1, error2, error1, warn1, warn2, warn1})

	assert.Equal(t, 1, strings.Count(lintResult.Error(), error1.Error()))
	assert.Equal(t, 1, strings.Count(lintResult.Error(), error2.Error()))
	assert.Equal(t, 1, strings.Count(lintResult.Error(), warn1.Error()))
	assert.Equal(t, 1, strings.Count(lintResult.Error(), warn2.Error()))

}
