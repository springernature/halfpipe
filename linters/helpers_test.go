package linters

import (
	"testing"

	"github.com/springernature/halfpipe/errors"
	"github.com/stretchr/testify/assert"
)

func assertMissingField(t *testing.T, name string, err error) {
	t.Helper()

	mf, ok := err.(errors.MissingField)
	if !ok {
		assert.Fail(t, "error is not a MissingField", err)
	} else {
		assert.Equal(t, name, mf.Name)
	}
}

func assertInvalidField(t *testing.T, name string, err error) {
	t.Helper()

	mf, ok := err.(errors.InvalidField)
	if !ok {
		assert.Fail(t, "error is not an InvalidField", err)
	} else {
		assert.Equal(t, name, mf.Name)
	}
}

func assertFileError(t *testing.T, path string, err error) {
	t.Helper()

	mf, ok := err.(errors.FileError)
	if !ok {
		assert.Fail(t, "error is not a FileError", err)
	} else {
		assert.Equal(t, path, mf.Path)
	}
}
