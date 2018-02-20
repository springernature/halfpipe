package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLintResultErrorOutputWithAnchor(t *testing.T) {
	DocHost = "localhost"
	lintResult := NewLintResult("Test Linter", []error{
		NewInvalidField("fieldname.value_x", "reason")})
	assert.Contains(t, lintResult.Error(),
		"[see: https://localhost/docs/test-linter#invalid-field-fieldname.value_x]")
}

func TestLintResultErrorOutputWithoutAnchor(t *testing.T) {
	DocHost = "localhost"
	lintResult := NewLintResult("Vault Linter", []error{
		NewVaultClientError("error message")})
	assert.Contains(t, lintResult.Error(),
		"[see: https://localhost/docs/vault-linter#vault-client-error]")
}
