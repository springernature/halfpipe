package halfpipe

import (
	"path/filepath"

	"github.com/concourse/atc"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/defaults"
	"github.com/springernature/halfpipe/linters"
	"github.com/springernature/halfpipe/linters/file_checker"
	"github.com/springernature/halfpipe/parser"
	"github.com/springernature/halfpipe/pipeline"
)

type Controller struct {
	Fs         afero.Afero
	CurrentDir string
	Linters    []linters.Linter
	Renderer   pipeline.Renderer
	Defaulter  defaults.Defaulter
}

func (c Controller) getManifest() (manifest parser.Manifest, errors []error) {
	yaml, err := file_checker.ReadFile(c.Fs, filepath.Join(c.CurrentDir, config.HalfpipeFile))
	if err != nil {
		errors = append(errors, err)
		return
	}

	manifest, errs := parser.Parse(yaml)
	if len(errs) != 0 {
		errors = append(errors, errs...)
		return
	}

	return
}

func (c Controller) Process() (config atc.Config, results linters.LintResults) {

	manifest, errs := c.getManifest()
	if errs != nil {
		results = append(results, linters.NewLintResult("Halfpipe", errs))
		return
	}

	project, err := defaults.NewConfig(c.Fs).Parse(c.CurrentDir)
	if err != nil {
		results = append(results, linters.NewLintResult("Halfpipe", []error{err}))
		return
	}

	manifest = c.Defaulter(manifest, project)

	for _, linter := range c.Linters {
		results = append(results, linter.Lint(manifest))
	}

	if results.HasErrors() {
		return
	}

	config = c.Renderer.Render(manifest)
	return
}
