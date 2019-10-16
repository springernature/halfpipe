package linters

import (
	goErrors "errors"
	"fmt"

	"strings"

	cfManifest "code.cloudfoundry.org/cli/util/manifest"
	"github.com/springernature/halfpipe/linters/linterrors"
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
			return result
		}

		apps, err := linter.readCfManifest(manifestPath, nil, nil)

		if err != nil {
			result.AddError(goErrors.New(fmt.Sprintf("cf-manifest error in %s, %s", manifestPath, err.Error())))
			return result
		}

		if len(apps) != 1 {
			result.AddError(linterrors.NewTooManyAppsError(manifestPath, "cf manifest must have exactly 1 application defined"))
			return result
		}

		app := apps[0]
		if app.Name == "" {
			result.AddError(linterrors.NewNoNameError(manifestPath, "app in cf manifest must have a name"))
		}

		result.AddError(lintRoutes(manifestPath, app)...)
		result.AddWarning(lintBuildpack(app)...)
	}
	return result
}

func lintRoutes(manifestPath string, man cfManifest.Application) (errs []error) {
	if man.NoRoute {
		if len(man.Routes) != 0 {
			errs = append(errs, linterrors.NewBadRoutesError(manifestPath, "You cannot specify both 'routes' and 'no-route'"))
			return errs
		}

		if man.HealthCheckType != "process" {
			errs = append(errs, linterrors.NewWrongHealthCheck(manifestPath, "If 'no-route' is true you must set 'health-check-type' to 'process'"))
			return errs
		}

		return errs
	}

	if len(man.Routes) == 0 {
		errs = append(errs, linterrors.NewNoRoutesError(manifestPath, "app in cf Manifest must have at least 1 route defined or in case of a worker app you must set 'no-route' to true"))
		return errs
	}

	for _, route := range man.Routes {
		if strings.HasPrefix(route, "http://") || strings.HasPrefix(route, "https://") {
			errs = append(errs, linterrors.NewNoRoutesError(manifestPath, fmt.Sprintf("Don't put http(s):// at the start of the route: '%s'", route)))
		}
	}

	return errs
}

func lintBuildpack(man cfManifest.Application) (errs []error) {
	if man.Buildpack.Value != "" {
		errs = append(errs, linterrors.NewDeprecatedBuildpackError())
	}
	return errs
}
