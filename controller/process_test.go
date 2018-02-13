package controller

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestHalfpipeFileDoesNotExist(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	result := Process(fs)
	assert.IsType(t, errors.New(""), result)
}

func TestHalfpipeFileExists(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile(".halfpipe.io", []byte("hello"), 0777)
	result := Process(fs)
	assert.Nil(t, result)
}
