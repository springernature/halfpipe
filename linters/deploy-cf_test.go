package linters

import (
	"code.cloudfoundry.org/cli/util/manifestparser"
	"github.com/cloudfoundry/bosh-cli/director/template"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func validCfManifest() func(pathToManifest string, pathsToVarsFiles []string, vars []template.VarKV) (manifestparser.Manifest, error) {
	cfManifest := `
applications:
- name: test1
  routes:
  - route: test1.com
  buildpacks:
  - staticfile
`
	return cfManifestReader(cfManifest, nil)
}

func TestCFDeployTaskWithEmptyTask(t *testing.T) {
	task := manifest.DeployCF{Manifest: "manifest.yml"}
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	errs, _ := LintDeployCFTask(task, validCfManifest(), fs)
	AssertContainsError(t, errs, NewErrMissingField("api"))
	AssertContainsError(t, errs, NewErrMissingField("space"))
	AssertContainsError(t, errs, NewErrMissingField("org"))
	AssertContainsError(t, errs, ErrInvalidField.WithValue("cli_version"))
	AssertContainsError(t, errs, ErrFileNotFound)
}

func TestCFDeployTaskWithEmptyTestDomain(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("manifest.yml", []byte("foo"), 0777)

	task := manifest.DeployCF{
		API:        "someRandomApi",
		Org:        "Something",
		Space:      "Something",
		Manifest:   "manifest.yml",
		CliVersion: "cf6",
	}

	errors, _ := LintDeployCFTask(task, validCfManifest(), fs)
	AssertContainsError(t, errors, NewErrMissingField("test_domain"))
}

func TestCfCliVersion(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("manifest.yml", []byte("foo"), 0777)

	t.Run("not set", func(t *testing.T) {
		task := manifest.DeployCF{}

		errors, _ := LintDeployCFTask(task, validCfManifest(), fs)
		AssertContainsError(t, errors, ErrInvalidField.WithValue("cli_version"))
	})

	t.Run("valid", func(t *testing.T) {
		task := manifest.DeployCF{
			CliVersion: "cf6",
		}

		errors, _ := LintDeployCFTask(task, validCfManifest(), fs)
		AssertNotContainsError(t, errors, ErrInvalidField.WithValue("cli_version"))
	})
}

func TestCfPushRetries(t *testing.T) {
	task := manifest.DeployCF{}

	task.Retries = -1
	errors, _ := LintDeployCFTask(task, validCfManifest(), afero.Afero{Fs: afero.NewMemMapFs()})
	AssertContainsError(t, errors, ErrInvalidField.WithValue("retries"))

	task.Retries = 6
	errors, _ = LintDeployCFTask(task, validCfManifest(), afero.Afero{Fs: afero.NewMemMapFs()})
	AssertContainsError(t, errors, ErrInvalidField.WithValue("retries"))
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

	_, warnings := LintDeployCFTask(task, validCfManifest(), fs)

	AssertContainsError(t, warnings, ErrCFFromArtifact)
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

	errors, _ := LintDeployCFTask(task, validCfManifest(), fs)

	AssertContainsError(t, errors, ErrCFPrePromoteArtifact)
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
	errors, _ := LintDeployCFTask(task, validCfManifest(), fs)
	assert.Empty(t, errors)

	task.PreStart = []string{"cf something good", "something bad", "cf something else good", "something else bad"}
	errors, _ = LintDeployCFTask(task, validCfManifest(), fs)
	assert.Len(t, errors, 2)
	AssertContainsError(t, errors, ErrInvalidField.WithValue("pre_start"))
}

func TestSubTasksDoesntDefineNotifications(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("manifest.yml", []byte("foo"), 0777)

	task := manifest.DeployCF{
		API:        "asdf",
		Space:      "asdf",
		Org:        "asdf",
		Manifest:   "manifest.yml",
		TestDomain: "asdf",
		PrePromote: manifest.TaskList{
			manifest.Run{Notifications: manifest.Notifications{OnSuccess: []string{"Meehp"}}},
			manifest.Run{},
			manifest.Run{Notifications: manifest.Notifications{OnFailure: []string{"Moohp"}}},
		},
		CliVersion: "cf6",
	}

	errors, _ := LintDeployCFTask(task, validCfManifest(), fs)
	assert.Len(t, errors, 2)
	AssertContainsError(t, errors, ErrInvalidField.WithValue("pre_promote[0].notifications"))
	AssertContainsError(t, errors, ErrInvalidField.WithValue("pre_promote[2].notifications"))
}

func TestCFDeployTaskWithRollingAndPreStart(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("manifest.yml", []byte("foo"), 0777)

	task := manifest.DeployCF{
		API:        "api",
		Org:        "Something",
		Space:      "Something",
		Manifest:   "manifest.yml",
		TestDomain: "foo",
		Rolling:    true,
		CliVersion: "cf6",
		PreStart:   []string{"cf logs"},
	}

	errors, warnings := LintDeployCFTask(task, validCfManifest(), fs)
	if assert.Len(t, errors, 1) {
		assert.Equal(t, NewErrInvalidField("pre_start", "cannot use pre_start with rolling deployment"), errors[0])
	}
	assert.Len(t, warnings, 0)

}

func TestDockerTag(t *testing.T) {
	t.Run("Docker image is not specified in the manifest", func(t *testing.T) {
		cfManifestReader := cfManifestReader(`
applications:
- name: test1
  routes:
  - route: test1.com
  buildpacks:
  - staticfile
`, nil)

		fs := afero.Afero{Fs: afero.NewMemMapFs()}
		fs.WriteFile("manifest.yml", []byte("foo"), 0777)

		task := manifest.DeployCF{
			API:        "api",
			Org:        "Something",
			Space:      "Something",
			Manifest:   "manifest.yml",
			TestDomain: "foo",
			DockerTag:  "gitref",
		}

		errors, warnings := LintDeployCFTask(task, cfManifestReader, fs)
		assert.Len(t, warnings, 0)
		assert.Len(t, errors, 1)
		AssertContainsError(t, errors, ErrInvalidField.WithValue("docker_tag"))
	})

	t.Run("Docker image is specified in the manifest", func(t *testing.T) {
		cfManifestReader := cfManifestReader(`
applications:
- name: test1
  routes:
  - route: test1.com
  docker:
    image: eu.gcr.io/halfpipe-io/asd
  buildpacks:
  - staticfile
`, nil)
		fs := afero.Afero{Fs: afero.NewMemMapFs()}
		fs.WriteFile("manifest.yml", []byte("foo"), 0777)

		task := manifest.DeployCF{
			API:        "api",
			Org:        "Something",
			Space:      "Something",
			Manifest:   "manifest.yml",
			TestDomain: "foo",
			CliVersion: "cf6",
		}

		task.DockerTag = "gitref"
		errors, warnings := LintDeployCFTask(task, cfManifestReader, fs)
		assert.Len(t, warnings, 0)
		assert.Len(t, errors, 0)

		task.DockerTag = "version"
		errors, warnings = LintDeployCFTask(task, cfManifestReader, fs)
		assert.Len(t, warnings, 0)
		assert.Len(t, errors, 0)

		task.DockerTag = "unknown"
		errors, warnings = LintDeployCFTask(task, cfManifestReader, fs)
		assert.Len(t, warnings, 0)
		assert.Len(t, errors, 1)
		AssertContainsError(t, errors, ErrInvalidField.WithValue("docker_tag"))
	})
}

func TestCFDeployTaskSSORoute(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("manifest.yml", []byte("foo"), 0777)

	cfManifest := `
applications:
- name: test1
  routes:
  - route: test1.com
  - route: my-route.public.springernature.app
  buildpacks:
  - staticfile
`
	task := manifest.DeployCF{
		Manifest:   "manifest.yml",
		API:        "api",
		Space:      "space",
		Org:        "org",
		TestDomain: "foo",
		CliVersion: "cf6",
	}

	t.Run("valid", func(t *testing.T) {
		task.SSORoute = "my-route.public.springernature.app"
		errs, warns := LintDeployCFTask(task, cfManifestReader(cfManifest, nil), fs)
		assert.Empty(t, errs)
		assert.Empty(t, warns)
	})

	t.Run("invalid route", func(t *testing.T) {
		task.SSORoute = "my-route.springernature.app"
		errs, warns := LintDeployCFTask(task, cfManifestReader(cfManifest, nil), fs)
		AssertContainsError(t, errs, ErrInvalidField.WithValue("sso_route"))
		assert.Empty(t, warns)
	})

	t.Run("route not in cf manifest routes", func(t *testing.T) {
		task.SSORoute = "my-route.public.springernature.app"
		errs, _ := LintDeployCFTask(task, validCfManifest(), fs)
		AssertContainsError(t, errs, ErrCFRouteMissing)
	})

}
