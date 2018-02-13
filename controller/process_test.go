package controller

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/linters"
	"github.com/springernature/halfpipe/model"
	"github.com/stretchr/testify/assert"
)

func setup() Controller {
	var fs = afero.Afero{Fs: afero.NewMemMapFs()}
	return Controller{Fs: fs}
}

func TestProcessDoesNothingWhenFileDoesntExistt(t *testing.T) {
	c := setup()
	pipeline, errs := c.Process()

	assert.Empty(t, pipeline)
	assert.Len(t, errs, 1)
	assert.IsType(t, errors.New(""), errs[0])
}

type fakeLinter struct {
	Error error
}

func (f fakeLinter) Lint(manifest model.Manifest) []error {
	return []error{f.Error}
}

func TestAppliesAllLinters(t *testing.T) {
	c := setup()
	c.Fs.WriteFile(".halfpipe.io", []byte("team:foo"), 0777)

	e1 := errors.New("Error1")
	e2 := errors.New("Error2")
	error1 := fakeLinter{e1}
	error2 := fakeLinter{e2}
	c.Linters = []linters.Linter{error1, error2}

	pipeline, errs := c.Process()

	assert.Empty(t, pipeline)
	assert.Len(t, errs, 2)
	assert.Equal(t, e1, errs[0])
	assert.Equal(t, e2, errs[1])
}

//
//func TestHalfpipeFileExists(t *testing.T) {
//	fs := afero.Afero{Fs: afero.NewMemMapFs()}
//	fs.WriteFile(".halfpipe.io", []byte("hello"), 0777)
//	result := Process(fs)
//	assert.Nil(t, result)
//}
//
//func TestParseManifestErrorsOutOnEmpty(t *testing.T) {
//	fs := afero.Afero{Fs: afero.NewMemMapFs()}
//	fs.WriteFile(".halfpipe.io", []byte(""), 0777)
//
//	result, err := ParseManifest(fs)
//
//	assert.Equal(t, result, model.Manifest{})
//	assert.IsType(t, errors.New(""), err)
//}
//
//func TestParseManifestErrorsOutBadYaml(t *testing.T) {
//	fs := afero.Afero{Fs: afero.NewMemMapFs()}
//	fs.WriteFile(".halfpipe.io", []byte("kehe====asd"), 0777)
//
//	result, err := ParseManifest(fs)
//
//	assert.Equal(t, result, model.Manifest{})
//	assert.IsType(t, model.ParseError{}, err)
//}
//
//func TestParseManifestGivesBackManifest(t *testing.T) {
//	fs := afero.Afero{Fs: afero.NewMemMapFs()}
//	fs.WriteFile(".halfpipe.io", []byte("team: poop"), 0777)
//
//	result, err := ParseManifest(fs)
//
//	assert.Equal(t, "poop", result.Team)
//	assert.Nil(t, err)
//}
