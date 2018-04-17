package halfpipe

import (
	"testing"

	errs "errors"

	"github.com/concourse/atc"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/defaults"
	"github.com/springernature/halfpipe/linters"
	"github.com/springernature/halfpipe/linters/errors"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func testController() Controller {
	return Controller{
		CurrentDir: "/pwd/foo",
		Defaulter:  defaults.DefaultValues,
	}
}

func TestReturnsErrorFromManifestReader(t *testing.T) {
	expectedError := errs.New("Meehp")
	c := testController()
	c.ManifestReader = func(dir string, fs afero.Afero) (man manifest.Manifest, err error) {
		err = expectedError
		return
	}

	pipeline, results := c.Process()

	assert.Empty(t, pipeline)
	assert.Len(t, results, 1)
	assert.Len(t, results[0].Errors, 1)
	assert.Equal(t, expectedError, results[0].Errors[0])
}

type fakeLinter struct {
	Error error
}

func (f fakeLinter) Lint(manifest manifest.Manifest) linters.LintResult {
	return linters.NewLintResult("fake", "url", []error{f.Error}, nil)
}

func TestAppliesAllLinters(t *testing.T) {
	c := testController()
	c.ManifestReader = func(dir string, fs afero.Afero) (man manifest.Manifest, err error) {
		return manifest.Manifest{}, nil
	}

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

func (f FakeRenderer) Render(manifest manifest.Manifest) atc.Config {
	return f.Config
}

func TestGivesBackAtcConfigWhenLinterPasses(t *testing.T) {
	c := testController()
	c.ManifestReader = func(dir string, fs afero.Afero) (man manifest.Manifest, err error) {
		return manifest.Manifest{}, nil
	}

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
