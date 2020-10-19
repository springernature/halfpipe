package tasks

import (
	cfManifest "code.cloudfoundry.org/cli/util/manifest"
	"github.com/cloudfoundry/bosh-cli/director/template"
	"testing"

	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/linters/linterrors"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func manifestReader(applications []cfManifest.Application, err error) func(pathToManifest string, pathsToVarsFiles []string, vars []template.VarKV) ([]cfManifest.Application, error) {
	return func(pathToManifest string, pathsToVarsFiles []string, vars []template.VarKV) ([]cfManifest.Application, error) {
		return applications, err
	}
}

func TestCFDeployTaskWithEmptyTask(t *testing.T) {
	task := manifest.DeployCF{Manifest: "manifest.yml"}
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	errors, warnings := LintDeployCFTask(task, manifest.Manifest{}, nil, fs, []string{})
	assert.Len(t, errors, 5)
	assert.Len(t, warnings, 0)

	linterrors.AssertMissingFieldInErrors(t, "api", errors)
	linterrors.AssertMissingFieldInErrors(t, "space", errors)
	linterrors.AssertMissingFieldInErrors(t, "org", errors)
	linterrors.AssertInvalidFieldInErrors(t, "cli_version", errors)
	linterrors.AssertFileErrorInErrors(t, "manifest.yml", errors)
}

func TestCFDeployTaskWithEmptyTestDomain(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("manifest.yml", []byte("foo"), 0777)

	task := manifest.DeployCF{
		API:        "((cloudfoundry.api-dev))",
		Org:        "Something",
		Space:      "Something",
		Manifest:   "manifest.yml",
		CliVersion: "cf6",
	}

	errors, warnings := LintDeployCFTask(task, manifest.Manifest{}, nil, fs, []string{})
	assert.Len(t, errors, 0)
	assert.Len(t, warnings, 0)

	task = manifest.DeployCF{
		API:        "",
		Org:        "Something",
		Space:      "Something",
		Manifest:   "manifest.yml",
		CliVersion: "cf6",
	}

	errors, warnings = LintDeployCFTask(task, manifest.Manifest{}, nil, fs, []string{})
	assert.Len(t, errors, 1)
	assert.Len(t, warnings, 0)
	linterrors.AssertMissingFieldInErrors(t, "api", errors)

	task = manifest.DeployCF{
		API:        "someRandomApi",
		Org:        "Something",
		Space:      "Something",
		Manifest:   "manifest.yml",
		CliVersion: "cf6",
	}

	errors, warnings = LintDeployCFTask(task, manifest.Manifest{}, nil, fs, []string{})
	assert.Len(t, errors, 1)
	assert.Len(t, warnings, 0)
	linterrors.AssertMissingFieldInErrors(t, "testDomain", errors)
}

func TestCfCliVersion(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("manifest.yml", []byte("foo"), 0777)

	t.Run("not set", func(t *testing.T) {
		task := manifest.DeployCF{
			API:      "((cloudfoundry.api-dev))",
			Org:      "Something",
			Space:    "Something",
			Manifest: "manifest.yml",
		}

		errors, warnings := LintDeployCFTask(task, manifest.Manifest{}, nil, fs, []string{})
		assert.Len(t, errors, 1)
		linterrors.AssertInvalidFieldInErrors(t, "cli_version", errors)
		assert.Len(t, warnings, 0)
	})

	t.Run("cf6", func(t *testing.T) {
		task := manifest.DeployCF{
			API:        "((cloudfoundry.api-dev))",
			Org:        "Something",
			Space:      "Something",
			Manifest:   "manifest.yml",
			CliVersion: "cf6",
		}

		errors, warnings := LintDeployCFTask(task, manifest.Manifest{}, nil, fs, []string{})
		assert.Len(t, errors, 0)
		assert.Len(t, warnings, 0)
	})

	t.Run("cf7", func(t *testing.T) {
		task := manifest.DeployCF{
			API:        "((cloudfoundry.api-dev))",
			Org:        "Something",
			Space:      "Something",
			Manifest:   "manifest.yml",
			CliVersion: "cf7",
		}

		errors, warnings := LintDeployCFTask(task, manifest.Manifest{}, nil, fs, []string{})
		assert.Len(t, errors, 0)
		assert.Len(t, warnings, 0)
	})
}

func TestCfPushRetries(t *testing.T) {
	task := manifest.DeployCF{}

	task.Retries = -1
	errors, _ := LintDeployCFTask(task, manifest.Manifest{}, nil, afero.Afero{Fs: afero.NewMemMapFs()}, []string{})
	linterrors.AssertInvalidFieldInErrors(t, "retries", errors)

	task.Retries = 6
	errors, _ = LintDeployCFTask(task, manifest.Manifest{}, nil, afero.Afero{Fs: afero.NewMemMapFs()}, []string{})
	linterrors.AssertInvalidFieldInErrors(t, "retries", errors)
}

