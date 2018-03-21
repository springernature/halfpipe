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
	result.DocsURL = "https://docs.halfpipe.io/docs/cf-deployment/"

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
			result.AddError(errors.NewTooManyAppsError(manifestPath, "cf manifest must have exactly 1 application defined"))
			return
		}

		app := apps[0]
		if app.Name == "" {
			result.AddError(errors.NewNoNameError(manifestPath, "app in cf manifest must have a name"))
		}

		if err := lintRoutes(manifestPath, app); err != nil {
			result.AddError(err)
		}
	}
	return
}

func lintRoutes(manifestPath string, man cfManifest.Application) (err error) {
	if man.NoRoute {
		if len(man.Routes) != 0 {
			return errors.NewBadRoutesError(manifestPath, "You cannot specify both 'routes' and 'no-route'")
		}

		if man.HealthCheckType != "process" {
			return errors.NewWrongHealthCheck(manifestPath, "If 'no-route' is true you must set 'health-check-type' to 'process'")
		}

		return
	}

	if len(man.Routes) == 0 {
		return errors.NewNoRoutesError(manifestPath, "app in cf Manifest must have at least 1 route defined or in case of a worker app you must set 'no-route' to true")
	}

	return
}
