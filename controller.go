package halfpipe

import (
	"github.com/concourse/concourse/atc"
	"github.com/springernature/halfpipe/defaults"
	"github.com/springernature/halfpipe/linters"
	"github.com/springernature/halfpipe/linters/result"
	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/pipeline"
)

type Controller interface {
	Process(man manifest.Manifest) (config atc.Config, results result.LintResults)
}

type controller struct {
	defaulter defaults.Defaults
	linters   []linters.Linter
	renderer  pipeline.Renderer
}

func NewController(defaulter defaults.Defaults, linters []linters.Linter, renderer pipeline.Renderer) Controller {
	return controller{
		defaulter: defaulter,
		linters:   linters,
		renderer:  renderer,
	}
}

func (c controller) Process(man manifest.Manifest) (config atc.Config, results result.LintResults) {
	updatedManifest := c.defaulter.Apply(man)

	for _, linter := range c.linters {
		results = append(results, linter.Lint(updatedManifest))
	}

	if results.HasErrors() {
		return
	}

	config = c.renderer.Render(updatedManifest)
	return config, results
}
