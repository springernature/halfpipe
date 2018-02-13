package controller

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/linters"
	"github.com/springernature/halfpipe/model"
	"github.com/stretchr/testify/assert"
	"github.com/concourse/atc"
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

func TestProcessDoesNothingWhenManifestIsEmpty(t *testing.T) {
	c := setup()
	c.Fs.WriteFile(".halfpipe.io", []byte(""), 0777)
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
	c.Fs.WriteFile(".halfpipe.io", []byte("team: asd"), 0777)

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

	pipeline, errs := c.Process()
	assert.Len(t, errs, 0)
	assert.Equal(t, config, pipeline)
}
