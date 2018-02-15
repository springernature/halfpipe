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

func (c Controller) halfpipeExists() []error {
	exists, err := c.Fs.Exists(halfpipeFile)
	if err != nil {
		return []error{err}
	}
	if !exists {
		return []error{errors.NewFileError(halfpipeFile, "is missing")}
	}
	return nil
}

func (c Controller) readManifest() (model model.Manifest, errors []error) {
	content, err := c.Fs.ReadFile(halfpipeFile)
	if err != nil {
		errors = append(errors, err)
		return
	}

	stringContent := string(content)
	model, errs := parser.Parse(stringContent)
	if len(errs) != 0 {
		errors = append(errors, errs...)
		return
	}

	return
}

func (c Controller) Process() (config atc.Config, result errors.LintResults) {
	if err := linters.CheckFile(c.Fs, halfpipeFile, false); err != nil {
		result = append(result, errors.LintResult{"Halfpipe", []error{err}})
		return
	}

	manifest, errs := c.readManifest()
	if errs != nil {
		result = append(result, errors.LintResult{"Halfpipe", errs})
		return
	}

	for _, linter := range c.Linters {
		result = append(result, linter.Lint(manifest))
	}

	for _, lintResult := range result {
		if lintResult.HasErrors() {
			return
		}
	}
	config = c.Renderer.Render(manifest)
	return

}
