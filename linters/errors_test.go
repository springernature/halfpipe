package linters

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_ErrorWrapping(t *testing.T) {
	baseErr := newError("bananas")
	err := baseErr.WithValue("are").WithValue("tasty").WithFile("fruit.txt")

	assert.ErrorIs(t, err, baseErr)
	assert.ErrorIs(t, err, baseErr.WithValue("are"))
	assert.ErrorIs(t, err, baseErr.WithValue("are").WithValue("tasty"))
	assert.ErrorIs(t, err, baseErr.WithValue("are").WithValue("tasty").WithFile("fruit.txt"))

	assert.NotErrorIs(t, err, baseErr.WithValue("tasty"))
	assert.NotErrorIs(t, err, baseErr.WithFile("fruit.txt"))
}

func Test_ErrorString(t *testing.T) {
	baseErr := newError("invalid field")
	err := baseErr.WithValue("script").WithValue("not found").WithFile("build.sh")

	assert.EqualError(t, baseErr, "invalid field")
	assert.EqualError(t, err, "invalid field: script: not found (build.sh)")
}

func Test_ErrorIsWarning(t *testing.T) {
	baseErr := newError("base")
	assert.False(t, baseErr.IsWarning())
	assert.True(t, baseErr.AsWarning().IsWarning())
	assert.True(t, baseErr.AsWarning().WithValue("v").WithFile("f").IsWarning())
}
