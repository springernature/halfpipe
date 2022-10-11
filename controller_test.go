package halfpipe

import (
	"errors"
	"testing"

	"github.com/springernature/halfpipe/mapper"
	"github.com/springernature/halfpipe/project"

	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/defaults"
	"github.com/springernature/halfpipe/linters"
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

	response := c.Process(validHalfpipeManifest)

	assert.Len(t, response.LintResults.Error(), 0)
}

func TestWorksForHalfpipeFile(t *testing.T) {
	c := testController()
	response := c.Process(validHalfpipeManifest)

	assert.Len(t, response.LintResults.Error(), 0)
}

type fakeLinter struct {
	Error error
}

func (f fakeLinter) Lint(manifest manifest.Manifest) linters.LintResult {
	return linters.NewLintResult("fake", "url", []error{f.Error})
}

func TestAppliesAllLinters(t *testing.T) {
	c := testController()

	linter1 := fakeLinter{errors.New("error from linter1")}
	linter2 := fakeLinter{errors.New("error from linter2")}
	c.linters = []linters.Linter{linter1, linter2}

	response := c.Process(validHalfpipeManifest)

	assert.Empty(t, response.ConfigYaml)
	assert.Len(t, response.LintResults, 2)
	assert.Equal(t, linter1.Error, response.LintResults[0].Errors[0])
	assert.Equal(t, linter2.Error, response.LintResults[1].Errors[0])
}

func TestGivesBackConfigWhenLinterPasses(t *testing.T) {
	c := testController()

	response := c.Process(validHalfpipeManifest)
	assert.Len(t, response.LintResults, 0)
	assert.Equal(t, "fake output", response.ConfigYaml)
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
	response := c.Process(validHalfpipeManifest)

	assert.Len(t, response.LintResults, 1)
	assert.True(t, response.LintResults.HasErrors())
	assert.False(t, response.LintResults.HasWarnings())
}
