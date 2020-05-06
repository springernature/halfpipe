package linters

import (
	"fmt"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"path"
	"testing"
	"time"
)

var deprecatedPrefixes = []string{
	"my-private-repo.tools",
	"docker-registry",
}

var oneMonthBeforeDeprecation = time.Now().AddDate(0, 0, -1)
var deprecationDate = time.Now().AddDate(0, 1, 0)
var oneWeekFromDeprecation = time.Now().AddDate(0, 0, 21)

var deprecatedImage1 = path.Join(deprecatedPrefixes[0], "something")
var deprecatedImage2 = path.Join(deprecatedPrefixes[1], "something")
var okImage1 = "something"
var okImage2 = "eu.gcr.io/halfpipe-io/something"

func TestRunTask(t *testing.T) {
	t.Run("No error/warning", func(t *testing.T) {
		man := manifest.Manifest{
			Tasks: manifest.TaskList{
				manifest.Run{
					Docker: manifest.Docker{
						Image: okImage1,
					},
				},
				manifest.Run{
					Docker: manifest.Docker{
						Image: okImage2,
					},
				},
			},
		}
		result := NewDeprecatedDockerRegistriesLinter(deprecatedPrefixes, deprecationDate, oneMonthBeforeDeprecation)
		assert.Empty(t, result.Lint(man).Errors)
		assert.Empty(t, result.Lint(man).Warnings)
	})

	t.Run("Warning", func(t *testing.T) {
		man := manifest.Manifest{
			Tasks: manifest.TaskList{
				manifest.Run{
					Docker: manifest.Docker{
						Image: okImage2,
					},
				},
				manifest.Run{
					Docker: manifest.Docker{
						Image: deprecatedImage1,
					},
				},
				manifest.Run{
					Docker: manifest.Docker{
						Image: deprecatedImage2,
					},
				},
			},
		}

		lintResult := NewDeprecatedDockerRegistriesLinter(deprecatedPrefixes, deprecationDate, oneMonthBeforeDeprecation).Lint(man)
		assert.Empty(t, lintResult.Errors)
		assert.Len(t, lintResult.Warnings, 2)
	})

	t.Run("Error", func(t *testing.T) {
		man := manifest.Manifest{
			Tasks: manifest.TaskList{
				manifest.Run{
					Docker: manifest.Docker{
						Image: okImage2,
					},
				},
				manifest.Run{
					Docker: manifest.Docker{
						Image: deprecatedImage1,
					},
				},
				manifest.Run{
					Docker: manifest.Docker{
						Image: deprecatedImage2,
					},
				},
			},
		}

		lintResult := NewDeprecatedDockerRegistriesLinter(deprecatedPrefixes, deprecationDate, oneWeekFromDeprecation).Lint(man)
		assert.Empty(t, lintResult.Warnings)
		assert.Len(t, lintResult.Errors, 2)

		fmt.Println(lintResult.Errors[1])
	})

	t.Run("Should error, but doesnt due to feature toggle", func(t *testing.T) {
		man := manifest.Manifest{
			Tasks: manifest.TaskList{
				manifest.Run{
					Docker: manifest.Docker{
						Image: okImage2,
					},
				},
				manifest.Run{
					Docker: manifest.Docker{
						Image: deprecatedImage1,
					},
				},
				manifest.Run{
					Docker: manifest.Docker{
						Image: deprecatedImage2,
					},
				},
			},
		}
		man.FeatureToggles = []string{
			manifest.FeatureToggleDisableDeprecatedDockerRegistryError,
		}
		lintResult := NewDeprecatedDockerRegistriesLinter(deprecatedPrefixes, deprecationDate, oneWeekFromDeprecation).Lint(man)
		assert.Len(t, lintResult.Warnings, 2)
		assert.Empty(t, lintResult.Errors)
	})
}

func TestDockerCompose(t *testing.T) {
	t.Run("When we fail to read the docker-compose file", func(t *testing.T) {
		panic("implement me")
	})

	t.Run("When we are referencing a deprecated repo", func(t *testing.T) {
		panic("implement me")
	})

	t.Run("When we are not referencing a deprecated repo", func(t *testing.T) {
		panic("implement me")
	})
}

func DockerPush(t testing.T) {
	t.Run("When we fail to read the Dockerfile", func(t *testing.T) {
		panic("implement me")
	})

	t.Run("When we are pushing to deprecated registries", func(t *testing.T) {
		panic("implement me")
	})

	t.Run("When we are referencing a deprecated registry in the Dockerfile", func(t *testing.T) {
		panic("implement me")
	})

	t.Run("When we are not referencing a deprecated repo", func(t *testing.T) {
		panic("implement me")
	})
}