func TestCFDeployTaskWithManifestFromArtifacts(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("manifest.yml", []byte("foo"), 0777)

	task := manifest.DeployCF{
		Manifest:   "../artifacts/manifest.yml",
		API:        "api",
		Space:      "space",
		Org:        "org",
		TestDomain: "test.domain",
		CliVersion: "cf6",
	}

	errors, warnings := LintDeployCFTask(task, manifest.Manifest{}, nil, fs, []string{})

	assert.Len(t, errors, 0)
	assert.Len(t, warnings, 1)
	linterrors.AssertFileErrorInErrors(t, "../artifacts/manifest.yml", warnings)
}

func TestCFDeployTaskWithManifestFromArtifactsAndPrePromoteShouldError(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("manifest.yml", []byte("foo"), 0777)

	task := manifest.DeployCF{
		Manifest:   "../artifacts/manifest.yml",
		API:        "api",
		Space:      "space",
		Org:        "org",
		TestDomain: "test.domain",
		CliVersion: "cf6",
		PrePromote: []manifest.Task{
			manifest.Run{},
		},
	}

	errors, warnings := LintDeployCFTask(task, manifest.Manifest{}, nil, fs, []string{})

	assert.Len(t, errors, 1)
	assert.Len(t, warnings, 1)
	linterrors.AssertInvalidFieldInErrors(t, "pre_promote", errors)
}

func TestCfPushPreStart(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("manifest.yml", []byte("foo"), 0777)

	task := manifest.DeployCF{
		Manifest:   "../artifacts/manifest.yml",
		API:        "api",
		Space:      "space",
		Org:        "org",
		TestDomain: "test.domain",
		CliVersion: "cf6",
	}

	task.PreStart = []string{"cf something good"}
	errors, _ := LintDeployCFTask(task, manifest.Manifest{}, nil, fs, []string{})
	assert.Empty(t, errors)

	task.PreStart = []string{"cf something good", "something bad", "cf something else good", "something else bad"}
	errors, _ = LintDeployCFTask(task, manifest.Manifest{}, nil, fs, []string{})
	assert.Len(t, errors, 2)
	linterrors.AssertInvalidFieldInErrors(t, "pre_start", errors)
}

func TestSubTasksDoesntDefineNotifications(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("manifest.yml", []byte("foo"), 0777)

	task := manifest.DeployCF{
		API:        "asdf",
		Space:      "asdf",
		Org:        "asdf",
		TestDomain: "asdf",
		PrePromote: manifest.TaskList{
			manifest.Run{Notifications: manifest.Notifications{OnSuccess: []string{"Meehp"}}},
			manifest.Run{},
			manifest.Run{Notifications: manifest.Notifications{OnFailure: []string{"Moohp"}}},
		},
		CliVersion: "cf6",
	}

	errors, _ := LintDeployCFTask(task, manifest.Manifest{}, nil, fs, []string{})
	assert.Len(t, errors, 2)
	linterrors.AssertInvalidFieldInErrors(t, "pre_promote[0].notifications", errors)
	linterrors.AssertInvalidFieldInErrors(t, "pre_promote[2].notifications", errors)
}

func TestCFDeployTaskWithDeprecatedCFApi(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("manifest.yml", []byte("foo"), 0777)

	task := manifest.DeployCF{
		API:        "deprecated.api",
		Org:        "Something",
		Space:      "Something",
		TestDomain: "foo",
		CliVersion: "cf6",
	}

	errors, warnings := LintDeployCFTask(task, manifest.Manifest{}, nil, fs, []string{"foo.bar", "deprecated.api"})
	assert.Len(t, errors, 0)
	if assert.Len(t, warnings, 1) {
		assert.Equal(t, linterrors.NewDeprecatedCFApiError("deprecated.api"), warnings[0])
	}
}

func TestCFDeployTaskWithRollingAndDeprecatedCFApi(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("manifest.yml", []byte("foo"), 0777)

	task := manifest.DeployCF{
		API:        "deprecated.api",
		Org:        "Something",
		Space:      "Something",
		TestDomain: "foo",
		Rolling:    true,
		CliVersion: "cf6",
	}

	errors, warnings := LintDeployCFTask(task, manifest.Manifest{}, nil, fs, []string{"foo.bar", "deprecated.api"})
	if assert.Len(t, errors, 1) {
		assert.Equal(t, linterrors.NewInvalidField("rolling", "cannot use rolling deployment with a deprecated api"), errors[0])
	}
	if assert.Len(t, warnings, 1) {
		assert.Equal(t, linterrors.NewDeprecatedCFApiError("deprecated.api"), warnings[0])
	}
}

