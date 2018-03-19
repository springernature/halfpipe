package linters

import (
	"testing"

	manifest2 "code.cloudfoundry.org/cli/util/manifest"
	"github.com/pkg/errors"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestNoCfDeployTasks(t *testing.T) {
	man := manifest.Manifest{}

	linter := cfManifestLinter{
		rManifest: func(s string) ([]manifest2.Application, error) {
			return nil, errors.New("blah")
		},
	}

	result := linter.Lint(man)
	assert.Len(t, result.Errors, 0)
}

func TestOneCfDeployTask(t *testing.T) {
	readerGivesOneApp := func(name string) ([]manifest2.Application, error) {
		return []manifest2.Application{
			{
				Name:   name,
				Routes: []string{"route"},
			},
		}, nil
	}
	linter := cfManifestLinter{rManifest: readerGivesOneApp}

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
	readerGivesError := func(s string) ([]manifest2.Application, error) {
		return nil, errors.New("invalid manifest error")
	}
	linter := cfManifestLinter{rManifest: readerGivesError}

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
	readerGivesTwoApps := func(s string) ([]manifest2.Application, error) {
		return []manifest2.Application{
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
	linter := cfManifestLinter{rManifest: readerGivesTwoApps}
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
	readerGivesOneApp := func(s string) ([]manifest2.Application, error) {
		return []manifest2.Application{
			{
				Name:   s,
				Routes: []string{"route"},
			},
		}, nil
	}

	linter := cfManifestLinter{rManifest: readerGivesOneApp}

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
	readerGivesOneAppWithoutRoute := func(s string) ([]manifest2.Application, error) {
		return []manifest2.Application{
			{
				Name: s,
			},
		}, nil
	}
	linter := cfManifestLinter{rManifest: readerGivesOneAppWithoutRoute}
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
	readerGivesOneAppWithoutRoute := func(s string) ([]manifest2.Application, error) {
		return []manifest2.Application{
			{
				Routes: []string{"route"},
			},
		}, nil
	}
	linter := cfManifestLinter{rManifest: readerGivesOneAppWithoutRoute}
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