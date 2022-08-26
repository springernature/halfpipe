package linters

import (
	"code.cloudfoundry.org/cli/util/manifestparser"
	"errors"
	"fmt"
	"github.com/cloudfoundry/bosh-cli/director/template"
	"github.com/springernature/halfpipe/config"
	errors2 "github.com/springernature/halfpipe/linters/linterrors"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"testing"
)

func cfManifestReader(manifestYaml string, err error) func(pathToManifest string, pathsToVarsFiles []string, vars []template.VarKV) (manifestparser.Manifest, error) {
	return func(pathToManifest string, pathsToVarsFiles []string, vars []template.VarKV) (manifestparser.Manifest, error) {
		var parsedManifest manifestparser.Manifest
		if parseErr := yaml.Unmarshal([]byte(manifestYaml), &parsedManifest); parseErr != nil {
			err = parseErr
		}
		return parsedManifest, err
	}
}

func TestNoCfDeployTasks(t *testing.T) {
	man := manifest.Manifest{}
	linter := cfManifestLinter{readCfManifest: cfManifestReader("", nil)}

	result := linter.Lint(man)
	assert.Len(t, result.Errors, 0)
}

func TestOneCfDeployTask(t *testing.T) {
	cfManifest := `
applications:
- name: test
  routes:
  - route: test.com
`
	linter := cfManifestLinter{readCfManifest: cfManifestReader(cfManifest, nil)}

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
	linter := cfManifestLinter{readCfManifest: cfManifestReader("", expectedErr)}

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
	cfManifest := `
applications:
- name: test1
  routes:
  - route: test1.com
- name: test2
  routes:
  - route: test2.com
`

	linter := cfManifestLinter{readCfManifest: cfManifestReader(cfManifest, nil)}
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
	cfManifest := `
applications:
- name: test
`

	linter := cfManifestLinter{readCfManifest: cfManifestReader(cfManifest, nil)}
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
	cfManifest := `
applications:
- routes:
  - route: test.com
`
	linter := cfManifestLinter{readCfManifest: cfManifestReader(cfManifest, nil)}
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
	cfManifest := `
applications:
- name: test
  no-route: true
`

	linter := cfManifestLinter{readCfManifest: cfManifestReader(cfManifest, nil)}
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
	cfManifest := `
applications:
- name: test
  no-route: true
  routes:
  - route: test.com
`

	linter := cfManifestLinter{readCfManifest: cfManifestReader(cfManifest, nil)}
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
	cfManifest := `
applications:
- name: test
  no-route: true
  health-check-type: process
`

	linter := cfManifestLinter{readCfManifest: cfManifestReader(cfManifest, nil)}
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
	linter := cfManifestLinter{readCfManifest: cfManifestReader("", errors.New("asdf"))}

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
	cfManifest := `
applications:
- name: test
  routes:
  - route: test.com
  buildpack: kehe
`

	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DeployCF{},
		},
	}

	linter := cfManifestLinter{readCfManifest: cfManifestReader(cfManifest, nil)}

	result := linter.Lint(man)
	assert.Equal(t, errors2.NewDeprecatedBuildpackError(), result.Warnings[0])
	assert.Len(t, result.Errors, 0)
}

func TestLintUnversionedBuildpack(t *testing.T) {

	cfManifest := `
applications:
- name: test
  routes:
  - route: test.com
  buildpacks:
  - "https://versioned.com#v1.1"
  - "https://unversioned.com"
  - system
`

	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DeployCF{},
		},
	}

	linter := cfManifestLinter{readCfManifest: cfManifestReader(cfManifest, nil)}

	result := linter.Lint(man)
	assert.Len(t, result.Errors, 0)

	if assert.Len(t, result.Warnings, 1) {
		assert.Equal(t, errors2.NewUnversionedBuildpackError("https://unversioned.com"), result.Warnings[0])
	}

}

func TestLintMissingBuildpack(t *testing.T) {
	cfManifest := `
applications:
- name: test
  routes:
  - route: test.com
`
	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DeployCF{},
		},
	}

	linter := cfManifestLinter{readCfManifest: cfManifestReader(cfManifest, nil)}

	result := linter.Lint(man)
	assert.Len(t, result.Errors, 0)
	if assert.Len(t, result.Warnings, 1) {
		assert.Equal(t, errors2.NewMissingBuildpackError(), result.Warnings[0])
	}

}

func TestLintNoHttpInRoutes(t *testing.T) {
	cfManifest := `
applications:
- name: test
  routes:
  - route: http://test.com
  - route: https://test.com
  - route: route1
  buildpacks:
  - java
`

	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DeployCF{},
		},
	}

	linter := cfManifestLinter{readCfManifest: cfManifestReader(cfManifest, nil)}

	result := linter.Lint(man)
	assert.Len(t, result.Warnings, 0)
	assert.Len(t, result.Errors, 2)
}

func TestLintDockerImagePush(t *testing.T) {
	t.Run("Errors when both docker image and deploy artifact is specified", func(t *testing.T) {
		cfManifest := `
applications:
- name: test
  docker:
    image: blah
  routes:
  - route: test.com
`
		linter := cfManifestLinter{readCfManifest: cfManifestReader(cfManifest, nil)}

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
		cfManifest := `
applications:
- name: test
  docker:
    image: blah
  routes:
  - route: test.com
`
		linter := cfManifestLinter{readCfManifest: cfManifestReader(cfManifest, nil)}

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
		cfManifest := fmt.Sprintf(`
applications:
- name: test
  docker:
    image: %s/blah
  routes:
  - route: test.com
`, config.DockerRegistry)

		linter := cfManifestLinter{readCfManifest: cfManifestReader(cfManifest, nil)}

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
