package controller

import (
	"testing"

	"github.com/concourse/atc"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/errors"
	"github.com/springernature/halfpipe/linters"
	"github.com/springernature/halfpipe/model"
	"github.com/stretchr/testify/assert"
)

func setup() Controller {
	var fs = afero.Afero{Fs: afero.NewMemMapFs()}
	return Controller{Fs: fs}
}

func TestProcessDoesNothingWhenFileDoesntExist(t *testing.T) {
	c := setup()
	pipeline, results := c.Process()

	assert.Empty(t, pipeline)
	assert.Len(t, results, 1)
	assert.IsType(t, errors.FileError{}, results[0].Errors[0])
}

func TestProcessDoesNothingWhenManifestIsEmpty(t *testing.T) {
	c := setup()
	c.Fs.WriteFile(".halfpipe.io", []byte(""), 0777)
	pipeline, results := c.Process()

	assert.Empty(t, pipeline)
	assert.Len(t, results, 1)
	assert.IsType(t, errors.FileError{}, results[0].Errors[0])
}

type fakeLinter struct {
	Error error
}

func (f fakeLinter) Lint(manifest model.Manifest) errors.LintResult {
	return errors.LintResult{
		Errors: []error{f.Error},
	}
}

func TestAppliesAllLinters(t *testing.T) {
	c := setup()
	c.Fs.WriteFile(".halfpipe.io", []byte("team: asd"), 0777)

	e1 := errors.NewFileError("file", "is missing")
	e2 := errors.NewMissingField("field")
	error1 := fakeLinter{e1}
	error2 := fakeLinter{e2}
	c.Linters = []linters.Linter{error1, error2}

	pipeline, results := c.Process()

	assert.Empty(t, pipeline)
	assert.Len(t, results, 2)
	assert.Equal(t, e1, results[0].Errors[0])
	assert.Equal(t, e2, results[1].Errors[0])
}

type FakeRenderer struct {
	Config atc.Config
}

func (f FakeRenderer) Render(manifest model.Manifest) atc.Config {
	return f.Config
}

func TestGivesBackAtcConfigWhenLinterPasses(t *testing.T) {
	c := setup()
	c.Fs.WriteFile(".halfpipe.io", []byte("team: asd"), 0777)

	config := atc.Config{
		Resources: atc.ResourceConfigs{
			atc.ResourceConfig{
				Name: "Yolo",
			},
		},
	}
	c.Renderer = FakeRenderer{Config: config}

	pipeline, results := c.Process()
	assert.Len(t, results, 0)
	assert.Equal(t, config, pipeline)
}
