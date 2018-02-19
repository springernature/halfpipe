package errors

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestLintResultErrorOutputWithAnchor(t *testing.T) {
	DocBaseUrl = "http://localhost"
	lintResult := NewLintResult("Test Linter", []error{
		NewInvalidField("fieldname", "reason")})
	assert.Contains(t, lintResult.Error(),
		"[see: http://localhost/docs/help_test_linter#fieldname]")
}

func TestLintResultErrorOutputWithoutAnchor(t *testing.T) {
	DocBaseUrl = "http://localhost"
	lintResult := NewLintResult("VaultLinter", []error{
		NewVaultClientError("error message")})
	assert.Contains(t, lintResult.Error(),
		"[see: http://localhost/docs/help_vault_linter]")
}
