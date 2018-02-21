package model

import (
	"testing"

	"github.com/springernature/halfpipe/errors"
	"github.com/stretchr/testify/assert"
)

func TestLintResultErrorOutputWithAnchor(t *testing.T) {
	docHost = "localhost"
	lintResult := NewLintResult("Test Linter", []error{
		errors.NewInvalidField("fieldname.value_x", "reason")})
	assert.Contains(t, lintResult.Error(),
		"[see: https://localhost/docs/test-linter#invalid-field-fieldname.value_x ]")
}

func TestLintResultErrorOutputWithoutAnchor(t *testing.T) {
	docHost = "localhost"
	lintResult := NewLintResult("Vault Linter", []error{
		errors.NewVaultClientError("error message")})
	assert.Contains(t, lintResult.Error(),
		"[see: https://localhost/docs/vault-linter#vault-client-error ]")
}
