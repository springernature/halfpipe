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
	DefaultAndMap(man manifest.Manifest) (updated manifest.Manifest)
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
	defaultedManifest := c.defaulter.Apply(man)

	for _, linter := range c.linters {
		results = append(results, linter.Lint(defaultedManifest))
	}

	if results.HasErrors() {
		return
	}

	mappedManifest := c.mapper.Apply(defaultedManifest)

	config = c.renderer.Render(mappedManifest)
	return config, results
}

func (c controller) DefaultAndMap(man manifest.Manifest) (updated manifest.Manifest) {
	return c.mapper.Apply(c.defaulter.Apply(man))
}
