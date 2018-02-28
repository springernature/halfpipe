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
	"github.com/springernature/halfpipe/project"
)

const halfpipeFile = ".halfpipe.io"

type Controller struct {
	Fs        afero.Afero
	Project   project.Project
	Linters   []linters.Linter
	Renderer  pipeline.Renderer
	Defaulter defaults.Defaulter
}

func (c Controller) getManifest() (manifest model.Manifest, errors []error) {
	yaml, err := file_checker.ReadFile(c.Fs, halfpipeFile)
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

func (c Controller) Process() (config atc.Config, results model.LintResults) {

	manifest, errs := c.getManifest()
	if errs != nil {
		results = append(results, model.NewLintResult("Halfpipe", errs))
		return
	}

	if manifest.Repo.Uri == "" {
		manifest.Repo.Uri = c.Project.GitUri
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
	config = c.Renderer.Render(c.Project, manifest)
	return
}
