package model

import (
	"testing"

	"github.com/springernature/halfpipe/errors"
	"github.com/stretchr/testify/assert"
)

func TestLintResultErrorOutputWithAnchor(t *testing.T) {
	docHost = "localhost"
	lintResult := NewLintResult("Test Linter", []error{
		errors.NewMissingField("repo.uri"),
	})

	assert.Contains(t, lintResult.Error(),
		"[see: localhost/docs/linter-errors/#missing-field-repouri ]")
}

func TestLintResultErrorDoesntContainDocLink(t *testing.T) {
	docHost = "localhost"
	lintResult := NewLintResult("Test Linter", []error{
		errors.NewFileError("some/path", "not found"),
	})

	assert.NotContains(t, lintResult.Error(), "[see: localhost/")
}
