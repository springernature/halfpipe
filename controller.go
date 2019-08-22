package halfpipe

import (
	"github.com/concourse/concourse/atc"
	"github.com/springernature/halfpipe/defaults"
	"github.com/springernature/halfpipe/linters"
	"github.com/springernature/halfpipe/linters/result"
	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/parallel"
	"github.com/springernature/halfpipe/pipeline"
	"github.com/springernature/halfpipe/triggers"
)

type Controller interface {
	Process(man manifest.Manifest) (config atc.Config, results result.LintResults)
}

type controller struct {
	defaulter         defaults.Defaults
	merger            parallel.Merger
	triggerTranslator triggers.Translator
	linters           []linters.Linter
	renderer          pipeline.Renderer
}

func NewController(defaulter defaults.Defaults, merger parallel.Merger, triggerTranslator triggers.Translator, linters []linters.Linter, renderer pipeline.Renderer) Controller {
	return controller{
		defaulter:         defaulter,
		merger:            merger,
		triggerTranslator: triggerTranslator,
		linters:           linters,
		renderer:          renderer,
	}
}

func (c controller) Process(man manifest.Manifest) (config atc.Config, results result.LintResults) {
	updatedManifest := c.triggerTranslator.Translate(man)
	updatedManifest.Tasks = c.merger.MergeParallelTasks(man.Tasks)

	updatedManifest = c.defaulter.Update(updatedManifest)

	for _, linter := range c.linters {
		results = append(results, linter.Lint(updatedManifest))
	}

	if results.HasErrors() {
		return
	}

	config = c.renderer.Render(updatedManifest)
	return
}
