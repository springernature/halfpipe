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

	errs := LintDeployCFTask(task, validCfManifest(), fs)
	assertContainsError(t, errs, NewErrMissingField("api"))
	assertContainsError(t, errs, NewErrMissingField("space"))
	assertContainsError(t, errs, NewErrMissingField("org"))
	assertContainsError(t, errs, ErrInvalidField.WithValue("cli_version"))
	assertContainsError(t, errs, ErrFileNotFound)
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

	errors := LintDeployCFTask(task, validCfManifest(), fs)
	assertContainsError(t, errors, NewErrMissingField("test_domain"))
}

func TestCfCliVersion(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("manifest.yml", []byte("foo"), 0777)

	t.Run("not set", func(t *testing.T) {
		task := manifest.DeployCF{}

		errors := LintDeployCFTask(task, validCfManifest(), fs)
		assertContainsError(t, errors, ErrInvalidField.WithValue("cli_version"))
	})

	t.Run("valid", func(t *testing.T) {
		task := manifest.DeployCF{
			CliVersion: "cf6",
		}

		errors := LintDeployCFTask(task, validCfManifest(), fs)
		assertNotContainsError(t, errors, ErrInvalidField.WithValue("cli_version"))
	})
}

func TestCfPushRetries(t *testing.T) {
	task := manifest.DeployCF{}

	task.Retries = -1
	errors := LintDeployCFTask(task, validCfManifest(), afero.Afero{Fs: afero.NewMemMapFs()})
	assertContainsError(t, errors, ErrInvalidField.WithValue("retries"))

	task.Retries = 6
	errs := LintDeployCFTask(task, validCfManifest(), afero.Afero{Fs: afero.NewMemMapFs()})
	assertContainsError(t, errs, ErrInvalidField.WithValue("retries"))
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

	errs := LintDeployCFTask(task, validCfManifest(), fs)

	assertContainsError(t, errs, ErrCFFromArtifact)
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

	errors := LintDeployCFTask(task, validCfManifest(), fs)

	assertContainsError(t, errors, ErrCFPrePromoteArtifact)
}

func TestCfPushPreStart(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("manifest.yml", []byte("foo"), 0777)

	task := manifest.DeployCF{}

	task.PreStart = []string{"cf something good"}
	errs := LintDeployCFTask(task, validCfManifest(), fs)
	assertNotContainsError(t, errs, ErrInvalidField.WithValue("pre_start"))

	task.PreStart = []string{"cf something good", "something bad", "cf something else good", "something else bad"}
	errs = LintDeployCFTask(task, validCfManifest(), fs)
	assertContainsError(t, errs, ErrInvalidField.WithValue("pre_start"))
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
			manifest.Run{Notifications: manifest.Notifications{Slack: manifest.Slack{OnSuccess: []string{"Meehp"}}}},
			manifest.Run{},
			manifest.Run{Notifications: manifest.Notifications{Slack: manifest.Slack{OnFailure: []string{"Moohp"}}}},
		},
		CliVersion: "cf6",
	}

	errors := LintDeployCFTask(task, validCfManifest(), fs)
	assertContainsError(t, errors, ErrInvalidField.WithValue("pre_promote[0].notifications"))
	assertContainsError(t, errors, ErrInvalidField.WithValue("pre_promote[2].notifications"))
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

	errs := LintDeployCFTask(task, validCfManifest(), fs)
	assertContainsError(t, errs, ErrInvalidField.WithValue("pre_start"))
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

		errors := LintDeployCFTask(task, cfManifestReader, fs)
		assertContainsError(t, errors, ErrInvalidField.WithValue("docker_tag"))
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

		tagError := ErrInvalidField.WithValue("docker_tag")

		task.DockerTag = "gitref"
		errors := LintDeployCFTask(task, cfManifestReader, fs)
		assertNotContainsError(t, errors, tagError)

		task.DockerTag = "version"
		errors = LintDeployCFTask(task, cfManifestReader, fs)
		assertNotContainsError(t, errors, tagError)

		task.DockerTag = "unknown"
		errors = LintDeployCFTask(task, cfManifestReader, fs)
		assertContainsError(t, errors, tagError)
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
  metadata:
    labels:
      product: xyz
      environment: dev
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
		errs := LintDeployCFTask(task, cfManifestReader(cfManifest, nil), fs)
		assert.Empty(t, errs)
	})

	t.Run("invalid route", func(t *testing.T) {
		task.SSORoute = "my-route.springernature.app"
		errs := LintDeployCFTask(task, cfManifestReader(cfManifest, nil), fs)
		assertContainsError(t, errs, ErrInvalidField.WithValue("sso_route"))
	})

	t.Run("route not in cf manifest routes", func(t *testing.T) {
		task.SSORoute = "my-route.public.springernature.app"
		errs := LintDeployCFTask(task, validCfManifest(), fs)
		assertContainsError(t, errs, ErrCFRouteMissing)
	})

}
