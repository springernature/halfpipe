package linters

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_ErrorWrapping(t *testing.T) {
	baseErr := NewError("bananas")
	err := baseErr.WithValue("are").WithValue("tasty").WithFile("fruit.txt")

	assert.ErrorIs(t, err, baseErr)
	assert.ErrorIs(t, err, baseErr.WithValue("are"))
	assert.ErrorIs(t, err, baseErr.WithValue("are").WithValue("tasty"))
	assert.ErrorIs(t, err, baseErr.WithValue("are").WithValue("tasty").WithFile("fruit.txt"))

	assert.NotErrorIs(t, err, baseErr.WithValue("tasty"))
}
