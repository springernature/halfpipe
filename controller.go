package halfpipe

import (
	"github.com/springernature/halfpipe/defaults"
	"github.com/springernature/halfpipe/linters"
	"github.com/springernature/halfpipe/linters/result"
	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/mapper"
)

type Controller interface {
	Process(man manifest.Manifest) (config string, results result.LintResults)
	DefaultAndMap(man manifest.Manifest) (updated manifest.Manifest, err error)
}

type Renderer interface {
	Render(manifest manifest.Manifest) (string, error)
}

type controller struct {
	defaulter defaults.Defaults
	mapper    mapper.Mapper
	linters   []linters.Linter
	renderer  Renderer
}

func NewController(defaulter defaults.Defaults, mapper mapper.Mapper, linters []linters.Linter, renderer Renderer) Controller {
	return controller{
		defaulter: defaulter,
		mapper:    mapper,
		linters:   linters,
		renderer:  renderer,
	}
}

func (c controller) Process(man manifest.Manifest) (config string, results result.LintResults) {
	defaultedManifest := c.defaulter.Apply(man)

	for _, linter := range c.linters {
		results = append(results, linter.Lint(defaultedManifest))
	}

	if results.HasErrors() {
		return
	}

	mappedManifest, err := c.mapper.Apply(defaultedManifest)
	if err != nil {
		results = append(results, result.LintResult{Linter: "Internal mapper", Errors: []error{err}})
		return
	}

	config, _ = c.renderer.Render(mappedManifest)
	return config, results
}

func (c controller) DefaultAndMap(man manifest.Manifest) (updated manifest.Manifest, err error) {
	return c.mapper.Apply(c.defaulter.Apply(man))
}
