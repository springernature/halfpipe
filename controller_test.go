package halfpipe

import (
	"testing"

	"github.com/concourse/concourse/atc"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/defaults"
	"github.com/springernature/halfpipe/linters"
	"github.com/springernature/halfpipe/linters/linterrors"
	"github.com/springernature/halfpipe/linters/result"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

var validHalfpipeManifest = manifest.Manifest{
	Team:     "asdf",
	Pipeline: "my-pipeline",
	Tasks: manifest.TaskList{
		manifest.DockerCompose{},
	},
}

func testController() controller {
	var fs = afero.Afero{Fs: afero.NewMemMapFs()}
	_ = fs.MkdirAll("/pwd/foo/.git", 0777)
	return controller{
		defaulter: defaults.DefaultValues,
	}
}

func TestWorksForHalfpipeFileWithYMLExtension(t *testing.T) {
	c := testController()

	config := atc.Config{
		Resources: atc.ResourceConfigs{
			atc.ResourceConfig{
				Name: "Yolo",
			},
		},
	}
	c.renderer = FakeRenderer{Config: config}

	_, results := c.Process(validHalfpipeManifest)

	assert.Len(t, results.Error(), 0)
}

func TestWorksForHalfpipeFile(t *testing.T) {
	c := testController()

	config := atc.Config{
		Resources: atc.ResourceConfigs{
			atc.ResourceConfig{
				Name: "Yolo",
			},
		},
	}
	c.renderer = FakeRenderer{Config: config}

	_, results := c.Process(validHalfpipeManifest)

	assert.Len(t, results.Error(), 0)
}

type fakeLinter struct {
	Error error
}

func (f fakeLinter) Lint(manifest manifest.Manifest) result.LintResult {
	return result.NewLintResult("fake", "url", []error{f.Error}, nil)
}

func TestAppliesAllLinters(t *testing.T) {
	c := testController()

	linter1 := fakeLinter{linterrors.NewFileError("file", "is missing")}
	linter2 := fakeLinter{linterrors.NewMissingField("field")}
	c.linters = []linters.Linter{linter1, linter2}

	pipeline, results := c.Process(validHalfpipeManifest)

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

	config := atc.Config{
		Resources: atc.ResourceConfigs{
			atc.ResourceConfig{
				Name: "Yolo",
			},
		},
	}
	c.renderer = FakeRenderer{Config: config}

	pipeline, results := c.Process(validHalfpipeManifest)
	assert.Len(t, results, 0)
	assert.Equal(t, config, pipeline)
}
