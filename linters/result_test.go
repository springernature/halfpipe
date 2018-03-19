package linters

import (
	"strings"
	"testing"

	"github.com/pkg/errors"
	ourErrors "github.com/springernature/halfpipe/linters/errors"
	"github.com/stretchr/testify/assert"
)

func TestHasErrors(t *testing.T) {

	lintResult := NewLintResult("blah", []error{}, nil)
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
	noErrors := NewLintResult("blah", []error{}, nil)

	assert.Contains(t, noErrors.Error(), `No errors \o/`)

	e1 := errors.New("error1")
	e2 := errors.New("error2")
	documentedError := ourErrors.NewMissingField("blurgh")
	hasErrors := NewLintResult("blah", []error{e1, e2, documentedError}, nil)

	assert.Contains(t, hasErrors.Error(), e1.Error())
	assert.Contains(t, hasErrors.Error(), e2.Error())
	assert.Contains(t, hasErrors.Error(), documentedError.DocID())
}

func TestDeduplicatesErrors(t *testing.T) {
	err := errors.New("This should only appear once in the output")
	lintResult := NewLintResult(
		"linter",
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

	lintResult := NewLintResult("linter", nil, nil)
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
	err := errors.New("some error")
	lintResult := NewLintResult("linter", []error{err}, []error{err})
	numTimesErrInLintResultStr := strings.Count(lintResult.Error(), err.Error())
	assert.Equal(t, 1, numTimesErrInLintResultStr)

}
