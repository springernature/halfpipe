package linters

import (
	"fmt"
	"github.com/springernature/halfpipe/cf"
	"github.com/springernature/halfpipe/config"
	"strings"

	"code.cloudfoundry.org/cli/util/manifestparser"
	"github.com/springernature/halfpipe/linters/linterrors"
	"github.com/springernature/halfpipe/linters/result"
	"github.com/springernature/halfpipe/manifest"
)

type cfManifestLinter struct {
	readCfManifest cf.ManifestReader
}

func NewCfManifestLinter(cfManifestReader cf.ManifestReader) cfManifestLinter {
	return cfManifestLinter{cfManifestReader}
}

func (linter cfManifestLinter) Lint(man manifest.Manifest) (result result.LintResult) {
	result.Linter = "CF Manifest"
	result.DocsURL = "https://ee.public.springernature.app/rel-eng/halfpipe/cf-deployment/"

	var tasks []manifest.DeployCF
	for _, task := range man.Tasks {
		switch t := task.(type) {
		case manifest.DeployCF:
			tasks = append(tasks, t)
		}
	}

	for _, task := range tasks {
		//skip linting if file provided as an artifact
		//task linter will warn that file needs to be generated in previous task
		manifestPath := task.Manifest
		if strings.HasPrefix(manifestPath, "../artifacts/") {
			return result
		}

		manifest, err := linter.readCfManifest(manifestPath, nil, nil)
		apps := manifest.Applications

		if err != nil {
			result.AddError(fmt.Errorf("cf-manifest error in %s, %s", manifestPath, err.Error()))
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
		result.AddError(lintDockerPush(task, app)...)
		result.AddWarning(lintBuildpack(app)...)
	}
	return result
}

func lintDockerPush(task manifest.DeployCF, app manifestparser.Application) (errs []error) {
	if app.Docker != nil {
		if task.DeployArtifact != "" {
			errs = append(errs, linterrors.NewDockerPushError(task.Manifest, "you cannot specify both 'deploy_artifact' in the task and 'docker' in the manifest"))
			return
		}

		if !strings.HasPrefix(app.Docker.Image, config.DockerRegistry) {
			errs = append(errs, linterrors.NewDockerPushError(task.Manifest, fmt.Sprintf("image must come from '%s'", config.DockerRegistry)))
			return
		}
	}

	return
}

func lintRoutes(manifestPath string, app manifestparser.Application) (errs []error) {
	if app.NoRoute {
		if app.RemainingManifestFields["routes"] != nil {
			errs = append(errs, linterrors.NewBadRoutesError(manifestPath, "you cannot specify both 'routes' and 'no-route'"))
			return errs
		}

		if app.HealthCheckType != "process" {
			errs = append(errs, linterrors.NewWrongHealthCheck(manifestPath, "if 'no-route' is true you must set 'health-check-type' to 'process'"))
			return errs
		}

		return errs
	}

	if app.RemainingManifestFields["routes"] == nil {
		errs = append(errs, linterrors.NewNoRoutesError(manifestPath, "app in cf Manifest must have at least 1 route defined or in case of a worker app you must set 'no-route' to true"))
		return errs
	}

	for _, route := range cf.Routes(app) {
		if strings.HasPrefix(route, "http://") || strings.HasPrefix(route, "https://") {
			errs = append(errs, linterrors.NewNoRoutesError(manifestPath, fmt.Sprintf("don't put http(s):// at the start of the route: '%s'", route)))
		}
	}

	return errs
}

func lintBuildpack(app manifestparser.Application) (errs []error) {
	if app.RemainingManifestFields["buildpack"] != nil {
		errs = append(errs, linterrors.NewDeprecatedBuildpackError())
	}

	buildpacks := cf.Buildpacks(app)

	if len(buildpacks) == 0 && app.Docker == nil {
		errs = append(errs, linterrors.NewMissingBuildpackError())
		return errs
	}

	for _, bp := range buildpacks {
		if strings.HasPrefix(bp, "http") && !strings.Contains(bp, "#") {
			errs = append(errs, linterrors.NewUnversionedBuildpackError(bp))
		}
	}

	return errs
}
