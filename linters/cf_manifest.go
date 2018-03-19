package linters

import (
	cfManifest "code.cloudfoundry.org/cli/util/manifest"
	"github.com/springernature/halfpipe/linters/errors"
	"github.com/springernature/halfpipe/manifest"
)

type cfManifestLinter struct {
	rManifest func(string) ([]cfManifest.Application, error)
}

func NewCfManifestLinter(readManifest func(string) ([]cfManifest.Application, error)) cfManifestLinter {
	return cfManifestLinter{readManifest}
}

func (linter cfManifestLinter) Lint(man manifest.Manifest) (result LintResult) {
	result.Linter = "CF Manifest Linter"
	var manifestPaths []string

	for _, task := range man.Tasks {
		switch t := task.(type) {
		case manifest.DeployCF:
			manifestPaths = append(manifestPaths, t.Manifest)
		}
	}

	for _, manifestPath := range manifestPaths {
		apps, err := linter.rManifest(manifestPath)

		if err != nil {
			result.AddError(err)
			return
		}

		if len(apps) != 1 {
			result.AddError(errors.NewTooManyAppsError(manifestPath, "cf manifest must have 1 application defined"))
		}

		if apps[0].Name == "" {
			result.AddError(errors.NewNoNameError(manifestPath, "app in cf manifest must have a name"))
		}

		if len(apps[0].Routes) == 0 {
			result.AddError(errors.NewNoRoutesError(manifestPath, "app in cf Manifest must have at least 1 route defined"))
		}
	}
	return
}