func TestCFDeployTaskWithRollingAndPreStart(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("manifest.yml", []byte("foo"), 0777)

	task := manifest.DeployCF{
		API:        "api",
		Org:        "Something",
		Space:      "Something",
		TestDomain: "foo",
		Rolling:    true,
		CliVersion: "cf6",
		PreStart:   []string{"cf logs"},
	}

	errors, warnings := LintDeployCFTask(task, manifest.Manifest{}, nil, fs, []string{})
	if assert.Len(t, errors, 1) {
		assert.Equal(t, linterrors.NewInvalidField("pre_start", "cannot use pre_start with rolling deployment"), errors[0])
	}
	assert.Len(t, warnings, 0)

}

func TestDockerTag(t *testing.T) {
	t.Run("Docker image is not specified in the manifest", func(t *testing.T) {
		application := cfManifest.Application{Name: "kehe"}
		cfManifestReader := manifestReader([]cfManifest.Application{application}, nil)

		fs := afero.Afero{Fs: afero.NewMemMapFs()}
		fs.WriteFile("manifest.yml", []byte("foo"), 0777)

		task := manifest.DeployCF{
			API:        "api",
			Org:        "Something",
			Space:      "Something",
			TestDomain: "foo",
			DockerTag:  "gitref",
		}

		errors, warnings := LintDeployCFTask(task, manifest.Manifest{}, cfManifestReader, fs, []string{})
		assert.Len(t, warnings, 0)
		assert.Len(t, errors, 1)
		linterrors.AssertInvalidFieldInErrors(t, "docker_tag", errors)
	})

	t.Run("Docker image is specified in the manifest", func(t *testing.T) {
		application := cfManifest.Application{Name: "kehe", DockerImage: "asd"}
		cfManifestReader := manifestReader([]cfManifest.Application{application}, nil)

		t.Run("Unknown", func(t *testing.T) {
			fs := afero.Afero{Fs: afero.NewMemMapFs()}
			fs.WriteFile("manifest.yml", []byte("foo"), 0777)

			task := manifest.DeployCF{
				API:        "api",
				Org:        "Something",
				Space:      "Something",
				TestDomain: "foo",
				DockerTag:  "unknown",
				CliVersion: "cf6",
			}

			errors, warnings := LintDeployCFTask(task, manifest.Manifest{}, cfManifestReader, fs, []string{})
			assert.Len(t, warnings, 0)
			assert.Len(t, errors, 1)
			linterrors.AssertInvalidFieldInErrors(t, "docker_tag", errors)
		})

		t.Run("gitref", func(t *testing.T) {
			fs := afero.Afero{Fs: afero.NewMemMapFs()}
			fs.WriteFile("manifest.yml", []byte("foo"), 0777)

			task := manifest.DeployCF{
				API:        "api",
				Org:        "Something",
				Space:      "Something",
				TestDomain: "foo",
				DockerTag:  "gitref",
				CliVersion: "cf6",
			}

			errors, warnings := LintDeployCFTask(task, manifest.Manifest{}, cfManifestReader, fs, []string{})
			assert.Len(t, warnings, 0)
			assert.Len(t, errors, 0)
		})

		t.Run("version", func(t *testing.T) {
			t.Run("pipeline is not versioned", func(t *testing.T) {
				fs := afero.Afero{Fs: afero.NewMemMapFs()}
				fs.WriteFile("manifest.yml", []byte("foo"), 0777)

				task := manifest.DeployCF{
					API:        "api",
					Org:        "Something",
					Space:      "Something",
					TestDomain: "foo",
					DockerTag:  "version",
					CliVersion: "cf6",
				}

				errors, warnings := LintDeployCFTask(task, manifest.Manifest{}, cfManifestReader, fs, []string{})
				assert.Len(t, warnings, 0)
				assert.Len(t, errors, 1)
				linterrors.AssertInvalidFieldInErrors(t, "docker_tag", errors)
			})

			t.Run("pipeline is versioned", func(t *testing.T) {
				fs := afero.Afero{Fs: afero.NewMemMapFs()}
				fs.WriteFile("manifest.yml", []byte("foo"), 0777)

				task := manifest.DeployCF{
					API:        "api",
					Org:        "Something",
					Space:      "Something",
					TestDomain: "foo",
					DockerTag:  "version",
					CliVersion: "cf6",
				}

				errors, warnings := LintDeployCFTask(task, manifest.Manifest{FeatureToggles: manifest.FeatureToggles{manifest.FeatureUpdatePipeline}}, cfManifestReader, fs, []string{})
				assert.Len(t, warnings, 0)
				assert.Len(t, errors, 0)
			})
		})
	})
}
