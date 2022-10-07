package linters

import (
	"errors"
	"testing"
)

func AssertContainsError(t *testing.T, errs []error, expected error) {
	t.Helper()
	if !containsError(errs, expected) {
		t.Fatalf("%s does not contain '%s'", errs, expected)
	}
}

func AssertNotContainsError(t *testing.T, errs []error, expected error) {
	t.Helper()
	if containsError(errs, expected) {
		t.Fatalf("%s should not contain '%s'", errs, expected)
	}
}

func containsError(errs []error, expected error) bool {
	for _, e := range errs {
		if errors.Is(e, expected) {
			return true
		}
	}
	return false
}
