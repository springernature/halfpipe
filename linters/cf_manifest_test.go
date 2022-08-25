package linters

import (
	"code.cloudfoundry.org/cli/util/manifestparser"
	"github.com/springernature/halfpipe/config"
	errors2 "github.com/springernature/halfpipe/linters/linterrors"
	"path"
	"testing"

	"errors"
	"github.com/cloudfoundry/bosh-cli/director/template"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func manifestReader(apps []manifestparser.Application, err error) func(pathToManifest string, pathsToVarsFiles []string, vars []template.VarKV) (manifestparser.Manifest, error) {
	return func(pathToManifest string, pathsToVarsFiles []string, vars []template.VarKV) (manifestparser.Manifest, error) {
		return manifestparser.Manifest{Applications: apps}, err
	}
}

func TestNoCfDeployTasks(t *testing.T) {
	man := manifest.Manifest{}

	linter := cfManifestLinter{
		readCfManifest: manifestReader([]manifestparser.Application{{}}, nil),
	}

	result := linter.Lint(man)
	assert.Len(t, result.Errors, 0)
}

func TestOneCfDeployTask(t *testing.T) {
	app := manifestparser.Application{
		Name:            "appName",
		NoRoute:         true,
		HealthCheckType: "process",
	}
	linter := cfManifestLinter{readCfManifest: manifestReader([]manifestparser.Application{app}, nil)}

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
	linter := cfManifestLinter{readCfManifest: manifestReader([]manifestparser.Application{{}}, expectedErr)}

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
	apps := []manifestparser.Application{
		{
			Name:    "app1",
			NoRoute: true,
		},
		{
			Name:    "app2",
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
	assertTooManyAppsError(t, "manifest.yml", result.Errors[0])
}

func TestOneCfDeployTaskWithAnAppWithoutARoute(t *testing.T) {
	apps := []manifestparser.Application{
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
	apps := []manifestparser.Application{
		{
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
	assert.Len(t, result.Errors, 1)
	assertNoNameError(t, "manifest.yml", result.Errors[0])
}

func TestWorkerAppGivesErrorIfHealthCheckIsNotProcess(t *testing.T) {
	apps := []manifestparser.Application{
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
	apps := []manifestparser.Application{
		{
			Name:    "My-app",
			NoRoute: true,
			RemainingManifestFields: map[string]interface{}{
				"routes": []map[string]string{
					{"route": "route1"},
					{"route": "route2"},
				},
			},
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
	apps := []manifestparser.Application{
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
	apps := []manifestparser.Application{
		{
			Name:            "My-app",
			NoRoute:         true,
			HealthCheckType: "process",
			RemainingManifestFields: map[string]interface{}{
				"buildpack": "kehe",
			},
		},
	}

	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DeployCF{},
		},
	}

	linter := cfManifestLinter{readCfManifest: manifestReader(apps, nil)}

	result := linter.Lint(man)
	assert.Equal(t, errors2.NewDeprecatedBuildpackError(), result.Warnings[0])
	assert.Len(t, result.Errors, 0)
}

func TestLintUnversionedBuildpack(t *testing.T) {
	apps := []manifestparser.Application{{
		Name:            "My-app",
		NoRoute:         true,
		HealthCheckType: "process",
		RemainingManifestFields: map[string]any{
			"buildpacks": []any{"https://versioned.com#v1.1", "https://unversioned.com", "system"},
		},
	}}

	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DeployCF{},
		},
	}

	linter := cfManifestLinter{readCfManifest: manifestReader(apps, nil)}

	result := linter.Lint(man)
	assert.Len(t, result.Errors, 0)

	if assert.Len(t, result.Warnings, 1) {
		assert.Equal(t, errors2.NewUnversionedBuildpackError("https://unversioned.com"), result.Warnings[0])
	}

}

func TestLintMissingBuildpack(t *testing.T) {
	apps := []manifestparser.Application{{
		Name:            "My-app",
		NoRoute:         true,
		HealthCheckType: "process",
	}}

	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DeployCF{},
		},
	}

	linter := cfManifestLinter{readCfManifest: manifestReader(apps, nil)}

	result := linter.Lint(man)
	assert.Len(t, result.Errors, 0)
	if assert.Len(t, result.Warnings, 1) {
		assert.Equal(t, errors2.NewMissingBuildpackError(), result.Warnings[0])
	}

}

func TestLintNoHttpInRoutes(t *testing.T) {
	apps := []manifestparser.Application{
		{
			Name: "My-app",
			RemainingManifestFields: map[string]any{
				"buildpacks": []any{"https://versioned.com#v1.1"},
				"routes": []any{
					map[any]any{"route": "http://route1"},
					map[any]any{"route": "https://route2"},
					map[any]any{"route": "route1"},
				},
			},
		},
	}

	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DeployCF{},
		},
	}

	linter := cfManifestLinter{readCfManifest: manifestReader(apps, nil)}

	result := linter.Lint(man)
	assert.Len(t, result.Warnings, 0)
	assert.Len(t, result.Errors, 2)
}

func TestLintDockerImagePush(t *testing.T) {
	t.Run("Errors when both docker image and deploy artifact is specified", func(t *testing.T) {
		apps := []manifestparser.Application{
			{
				Name:            "appName",
				NoRoute:         true,
				HealthCheckType: "process",
				Docker: &manifestparser.Docker{
					Image: "nginx",
				},
			},
		}
		linter := cfManifestLinter{readCfManifest: manifestReader(apps, nil)}

		man := manifest.Manifest{
			Tasks: []manifest.Task{
				manifest.DeployCF{
					Manifest:       "manifest.yml",
					DeployArtifact: "somePath/file.jar",
				},
			},
		}

		result := linter.Lint(man)
		assert.Len(t, result.Errors, 1)
		assert.Contains(t, result.Errors[0].Error(), "you cannot specify both 'deploy_artifact' in the task")

	})

	t.Run("Errors when the image isn't from our repo", func(t *testing.T) {
		apps := []manifestparser.Application{
			{
				Name:            "appName",
				NoRoute:         true,
				HealthCheckType: "process",
				Docker: &manifestparser.Docker{
					Image: "nginx",
				},
			},
		}
		linter := cfManifestLinter{readCfManifest: manifestReader(apps, nil)}

		man := manifest.Manifest{
			Tasks: []manifest.Task{
				manifest.DeployCF{
					Manifest: "manifest.yml",
					API:      "((cloudfoundry.api-snpaas))",
				},
			},
		}

		result := linter.Lint(man)
		assert.Len(t, result.Errors, 1)
		assert.Contains(t, result.Errors[0].Error(), "image must come from")
	})

	t.Run("All is good", func(t *testing.T) {
		apps := []manifestparser.Application{
			{
				Name:            "appName",
				NoRoute:         true,
				HealthCheckType: "process",
				Docker: &manifestparser.Docker{
					Image: path.Join(config.DockerRegistry, "nginx"),
				},
			},
		}
		linter := cfManifestLinter{readCfManifest: manifestReader(apps, nil)}

		man := manifest.Manifest{
			Tasks: []manifest.Task{
				manifest.DeployCF{
					Manifest: "manifest.yml",
					API:      "((cloudfoundry.api-snpaas))",
				},
			},
		}

		result := linter.Lint(man)
		assert.Len(t, result.Errors, 0)
	})

}
