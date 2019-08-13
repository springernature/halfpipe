package halfpipe

import (
	"github.com/concourse/concourse/atc"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/defaults"
	"github.com/springernature/halfpipe/linters"
	"github.com/springernature/halfpipe/linters/filechecker"
	"github.com/springernature/halfpipe/linters/result"
	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/parallel"
	"github.com/springernature/halfpipe/pipeline"
	"github.com/springernature/halfpipe/triggers"
	"path"
)

type Controller struct {
	Fs                afero.Afero
	CurrentDir        string
	Defaulter         defaults.Defaults
	Merger            parallel.Merger
	TriggerTranslator triggers.Translator
	Linters           []linters.Linter
	Renderer          pipeline.Renderer
	HalfpipeFilePath  string
}

func (c Controller) getManifest() (man manifest.Manifest, errors []error) {
	yaml, err := filechecker.ReadFile(c.Fs, path.Join(c.CurrentDir, c.HalfpipeFilePath))
	if err != nil {
		errors = append(errors, err)
		return
	}

	man, errs := manifest.Parse(yaml)
	if len(errs) != 0 {
		errors = append(errors, errs...)
		return
	}

	return
}

func (c Controller) Process() (config atc.Config, results result.LintResults) {
	man, errs := c.getManifest()
	if errs != nil {
		results = append(results, result.NewLintResult("Halfpipe", "https://docs.halfpipe.io/manifest/", errs, nil))
		return
	}

	man = c.Defaulter.Update(man)
	man.Tasks = c.Merger.MergeParallelTasks(man.Tasks)

	for _, linter := range c.Linters {
		results = append(results, linter.Lint(man))
	}

	if results.HasErrors() {
		return
	}
	man = c.TriggerTranslator.Translate(man)

	config = c.Renderer.Render(man)
	return
}
