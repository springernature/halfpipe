package linters

import (
	"testing"

	cfManifest "code.cloudfoundry.org/cli/util/manifest"
	"github.com/pkg/errors"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestNoCfDeployTasks(t *testing.T) {
	man := manifest.Manifest{}

	linter := cfManifestLinter{
		readCfManifest: func(s string) ([]cfManifest.Application, error) {
			return nil, errors.New("blah")
		},
	}

	result := linter.Lint(man)
	assert.Len(t, result.Errors, 0)
}

func TestOneCfDeployTask(t *testing.T) {
	readerGivesOneApp := func(name string) ([]cfManifest.Application, error) {
		return []cfManifest.Application{
			{
				Name:   name,
				Routes: []string{"route"},
			},
		}, nil
	}
	linter := cfManifestLinter{readCfManifest: readerGivesOneApp}

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
	readerGivesError := func(s string) ([]cfManifest.Application, error) {
		return nil, errors.New("invalid manifest error")
	}
	linter := cfManifestLinter{readCfManifest: readerGivesError}

	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DeployCF{
				Manifest: "manifest.yml",
			},
		},
	}

	result := linter.Lint(man)
	assert.Len(t, result.Errors, 1)
	assert.Contains(t, result.Errors[0].Error(), "invalid manifest error")
}

func TestOneCfDeployTaskWithTwoApps(t *testing.T) {
	readerGivesTwoApps := func(s string) ([]cfManifest.Application, error) {
		return []cfManifest.Application{
			{
				Name:   s,
				Routes: []string{"route"},
			},
			{
				Name:   s + "1",
				Routes: []string{"route1"},
			},
		}, nil
	}
	linter := cfManifestLinter{readCfManifest: readerGivesTwoApps}
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
	readerGivesOneApp := func(s string) ([]cfManifest.Application, error) {
		return []cfManifest.Application{
			{
				Name:   s,
				Routes: []string{"route"},
			},
		}, nil
	}

	linter := cfManifestLinter{readCfManifest: readerGivesOneApp}

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
	readerGivesOneAppWithoutRoute := func(s string) ([]cfManifest.Application, error) {
		return []cfManifest.Application{
			{
				Name: s,
			},
		}, nil
	}
	linter := cfManifestLinter{readCfManifest: readerGivesOneAppWithoutRoute}
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
	readerGivesOneAppWithoutRoute := func(s string) ([]cfManifest.Application, error) {
		return []cfManifest.Application{
			{
				Routes: []string{"route"},
			},
		}, nil
	}
	linter := cfManifestLinter{readCfManifest: readerGivesOneAppWithoutRoute}
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
	reader := func(s string) ([]cfManifest.Application, error) {
		return []cfManifest.Application{
			{
				Name:    "My-app",
				NoRoute: true,
			},
		}, nil
	}

	linter := cfManifestLinter{readCfManifest: reader}
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
	reader := func(s string) ([]cfManifest.Application, error) {
		return []cfManifest.Application{
			{
				Name:    "My-app",
				NoRoute: true,
				Routes:  []string{"route1", "route2"},
			},
		}, nil
	}

	linter := cfManifestLinter{readCfManifest: reader}
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
	reader := func(s string) ([]cfManifest.Application, error) {
		return []cfManifest.Application{
			{
				Name:            "My-app",
				NoRoute:         true,
				HealthCheckType: "process",
			},
		}, nil
	}

	linter := cfManifestLinter{readCfManifest: reader}
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
