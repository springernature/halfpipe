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

func testController() Controller {
	var fs = afero.Afero{Fs: afero.NewMemMapFs()}
	return Controller{
		Fs:        fs,
		Defaulter: func(m model.Manifest) model.Manifest { return m },
	}
}

func TestProcessDoesNothingWhenFileDoesNotExist(t *testing.T) {
	c := testController()
	pipeline, results := c.Process()

	assert.Empty(t, pipeline)
	assert.Len(t, results, 1)
	assert.IsType(t, errors.FileError{}, results[0].Errors[0])
}

func TestProcessDoesNothingWhenManifestIsEmpty(t *testing.T) {
	c := testController()
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
	return errors.NewLintResult("fake", []error{f.Error})
}

func TestAppliesAllLinters(t *testing.T) {
	c := testController()
	c.Fs.WriteFile(".halfpipe.io", []byte("team: asd"), 0777)

	linter1 := fakeLinter{errors.NewFileError("file", "is missing")}
	linter2 := fakeLinter{errors.NewMissingField("field")}
	c.Linters = []linters.Linter{linter1, linter2}

	pipeline, results := c.Process()

	assert.Empty(t, pipeline)
	assert.Len(t, results, 2)
	assert.Equal(t, linter1.Error, results[0].Errors[0])
	assert.Equal(t, linter2.Error, results[1].Errors[0])
}

type FakeRenderer struct {
	Config atc.Config
}

func (f FakeRenderer) Render(manifest model.Manifest) atc.Config {
	return f.Config
}

func TestGivesBackAtcConfigWhenLinterPasses(t *testing.T) {
	c := testController()
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

type fakeLinterFunc struct {
	LintFunc func(model.Manifest) errors.LintResult
}

func (f fakeLinterFunc) Lint(manifest model.Manifest) errors.LintResult {
	return f.LintFunc(manifest)
}

func TestCallsTheDefaultsUpdater(t *testing.T) {
	c := testController()
	c.Fs.WriteFile(".halfpipe.io", []byte("team: before"), 0777)

	c.Defaulter = func(m model.Manifest) model.Manifest {
		m.Team = "after"
		return m
	}

	//very hacky - use a linter to check the manifest has been updated
	linter := fakeLinterFunc{func(m model.Manifest) errors.LintResult {
		return errors.NewLintResult("fake", []error{errors.NewInvalidField("team", m.Team)})
	}}
	c.Linters = []linters.Linter{linter}

	_, results := c.Process()

	assert.Equal(t, "after", results[0].Errors[0].(errors.InvalidField).Reason)
}
