package halfpipe

import (
	"errors"
	"github.com/springernature/halfpipe/mapper"
	"github.com/springernature/halfpipe/project"
	"testing"

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

type fakeRenderer struct{}

func (f fakeRenderer) Render(manifest manifest.Manifest) (string, error) {
	return "fake output", nil
}

func testController() controller {
	var fs = afero.Afero{Fs: afero.NewMemMapFs()}
	_ = fs.MkdirAll("/pwd/foo/.git", 0777)
	return controller{
		defaulter: defaults.New(defaults.Concourse, project.Data{}),
		mapper:    mapper.New(),
		renderer:  fakeRenderer{},
	}
}

func TestWorksForHalfpipeFileWithYMLExtension(t *testing.T) {
	c := testController()

	_, results := c.Process(validHalfpipeManifest)

	assert.Len(t, results.Error(), 0)
}

func TestWorksForHalfpipeFile(t *testing.T) {
	c := testController()
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

func TestGivesBackConfigWhenLinterPasses(t *testing.T) {
	c := testController()

	pipeline, results := c.Process(validHalfpipeManifest)
	assert.Len(t, results, 0)
	assert.Equal(t, "fake output", pipeline)
}

type FakeMapper struct {
	err error
}

func (f FakeMapper) Apply(original manifest.Manifest) (updated manifest.Manifest, err error) {
	return original, f.err
}

func TestGivesBackABadTestResultWhenAMapperFails(t *testing.T) {
	c := testController()
	c.mapper = FakeMapper{err: errors.New("blurgh")}
	_, results := c.Process(validHalfpipeManifest)

	assert.Len(t, results, 1)
	assert.True(t, results.HasErrors())
	assert.False(t, results.HasWarnings())
}
