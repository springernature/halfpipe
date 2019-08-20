package linters

import (
	"strings"
	"testing"

	"fmt"

	"github.com/springernature/halfpipe/linters/errors"
	"github.com/stretchr/testify/assert"
)

func assertMissingField(t *testing.T, name string, err error) {
	t.Helper()

	mf, ok := err.(errors.MissingFieldError)
	if !ok {
		assert.Fail(t, "error is not a MissingField", err)
	} else {
		assert.Equal(t, name, mf.Name)
	}
}

func assertInvalidField(t *testing.T, name string, err error) {
	t.Helper()

	mf, ok := err.(errors.InvalidFieldError)
	if !ok {
		assert.Fail(t, "error is not an InvalidField", err)
	} else {
		assert.Equal(t, name, mf.Name)
	}
}

func assertInvalidFieldInErrors(t *testing.T, name string, errs []error) {
	t.Helper()

	for _, err := range errs {
		mf, ok := err.(errors.InvalidFieldError)
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
		mf, ok := err.(errors.TriggerError)
		if ok {
			if mf.TriggerName == name {
				return
			}
		}
	}
	assert.Fail(t, fmt.Sprintf("Could not find TriggerError error for '%s' in %s", name, errs))
}

func assertFileErrorInErrors(t *testing.T, glob string, errs []error) {
	t.Helper()

	for _, err := range errs {
		mf, ok := err.(errors.FileError)
		if ok {
			if strings.Contains(mf.Path, glob) {
				return
			}
		}
	}
	assert.Fail(t, fmt.Sprintf("Could not find invalid field error for '%s' in %s", glob, errs))
}

func assertTooManyAppsError(t *testing.T, name string, err error) {
	t.Helper()

	mf, ok := err.(errors.TooManyAppsError)
	if !ok {
		assert.Fail(t, "error is not an TooManyAppsError", err)
	} else {
		assert.Equal(t, name, mf.Path)
	}
}

func assertNoRoutesError(t *testing.T, name string, err error) {
	t.Helper()

	mf, ok := err.(errors.NoRoutesError)
	if !ok {
		assert.Fail(t, "error is not an TooManyAppsError", err)
	} else {
		assert.Equal(t, name, mf.Path)
	}
}

func assertNoNameError(t *testing.T, name string, err error) {
	t.Helper()

	mf, ok := err.(errors.NoNameError)
	if !ok {
		assert.Fail(t, "error is not an TooManyAppsError", err)
	} else {
		assert.Equal(t, name, mf.Path)
	}
}

func assertWrongHealthCheck(t *testing.T, name string, err error) {
	t.Helper()

	mf, ok := err.(errors.WrongHealthCheck)
	if !ok {
		assert.Fail(t, "error is not an TooManyAppsError", err)
	} else {
		assert.Equal(t, name, mf.Path)
	}
}

func assertBadRoutes(t *testing.T, name string, err error) {
	t.Helper()

	mf, ok := err.(errors.BadRoutesError)
	if !ok {
		assert.Fail(t, "error is not an BadRoutesError", err)
	} else {
		assert.Equal(t, name, mf.Path)
	}
}
