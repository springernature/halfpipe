package halfpipe

import (
	"github.com/springernature/halfpipe/defaults"
	"github.com/springernature/halfpipe/linters"
	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/mapper"
	"github.com/springernature/halfpipe/project"
)

type Controller interface {
	Process(man manifest.Manifest) Response
	DefaultAndMap(man manifest.Manifest) (updated manifest.Manifest, err error)
}

type Response struct {
	ConfigYaml  string
	Project     project.Data
	LintResults linters.LintResults
	Platform    manifest.Platform
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

func (c controller) Process(man manifest.Manifest) (response Response) {
	defaultedManifest := c.defaulter.Apply(man)

	for _, linter := range c.linters {
		response.LintResults = append(response.LintResults, linter.Lint(defaultedManifest))
	}

	if response.LintResults.HasErrors() {
		return
	}

	mappedManifest, err := c.mapper.Apply(defaultedManifest)
	if err != nil {
		response.LintResults = append(response.LintResults, linters.NewLintResult("Internal mapper", "", []error{err}))
		return
	}

	config, _ := c.renderer.Render(mappedManifest)
	response.ConfigYaml = config
	response.Project = c.defaulter.Project
	response.Platform = man.Platform
	return response
}

func (c controller) DefaultAndMap(man manifest.Manifest) (updated manifest.Manifest, err error) {
	return c.mapper.Apply(c.defaulter.Apply(man))
}
