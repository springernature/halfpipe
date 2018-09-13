package helpers

import (
	"fmt"
	"github.com/springernature/halfpipe/linters/errors"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func AssertMissingField(t *testing.T, name string, err error) {
	t.Helper()

	mf, ok := err.(errors.MissingFieldError)
	if !ok {
		assert.Fail(t, "error is not a MissingField", err)
	} else {
		assert.Equal(t, name, mf.Name)
	}
}

func AssertMissingFieldInErrors(t *testing.T, name string, errs []error) {
	t.Helper()

	for _, err := range errs {
		mf, ok := err.(errors.MissingFieldError)
		if ok {
			if strings.Contains(mf.Name, name) {
				return
			}
		}
	}
	assert.Fail(t, fmt.Sprintf("Could not find invalid field error for '%s' in %s", name, errs))
}

func AssertInvalidField(t *testing.T, name string, err error) {
	t.Helper()

	mf, ok := err.(errors.InvalidFieldError)
	if !ok {
		assert.Fail(t, "error is not an InvalidField", err)
	} else {
		assert.Equal(t, name, mf.Name)
	}
}

func AssertInvalidFieldInErrors(t *testing.T, name string, errs []error) {
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

func AssertInvalidFieldShouldNotBeInErrors(t *testing.T, name string, errs []error) {
	t.Helper()

	for _, err := range errs {
		mf, ok := err.(errors.InvalidFieldError)
		if ok {
			if mf.Name == name {
				assert.Fail(t, fmt.Sprintf("Invalid field error for '%s' should not be in %s!", name, errs))
			}
		}
	}
}

func AssertFileError(t *testing.T, path string, err error) {
	t.Helper()

	mf, ok := err.(errors.FileError)
	if !ok {
		assert.Fail(t, "error is not a FileError", err)
	} else {
		assert.Equal(t, path, mf.Path)
	}
}

func AssertFileErrorInErrors(t *testing.T, path string, errs []error) {
	t.Helper()

	for _, err := range errs {
		e, ok := err.(errors.FileError)
		if ok {
			if e.Path == path {
				return
			}
		}
	}
	assert.Fail(t, fmt.Sprintf("Could not find FileError for path '%s' in %s", path, errs))

}

func AssertTooManyAppsError(t *testing.T, name string, err error) {
	t.Helper()

	mf, ok := err.(errors.TooManyAppsError)
	if !ok {
		assert.Fail(t, "error is not an TooManyAppsError", err)
	} else {
		assert.Equal(t, name, mf.Path)
	}
}

func AssertNoRoutesError(t *testing.T, name string, err error) {
	t.Helper()

	mf, ok := err.(errors.NoRoutesError)
	if !ok {
		assert.Fail(t, "error is not an TooManyAppsError", err)
	} else {
		assert.Equal(t, name, mf.Path)
	}
}

func AssertNoNameError(t *testing.T, name string, err error) {
	t.Helper()

	mf, ok := err.(errors.NoNameError)
	if !ok {
		assert.Fail(t, "error is not an TooManyAppsError", err)
	} else {
		assert.Equal(t, name, mf.Path)
	}
}

func AssertWrongHealthCheck(t *testing.T, name string, err error) {
	t.Helper()

	mf, ok := err.(errors.WrongHealthCheck)
	if !ok {
		assert.Fail(t, "error is not an TooManyAppsError", err)
	} else {
		assert.Equal(t, name, mf.Path)
	}
}

func AssertBadRoutes(t *testing.T, name string, err error) {
	t.Helper()

	mf, ok := err.(errors.BadRoutesError)
	if !ok {
		assert.Fail(t, "error is not an BadRoutesError", err)
	} else {
		assert.Equal(t, name, mf.Path)
	}
}
