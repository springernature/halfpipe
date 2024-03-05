package linters

import (
	"fmt"
	"github.com/springernature/halfpipe/cf"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/renderers/shared/secrets"
	"golang.org/x/exp/slices"
	"strings"

	"code.cloudfoundry.org/cli/util/manifestparser"
	"github.com/springernature/halfpipe/manifest"
)

func LintCfManifest(task manifest.DeployCF, readCfManifest cf.ManifestReader) (errs []error) {

	//skip linting if file provided as an artifact
	//task linter will warn that file needs to be generated in previous task
	if strings.HasPrefix(task.Manifest, "../artifacts/") {
		return errs
	}

	cfManifest, err := readCfManifest(task.Manifest, nil, nil)
	apps := cfManifest.Applications

	if err != nil {
		errs = append(errs, ErrFileInvalid.WithValue(err.Error()).WithFile(task.Manifest))
		return errs
	}

	if len(apps) != 1 {
		errs = append(errs, ErrCFMultipleApps.WithFile(task.Manifest))
		return errs
	}

	app := apps[0]
	if app.Name == "" {
		errs = append(errs, ErrCFMissingName.WithFile(task.Manifest))
	}

	errs = append(errs, lintRoutes(task, app)...)
	errs = append(errs, lintCandidateAppRoute(task, cfManifest)...)
	errs = append(errs, lintDockerPush(task, app)...)
	errs = append(errs, lintBuildpack(app, task.Manifest)...)
	errs = append(errs, lintLabels(app)...)

	return errs
}

func lintCandidateAppRoute(task manifest.DeployCF, m manifestparser.Manifest) (errs []error) {
	if secrets.IsSecret(task.Space) {
		return errs
	}

	testRouteHost := fmt.Sprintf("%s-%s-CANDIDATE", m.GetFirstApp().Name, task.Space)

	if len(testRouteHost) > 64 {
		errs = append(errs, ErrCFCandidateRouteTooLong.WithValue(fmt.Sprintf("%s length is %v", testRouteHost, len(testRouteHost))))
		return
	}

	return errs
}

func lintDockerPush(task manifest.DeployCF, app manifestparser.Application) (errs []error) {
	if app.Docker != nil {
		if task.DeployArtifact != "" {
			errs = append(errs, ErrCFArtifactAndDocker.WithFile(task.Manifest))
			return
		}

		if !strings.HasPrefix(app.Docker.Image, config.DockerRegistry) {
			errs = append(errs, ErrUnsupportedRegistry.WithValue(app.Docker.Image).WithFile(task.Manifest))
			return
		}
	}

	return errs
}

func lintRoutes(task manifest.DeployCF, app manifestparser.Application) (errs []error) {
	if app.NoRoute {
		if app.RemainingManifestFields["routes"] != nil {
			errs = append(errs, ErrCFRoutesAndNoRoute.WithFile(task.Manifest))
			return errs
		}

		if app.HealthCheckType != "process" {
			errs = append(errs, ErrCFNoRouteHealthcheck.WithFile(task.Manifest))
			return errs
		}

		return errs
	}

	if app.RemainingManifestFields["routes"] == nil {
		errs = append(errs, ErrCFMissingRoutes.WithFile(task.Manifest))
		return errs
	}

	for _, route := range cf.Routes(app) {
		if strings.HasPrefix(route, "http://") || strings.HasPrefix(route, "https://") {
			errs = append(errs, ErrCFRouteScheme.WithValue(route).WithFile(task.Manifest))
		}
	}

	if task.SSORoute != "" {
		hasSSORoute := slices.Contains(cf.Routes(app), task.SSORoute)
		if !hasSSORoute {
			errs = append(errs, ErrCFRouteMissing.WithValue(task.SSORoute).WithFile(task.Manifest))
		}
	}

	return errs
}

func lintBuildpack(app manifestparser.Application, manifestPath string) (errs []error) {
	if app.RemainingManifestFields["buildpack"] != nil {
		errs = append(errs, ErrCFBuildpackDeprecated.WithFile(manifestPath).AsWarning())
	}

	buildpacks := cf.Buildpacks(app)

	if len(buildpacks) == 0 && app.Docker == nil {
		errs = append(errs, ErrCFBuildpackMissing.WithFile(manifestPath).AsWarning())
		return errs
	}

	for _, bp := range buildpacks {
		if strings.HasPrefix(bp, "http") && !strings.Contains(bp, "#") {
			errs = append(errs, ErrCFBuildpackUnversioned.WithValue(bp).WithFile(manifestPath).AsWarning())
		}
	}

	return errs
}
func lintLabels(app manifestparser.Application) (errs []error) {
	if app.RemainingManifestFields["metadata"] != nil {
		metadata := app.RemainingManifestFields["metadata"].(map[any]any)
		if metadata["labels"] != nil {
			labels := metadata["labels"].(map[any]any)
			_, found := labels["team"]
			if found {
				errs = append(errs, ErrCFTeamLabelWillBeOverwritten)
			}
		}
	}
	return
}
