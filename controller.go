package halfpipe

import (
	"github.com/concourse/atc"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/defaults"
	"github.com/springernature/halfpipe/linters"
	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/pipeline"
)

type Controller struct {
	Fs             afero.Afero
	ManifestReader manifest.ManifestReader
	CurrentDir     string
	Defaulter      defaults.Defaults
	Linters        []linters.Linter
	Renderer       pipeline.Renderer
}

func (c Controller) Process() (config atc.Config, results linters.LintResults) {
	man, err := c.ManifestReader(c.CurrentDir, c.Fs)
	if err != nil {
		results = append(results, linters.NewLintResult("Halfpipe", "https://docs.halfpipe.io/docs/manifest/#Manifest", []error{err}, nil))
		return
	}

	man = c.Defaulter.Update(man)

	for _, linter := range c.Linters {
		results = append(results, linter.Lint(man))
	}

	if results.HasErrors() {
		return
	}

	config = c.Renderer.Render(man)
	return
}
