package linters

import (
	"strings"
	"testing"

	"fmt"

	"github.com/springernature/halfpipe/linters/linterrors"
	"github.com/stretchr/testify/assert"
)

func assertMissingField(t *testing.T, name string, err error) {
	t.Helper()

	mf, ok := err.(linterrors.MissingFieldError)
	if !ok {
		assert.Fail(t, "error is not a MissingField", err)
	} else {
		assert.Equal(t, name, mf.Name)
	}
}

func assertInvalidField(t *testing.T, name string, err error) {
	t.Helper()

	mf, ok := err.(linterrors.InvalidFieldError)
	if !ok {
		assert.Fail(t, "error is not an InvalidField", err)
	} else {
		assert.Equal(t, name, mf.Name)
	}
}

func assertInvalidFieldInErrors(t *testing.T, name string, errs []error) {
	t.Helper()

	for _, err := range errs {
		mf, ok := err.(linterrors.InvalidFieldError)
		if ok {
			if strings.Contains(mf.Name, name) {
				return
			}
		}
	}
	assert.Fail(t, fmt.Sprintf("Could not find invalid field error for '%s' in %s", name, errs))
}

func assertTriggerErrorInErrors(t *testing.T, name string, errs []error) {
	t.Helper()

	for _, err := range errs {
		mf, ok := err.(linterrors.TriggerError)
		if ok {
			if mf.TriggerName == name {
				return
			}
		}
	}
	assert.Fail(t, fmt.Sprintf("Could not find TriggerError error for '%s' in %s", name, errs))
}
