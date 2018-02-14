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

func (c Controller) readManifest() (model.Manifest, []error) {
	content, err := c.Fs.ReadFile(halfpipeFile)
	if err != nil {
		return model.Manifest{}, nil
	}

	if len(content) == 0 {
		return model.Manifest{}, []error{errors.NewFileError(halfpipeFile, "must not be empty")}
	}

	stringContent := string(content)
	man, errs := parser.Parse(stringContent)
	if len(errs) != 0 {
		return model.Manifest{}, errs
	}

	return man, nil
}

func (c Controller) Process() (atc.Config, []error) {
	if err := linters.CheckFile(c.Fs, halfpipeFile, false); err != nil {
		return atc.Config{}, []error{err}
	}

	manifest, errs := c.readManifest()
	if errs != nil {
		return atc.Config{}, errs
	}

	for _, linter := range c.Linters {
		errs = append(errs, linter.Lint(manifest)...)
	}

	if errs != nil {
		return atc.Config{}, errs
	}

	return c.Renderer.Render(manifest), nil

}
