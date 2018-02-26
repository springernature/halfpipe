package model

import (
	"testing"

	"strings"

	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/errors"
	"github.com/stretchr/testify/assert"
)

func TestLintResultErrorOutputWithAnchor(t *testing.T) {
	config.DocHost = "localhost"
	lintResult := NewLintResult("Test Linter", []error{
		errors.NewMissingField("repo.uri"),
	})

	assert.Contains(t, lintResult.Error(),
		"[see: localhost/docs/linter-errors/#missing-field-repouri ]")
}

func TestLintResultErrorDoesntContainDocLink(t *testing.T) {
	config.DocHost = "localhost"
	lintResult := NewLintResult("Test Linter", []error{
		errors.NewFileError("some/path", "not found"),
	})

	assert.NotContains(t, lintResult.Error(), "[see: localhost/")
}

func TestLintResultErrorDoesntContainDuplicates(t *testing.T) {
	error1 := errors.NewFileError("some/path", "not found")
	error1DifferentInstance := errors.NewFileError("some/path", "not found")
	error2 := errors.NewFileError("some/other/path", "not found")
	error3 := errors.NewInvalidField("repo", "must not be empty")
	error4 := errors.NewInvalidField("repo.uri", "must not be empty")
	error5 := errors.NewVaultSecretNotFoundError("prefix", "team", "pipeline", "((a.b))")

	lintResult := NewLintResult("Test Linter", []error{
		error1,
		error5,
		error3,
		error2,
		error1DifferentInstance,
		error3,
		error4,
		error5,
		error1DifferentInstance,
		error4,
		error3,
		error5,
		error1DifferentInstance,
		error5,
		error1,
		error5,
	})

	assert.Equal(t, 1, strings.Count(lintResult.Error(), error1.Error()))
	assert.Equal(t, 1, strings.Count(lintResult.Error(), error2.Error()))
	assert.Equal(t, 1, strings.Count(lintResult.Error(), error3.Error()))
	assert.Equal(t, 1, strings.Count(lintResult.Error(), error4.Error()))
	assert.Equal(t, 1, strings.Count(lintResult.Error(), error5.Error()))
	assert.Equal(t, 5, strings.Count(lintResult.Error(), "*"))
}
