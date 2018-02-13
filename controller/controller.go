package controller

import (
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/linters"
	"github.com/springernature/halfpipe/model"
	"github.com/springernature/halfpipe/parser"
	"github.com/concourse/atc"
	"github.com/springernature/halfpipe/pipeline"
)

type Controller struct {
	Fs       afero.Afero
	Linters  []linters.Linter
	Renderer pipeline.Renderer
}

func (c Controller) halfpipeExists() []error {
	exists, err := c.Fs.Exists(".halfpipe.io")
	if err != nil {
		return []error{err}
	}
	if !exists {
		return []error{errors.New(".halfpipe.io does not exist")}
	}
	return nil
}

func (c Controller) readManifest() (model.Manifest, []error) {
	content, err := c.Fs.ReadFile(".halfpipe.io")
	if err != nil {
		return model.Manifest{}, nil
	}

	if len(content) == 0 {
		return model.Manifest{}, []error{errors.New(".halfpipe.io must not be empty")}
	}

	stringContent := string(content)
	man, errs := parser.Parse(stringContent)
	if len(errs) != 0 {
		return model.Manifest{}, errs
	}

	return man, nil
}

func (c Controller) Process() (atc.Config, []error) {
	errs := c.halfpipeExists()
	if errs != nil {
		return atc.Config{}, errs
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
