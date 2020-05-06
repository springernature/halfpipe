package linters

import (
	"fmt"
	"github.com/spf13/afero"
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
		linter := NewDeprecatedDockerRegistriesLinter(afero.Afero{}, deprecatedPrefixes, deprecationDate, oneMonthBeforeDeprecation)
		assert.Empty(t, linter.Lint(man).Errors)
		assert.Empty(t, linter.Lint(man).Warnings)
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

		lintResult := NewDeprecatedDockerRegistriesLinter(afero.Afero{}, deprecatedPrefixes, deprecationDate, oneMonthBeforeDeprecation).Lint(man)
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

		lintResult := NewDeprecatedDockerRegistriesLinter(afero.Afero{}, deprecatedPrefixes, deprecationDate, oneWeekFromDeprecation).Lint(man)
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
		lintResult := NewDeprecatedDockerRegistriesLinter(afero.Afero{}, deprecatedPrefixes, deprecationDate, oneWeekFromDeprecation).Lint(man)
		assert.Len(t, lintResult.Warnings, 2)
		assert.Empty(t, lintResult.Errors)
	})
}

func TestDockerCompose(t *testing.T) {
	t.Run("When we fail to read the docker-compose file", func(t *testing.T) {
		man := manifest.Manifest{
			Tasks: manifest.TaskList{
				manifest.DockerCompose{ComposeFile: "some/path"},
			},
		}

		linter := NewDeprecatedDockerRegistriesLinter(afero.Afero{afero.NewMemMapFs()}, deprecatedPrefixes, deprecationDate, oneMonthBeforeDeprecation)
		assert.Len(t, linter.Lint(man).Errors, 1)
		assert.Len(t, linter.Lint(man).Warnings, 0)
		assert.Equal(t, "open some/path: file does not exist", linter.Lint(man).Errors[0].Error())
	})

	t.Run("When we are referencing a deprecated repo", func(t *testing.T) {
		man := manifest.Manifest{
			Tasks: manifest.TaskList{
				manifest.DockerCompose{ComposeFile: "bad/path1"},
				manifest.DockerCompose{ComposeFile: "bad/path2"},
				manifest.DockerCompose{ComposeFile: "ok/path"},
			},
		}

		fs := afero.Afero{afero.NewMemMapFs()}
		fs.WriteFile("bad/path1", []byte(fmt.Sprintf(`
code:
  image: %s
  dns: *DNS
  volumes:
    - .:/home/dev/code
    - $HOME/.ivy2:/home/dev/.ivy2
    - $HOME/.sbt:/home/dev/.sbt

deploy:
  image: %s
  dns: *DNS
  volumes:
    - .:/home/dev/code
`, deprecatedImage1, okImage1)), 0777)
		fs.WriteFile("bad/path2", []byte(fmt.Sprintf(`
code:
  image: %s
  dns: *DNS
  volumes:
    - .:/home/dev/code
    - $HOME/.ivy2:/home/dev/.ivy2
    - $HOME/.sbt:/home/dev/.sbt

deploy:
  image: %s
  dns: *DNS
  volumes:
    - .:/home/dev/code
`, okImage1, deprecatedImage2)), 0777)
		fs.WriteFile("ok/path", []byte(fmt.Sprintf(`
code:
  image: %s
  dns: *DNS
  volumes:
    - .:/home/dev/code
    - $HOME/.ivy2:/home/dev/.ivy2
    - $HOME/.sbt:/home/dev/.sbt

deploy:
  image: %s
  dns: *DNS
  volumes:
    - .:/home/dev/code
`, okImage1, okImage2)), 0777)
		linter := NewDeprecatedDockerRegistriesLinter(fs, deprecatedPrefixes, deprecationDate, oneMonthBeforeDeprecation)
		result := linter.Lint(man)
		assert.False(t, result.HasErrors())
		assert.Len(t, result.Warnings, 2)
		assert.Contains(t, result.Warnings[0].Error(), deprecatedPrefixes[0])
		assert.Contains(t, result.Warnings[1].Error(), deprecatedPrefixes[1])
	})
}

func TestDockerPush(t *testing.T) {
	t.Run("When we are pushing to deprecated registries", func(t *testing.T) {
		man := manifest.Manifest{
			Tasks: manifest.TaskList{
				manifest.DockerPush{Image: deprecatedImage1},
				manifest.DockerPush{Image: deprecatedImage2},
				manifest.DockerPush{Image: okImage1},
			},
		}

		fs := afero.Afero{afero.NewMemMapFs()}
		linter := NewDeprecatedDockerRegistriesLinter(fs, deprecatedPrefixes, deprecationDate, oneMonthBeforeDeprecation)
		result := linter.Lint(man)
		assert.False(t, result.HasErrors())
		assert.Len(t, result.Warnings, 2)
		assert.Contains(t, result.Warnings[0].Error(), deprecatedImage1)
		assert.Contains(t, result.Warnings[1].Error(), deprecatedImage2)
	})

	t.Run("When we fail to read the Dockerfile", func(t *testing.T) {
		man := manifest.Manifest{
			Tasks: manifest.TaskList{
				manifest.DockerPush{DockerfilePath: "some/path"},
			},
		}

		fs := afero.Afero{afero.NewMemMapFs()}
		linter := NewDeprecatedDockerRegistriesLinter(fs, deprecatedPrefixes, deprecationDate, oneMonthBeforeDeprecation)
		result := linter.Lint(man)
		assert.Len(t, result.Warnings, 0)
		assert.Len(t, result.Errors, 1)
		assert.Contains(t, result.Errors[0].Error(), "some/path")

	})

	t.Run("When we are referencing a deprecated registry in the Dockerfile", func(t *testing.T) {
		man := manifest.Manifest{
			Tasks: manifest.TaskList{
				manifest.DockerPush{DockerfilePath: "bad/path1"},
				manifest.DockerPush{DockerfilePath: "bad/path2"},
				manifest.DockerPush{DockerfilePath: "ok/path"},
			},
		}

		fs := afero.Afero{afero.NewMemMapFs()}
		fs.WriteFile("bad/path1", []byte(fmt.Sprintf(`
FROM %s
RUN ls
`, deprecatedImage1)), 0777)
		fs.WriteFile("bad/path2", []byte(fmt.Sprintf(`
FROM %s
RUN rm -rf /
`, deprecatedImage2)), 0777)
		fs.WriteFile("ok/path", []byte(fmt.Sprintf(`
FROM %s
RUN echo yay
`, okImage1)), 0777)
		linter := NewDeprecatedDockerRegistriesLinter(fs, deprecatedPrefixes, deprecationDate, oneMonthBeforeDeprecation)
		result := linter.Lint(man)
		assert.False(t, result.HasErrors())
		assert.Len(t, result.Warnings, 2)
		assert.Contains(t, result.Warnings[0].Error(), deprecatedPrefixes[0])
		assert.Contains(t, result.Warnings[1].Error(), deprecatedPrefixes[1])
	})

}
