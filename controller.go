package halfpipe

import (
	"github.com/concourse/concourse/atc"
	"github.com/springernature/halfpipe/defaults"
	"github.com/springernature/halfpipe/linters"
	"github.com/springernature/halfpipe/linters/result"
	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/mapper"
	"github.com/springernature/halfpipe/pipeline"
)

type Controller interface {
	Process(man manifest.Manifest) (config atc.Config, results result.LintResults)
}

type controller struct {
	defaulter defaults.Defaults
	mapper    mapper.Mapper
	linters   []linters.Linter
	renderer  pipeline.Renderer
}

func NewController(defaulter defaults.Defaults, mapper mapper.Mapper, linters []linters.Linter, renderer pipeline.Renderer) Controller {
	return controller{
		defaulter: defaulter,
		mapper:    mapper,
		linters:   linters,
		renderer:  renderer,
	}
}

func (c controller) Process(man manifest.Manifest) (config atc.Config, results result.LintResults) {
	// Apply defaults for values
	updatedManifest := c.defaulter.Apply(man)

	// Map fields to other fields
	updatedManifest = c.mapper.Apply(updatedManifest)

	for _, linter := range c.linters {
		results = append(results, linter.Lint(updatedManifest))
	}

	if results.HasErrors() {
		return
	}

	config = c.renderer.Render(updatedManifest)
	return config, results
}
