package controller

import (
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
	Fs        afero.Afero
	Project   defaults.Project
	Linters   []linters.Linter
	Renderer  pipeline.Renderer
	Defaulter defaults.Defaulter
}

func (c Controller) getManifest() (manifest parser.Manifest, errors []error) {
	yaml, err := file_checker.ReadFile(c.Fs, config.HalfpipeFile)
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

	if manifest.Repo.Uri == "" {
		manifest.Repo.Uri = c.Project.GitUri
	}

	manifest = c.Defaulter(manifest)

	for _, linter := range c.Linters {
		results = append(results, linter.Lint(manifest))
	}

	if results.HasErrors() {
		return
	}

	config = c.Renderer.Render(c.Project, manifest)
	return
}
