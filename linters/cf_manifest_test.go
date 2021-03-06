package linters

import (
	"code.cloudfoundry.org/cli/types"
	"github.com/springernature/halfpipe/config"
	errors2 "github.com/springernature/halfpipe/linters/linterrors"
	"path"
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
			manifest.DeployCF{},
		},
	}

	linter := cfManifestLinter{readCfManifest: manifestReader(apps, nil)}

	result := linter.Lint(man)
	assert.Len(t, result.Warnings, 1)
	assert.Equal(t, errors2.NewDeprecatedBuildpackError(), result.Warnings[0])
	assert.Len(t, result.Errors, 0)
}

func TestLintUnversionedBuildpack(t *testing.T) {
	buildpacks := []string{"https://versioned.com#v1.1", "https://unversioned.com", "system"}

	apps := []cfManifest.Application{{
		Name:       "My-app",
		Routes:     []string{"route1"},
		Buildpacks: buildpacks,
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
	apps := []cfManifest.Application{{
		Name:   "My-app",
		Routes: []string{"route1"},
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
	apps := []cfManifest.Application{
		{
			Name:       "My-app",
			Routes:     []string{"http://route1", "https://route2", "route1"},
			Buildpacks: []string{"java"},
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
	t.Run("Errors when both docker image and deploy artefact is specified", func(t *testing.T) {
		apps := []cfManifest.Application{
			{
				Name:        "appName",
				Routes:      []string{"route"},
				DockerImage: "nginx",
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

	t.Run("Errors when the images isn't from our repo", func(t *testing.T) {
		apps := []cfManifest.Application{
			{
				Name:        "appName",
				Routes:      []string{"route"},
				DockerImage: "nginx",
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
		apps := []cfManifest.Application{
			{
				Name:        "appName",
				Routes:      []string{"route"},
				DockerImage: path.Join(config.DockerRegistry, "nginx"),
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
