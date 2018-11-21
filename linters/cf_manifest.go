package linters

import (
	goErrors "errors"
	"fmt"

	"strings"

	cfManifest "code.cloudfoundry.org/cli/util/manifest"
	"github.com/springernature/halfpipe/linters/errors"
	"github.com/springernature/halfpipe/linters/result"
	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/pipeline"
)

type cfManifestLinter struct {
	readCfManifest pipeline.CfManifestReader
}

func NewCfManifestLinter(cfManifestReader pipeline.CfManifestReader) cfManifestLinter {
	return cfManifestLinter{cfManifestReader}
}

func (linter cfManifestLinter) Lint(man manifest.Manifest) (result result.LintResult) {
	result.Linter = "CF Manifest Linter"
	result.DocsURL = "https://docs.halfpipe.io/cf-deployment/"

	var manifestPaths []string

	for _, task := range man.Tasks {
		switch t := task.(type) {
		case manifest.DeployCF:
			manifestPaths = append(manifestPaths, t.Manifest)
		}
	}

	for _, manifestPath := range manifestPaths {

		//skip linting if file provided as an artifact
		//task linter will warn that file needs to be generated in previous task
		if strings.HasPrefix(manifestPath, "../artifacts/") {
			return
		}

		apps, err := linter.readCfManifest(manifestPath, nil, nil)

		if err != nil {
			result.AddError(goErrors.New(fmt.Sprintf("cf-manifest error in %s, %s", manifestPath, err.Error())))
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

		if err := lintBuildpack(app); err != nil {
			result.AddWarning(err)
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

func lintBuildpack(man cfManifest.Application) (err error) {
	if man.Buildpack.Value != "" {
		return errors.NewDeprecatedBuildpackError()
	}
	return nil
}
