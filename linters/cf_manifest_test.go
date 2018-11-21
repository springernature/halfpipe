package linters

import (
	"code.cloudfoundry.org/cli/types"
	errors2 "github.com/springernature/halfpipe/linters/errors"
	"testing"

	cfManifest "code.cloudfoundry.org/cli/util/manifest"
	"errors"
	"github.com/cloudfoundry/bosh-cli/director/template"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func manifestReader(applications []cfManifest.Application, err error) func(pathToManifest string, pathsToVarsFiles []string, vars []template.VarKV) ([]cfManifest.Application, error) {
	return func(pathToManifest string, pathsToVarsFiles []string, vars []template.VarKV) ([]cfManifest.Application, error) {
		return applications, err
	}
}

func TestNoCfDeployTasks(t *testing.T) {
	man := manifest.Manifest{}

	linter := cfManifestLinter{
		readCfManifest: manifestReader(nil, nil),
	}

	result := linter.Lint(man)
	assert.Len(t, result.Errors, 0)
}

//
func TestOneCfDeployTask(t *testing.T) {
	apps := []cfManifest.Application{
		{
			Name:   "appName",
			Routes: []string{"route"},
		},
	}
	linter := cfManifestLinter{readCfManifest: manifestReader(apps, nil)}

	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DeployCF{
				Manifest: "manifest.yml",
			},
		},
	}

	result := linter.Lint(man)
	assert.Len(t, result.Errors, 0)
}

func TestOneCfDeployTaskWithInvalidManifest(t *testing.T) {
	expectedErr := errors.New("invalid manifest error")
	linter := cfManifestLinter{readCfManifest: manifestReader(nil, expectedErr)}

	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DeployCF{
				Manifest: "manifest.yml",
			},
		},
	}

	result := linter.Lint(man)
	assert.Len(t, result.Errors, 1)
	assert.Contains(t, result.Errors[0].Error(), expectedErr.Error())
}

func TestOneCfDeployTaskWithTwoApps(t *testing.T) {
	apps := []cfManifest.Application{
		{
			Name:   "app1",
			Routes: []string{"route"},
		},
		{
			Name:   "app2",
			Routes: []string{"route1"},
		},
	}

	linter := cfManifestLinter{readCfManifest: manifestReader(apps, nil)}
	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DeployCF{
				Manifest: "manifest.yml",
			},
		},
	}

	result := linter.Lint(man)
	assert.Len(t, result.Errors, 1)
	assertTooManyAppsError(t, "manifest.yml", result.Errors[0])
}

func TestTwoCfDeployTasksWithOneApp(t *testing.T) {
	apps := []cfManifest.Application{
		{
			Name:   "app",
			Routes: []string{"route"},
		},
	}

	linter := cfManifestLinter{readCfManifest: manifestReader(apps, nil)}

	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DeployCF{
				Manifest: "manifest.yml",
			},
			manifest.DeployCF{
				Manifest: "manifest.yml",
			},
		},
	}

	result := linter.Lint(man)
	assert.Len(t, result.Errors, 0)
}

func TestOneCfDeployTaskWithAnAppWithoutARoute(t *testing.T) {
	apps := []cfManifest.Application{
		{
			Name: "app",
		},
	}

	linter := cfManifestLinter{readCfManifest: manifestReader(apps, nil)}
	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DeployCF{
				Manifest: "manifest.yml",
			},
		},
	}

	result := linter.Lint(man)
	assert.Len(t, result.Errors, 1)
	assertNoRoutesError(t, "manifest.yml", result.Errors[0])
}

func TestOneCfDeployTaskWithAnAppWithoutAName(t *testing.T) {
	apps := []cfManifest.Application{
		{
			Routes: []string{"route"},
		},
	}
	linter := cfManifestLinter{readCfManifest: manifestReader(apps, nil)}
	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DeployCF{
				Manifest: "manifest.yml",
			},
		},
	}

	result := linter.Lint(man)
	assert.Len(t, result.Errors, 1)
	assertNoNameError(t, "manifest.yml", result.Errors[0])
}

func TestWorkerAppGivesErrorIfHealthCheckIsNotProcess(t *testing.T) {
	apps := []cfManifest.Application{
		{
			Name:    "My-app",
			NoRoute: true,
		},
	}

	linter := cfManifestLinter{readCfManifest: manifestReader(apps, nil)}
	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DeployCF{
				Manifest: "manifest.yml",
			},
		},
	}

	result := linter.Lint(man)
	assert.Len(t, result.Errors, 1)
	assertWrongHealthCheck(t, "manifest.yml", result.Errors[0])
}

func TestErrorsIfBothNoRouteAndRoutes(t *testing.T) {
	apps := []cfManifest.Application{
		{
			Name:    "My-app",
			NoRoute: true,
			Routes:  []string{"route1", "route2"},
		},
	}

	linter := cfManifestLinter{readCfManifest: manifestReader(apps, nil)}
	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DeployCF{
				Manifest: "manifest.yml",
			},
		},
	}

	result := linter.Lint(man)
	assert.Len(t, result.Errors, 1)
	assertBadRoutes(t, "manifest.yml", result.Errors[0])
}

func TestWorkerApp(t *testing.T) {
	apps := []cfManifest.Application{
		{
			Name:            "My-app",
			NoRoute:         true,
			HealthCheckType: "process",
		},
	}

	linter := cfManifestLinter{readCfManifest: manifestReader(apps, nil)}
	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DeployCF{
				Manifest: "manifest.yml",
			},
		},
	}

	result := linter.Lint(man)
	assert.Empty(t, result.Errors)
}

func TestDoesNotLintWhenManifestFromArtifacts(t *testing.T) {
	linter := cfManifestLinter{readCfManifest: manifestReader(nil, errors.New("asdf"))}

	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DeployCF{
				Manifest: "../artifacts/manifest.yml",
			},
		},
	}

	result := linter.Lint(man)
	assert.Len(t, result.Warnings, 0)
	assert.Len(t, result.Errors, 0)
}

func TestLintsBuildpackField(t *testing.T) {
	apps := []cfManifest.Application{
		{
			Name:      "My-app",
			Routes:    []string{"route1", "route2"},
			Buildpack: types.FilteredString{Value: "kehe"},
		},
	}

	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DeployCF{
			},
		},
	}

	linter := cfManifestLinter{readCfManifest: manifestReader(apps, nil)}

	result := linter.Lint(man)
	assert.Len(t, result.Warnings, 1)
	assert.Equal(t, errors2.NewDeprecatedBuildpackError(), result.Warnings[0])
	assert.Len(t, result.Errors, 0)
}
