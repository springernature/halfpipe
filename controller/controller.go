package controller

import (
	"github.com/concourse/atc"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/defaults"
	"github.com/springernature/halfpipe/helpers/file_checker"
	"github.com/springernature/halfpipe/linters"
	"github.com/springernature/halfpipe/model"
	"github.com/springernature/halfpipe/parser"
	"github.com/springernature/halfpipe/pipeline"
)

const halfpipeFile = ".halfpipe.io"

type Controller struct {
	Fs        afero.Afero
	Linters   []linters.Linter
	Renderer  pipeline.Renderer
	Defaulter defaults.Defaulter
}

func (c Controller) getManifest() (manifest model.Manifest, errors []error) {
	if err := file_checker.CheckFile(c.Fs, halfpipeFile, false); err != nil {
		errors = append(errors, err)
		return
	}

	content, err := c.Fs.ReadFile(halfpipeFile)
	if err != nil {
		errors = append(errors, err)
		return
	}

	stringContent := string(content)
	manifest, errs := parser.Parse(stringContent)
	if len(errs) != 0 {
		errors = append(errors, errs...)
		return
	}

	return
}

func (c Controller) Process() (config atc.Config, results model.LintResults) {

	manifest, errs := c.getManifest()
	if errs != nil {
		results = append(results, model.NewLintResult("Halfpipe", errs))
		return
	}

	manifest = c.Defaulter(manifest)

	for _, linter := range c.Linters {
		results = append(results, linter.Lint(manifest))
	}

	for _, lintResult := range results {
		if lintResult.HasErrors() {
			return
		}
	}
	config = c.Renderer.Render(manifest)
	return
}
