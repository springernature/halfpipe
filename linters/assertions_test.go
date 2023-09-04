package linters

import (
	"errors"
	"golang.org/x/exp/slices"
	"testing"
)

func assertContainsError(t *testing.T, errs []error, expected error) {
	t.Helper()
	if !containsError(errs, expected) {
		t.Fatalf("%s does not contain '%s'", errs, expected)
	}
}

func assertNotContainsError(t *testing.T, errs []error, expected error) {
	t.Helper()
	if containsError(errs, expected) {
		t.Fatalf("%s should not contain '%s'", errs, expected)
	}
}

func containsError(errs []error, expected error) bool {
	return slices.ContainsFunc(errs, func(e error) bool { return errors.Is(e, expected) })
}
