package controller

import (
	"github.com/concourse/atc"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/errors"
	"github.com/springernature/halfpipe/linters"
	"github.com/springernature/halfpipe/model"
	"github.com/springernature/halfpipe/parser"
	"github.com/springernature/halfpipe/pipeline"
)

const halfpipeFile = ".halfpipe.io"

type Controller struct {
	Fs       afero.Afero
	Linters  []linters.Linter
	Renderer pipeline.Renderer
}

func (c Controller) readManifest() (manifest model.Manifest, errors []error) {
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

func (c Controller) Process() (config atc.Config, results errors.LintResults) {
	if err := linters.CheckFile(c.Fs, halfpipeFile, false); err != nil {
		results = append(results, errors.LintResult{"Halfpipe", []error{err}})
		return
	}

	manifest, errs := c.readManifest()
	if errs != nil {
		results = append(results, errors.LintResult{"Halfpipe", errs})
		return
	}

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
