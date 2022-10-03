package tasks

import (
	"code.cloudfoundry.org/cli/util/manifestparser"
	"errors"
	"fmt"
	"github.com/cloudfoundry/bosh-cli/director/template"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/linters/linterrors"
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

func TestInvalidManifest(t *testing.T) {
	task := manifest.DeployCF{}

	expectedErr := errors.New("invalid manifest error")

	errs, warns := LintCfManifest(task, cfManifestReader("", expectedErr))
	assert.Len(t, errs, 1)
	assert.Empty(t, warns)
	assert.Contains(t, errs[0].Error(), expectedErr.Error())
}

func TestTwoApps(t *testing.T) {
	cfManifest := `
applications:
- name: test1
  routes:
  - route: test1.com
- name: test2
  routes:
  - route: test2.com
`

	task := manifest.DeployCF{Manifest: "some-manifest.yml"}
	errs, _ := LintCfManifest(task, cfManifestReader(cfManifest, nil))

	assertTooManyAppsError(t, "some-manifest.yml", errs[0])
}

func TestWithoutARoute(t *testing.T) {
	cfManifest := `
applications:
- name: test
`

	errs, _ := LintCfManifest(manifest.DeployCF{Manifest: "some-manifest.yml"}, cfManifestReader(cfManifest, nil))
	assert.Len(t, errs, 1)
	assertNoRoutesError(t, "some-manifest.yml", errs[0])
}

func TestWithoutAName(t *testing.T) {
	cfManifest := `
applications:
- routes:
  - route: test.com
`
	errs, _ := LintCfManifest(manifest.DeployCF{Manifest: "some-manifest.yml"}, cfManifestReader(cfManifest, nil))
	assert.Len(t, errs, 1)
	assertNoNameError(t, "some-manifest.yml", errs[0])
}

func TestWorkerAppGivesErrorIfHealthCheckIsNotProcess(t *testing.T) {
	cfManifest := `
applications:
- name: test
  no-route: true
`

	errs, _ := LintCfManifest(manifest.DeployCF{Manifest: "some-manifest.yml"}, cfManifestReader(cfManifest, nil))
	assert.Len(t, errs, 1)
	assertWrongHealthCheck(t, "some-manifest.yml", errs[0])
}

func TestErrorsIfBothNoRouteAndRoutes(t *testing.T) {
	cfManifest := `
applications:
- name: test
  no-route: true
  routes:
  - route: test.com
`

	errs, _ := LintCfManifest(manifest.DeployCF{Manifest: "some-manifest.yml"}, cfManifestReader(cfManifest, nil))
	assert.Len(t, errs, 1)
	assertBadRoutes(t, "some-manifest.yml", errs[0])
}

func TestWorkerApp(t *testing.T) {
	cfManifest := `
applications:
- name: test
  no-route: true
  health-check-type: process
`

	errs, _ := LintCfManifest(manifest.DeployCF{}, cfManifestReader(cfManifest, nil))
	assert.Empty(t, errs)
}

func TestDoesNotLintWhenManifestFromArtifacts(t *testing.T) {
	task := manifest.DeployCF{Manifest: "../artifacts/manifest.yml"}
	errs, warns := LintCfManifest(task, cfManifestReader("", errors.New("manifest not found")))
	assert.Empty(t, errs)
	assert.Empty(t, warns)
}

func TestLintsBuildpackField(t *testing.T) {
	cfManifest := `
applications:
- name: test
  routes:
  - route: test.com
  buildpack: kehe
`

	_, warns := LintCfManifest(manifest.DeployCF{}, cfManifestReader(cfManifest, nil))
	assert.Equal(t, linterrors.NewDeprecatedBuildpackError(), warns[0])
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

	_, warns := LintCfManifest(manifest.DeployCF{}, cfManifestReader(cfManifest, nil))
	assert.Equal(t, linterrors.NewUnversionedBuildpackError("https://unversioned.com"), warns[0])
}

func TestLintMissingBuildpack(t *testing.T) {
	cfManifest := `
applications:
- name: test
  routes:
  - route: test.com
`
	_, warns := LintCfManifest(manifest.DeployCF{}, cfManifestReader(cfManifest, nil))
	assert.Equal(t, linterrors.NewMissingBuildpackError(), warns[0])
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

	errs, _ := LintCfManifest(manifest.DeployCF{}, cfManifestReader(cfManifest, nil))
	assert.Contains(t, errs[0].Error(), "http://test.com")
	assert.Contains(t, errs[1].Error(), "https://test.com")
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
		task := manifest.DeployCF{
			Manifest:       "manifest.yml",
			DeployArtifact: "somePath/file.jar",
		}

		errs, _ := LintCfManifest(task, cfManifestReader(cfManifest, nil))
		assert.Contains(t, errs[0].Error(), "you cannot specify both 'deploy_artifact' in the task")
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
		errs, _ := LintCfManifest(manifest.DeployCF{}, cfManifestReader(cfManifest, nil))
		assert.Contains(t, errs[0].Error(), "image must come from")
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

		errs, warns := LintCfManifest(manifest.DeployCF{}, cfManifestReader(cfManifest, nil))
		assert.Empty(t, errs)
		assert.Empty(t, warns)
	})

}

func assertTooManyAppsError(t *testing.T, name string, err error) {
	t.Helper()

	mf, ok := err.(linterrors.TooManyAppsError)
	if !ok {
		assert.Fail(t, "error is not an TooManyAppsError", err)
	} else {
		assert.Equal(t, name, mf.Path)
	}
}

func assertNoRoutesError(t *testing.T, name string, err error) {
	t.Helper()

	mf, ok := err.(linterrors.NoRoutesError)
	if !ok {
		assert.Fail(t, "error is not an TooManyAppsError", err)
	} else {
		assert.Equal(t, name, mf.Path)
	}
}

func assertNoNameError(t *testing.T, name string, err error) {
	t.Helper()

	mf, ok := err.(linterrors.NoNameError)
	if !ok {
		assert.Fail(t, "error is not an TooManyAppsError", err)
	} else {
		assert.Equal(t, name, mf.Path)
	}
}

func assertWrongHealthCheck(t *testing.T, name string, err error) {
	t.Helper()

	mf, ok := err.(linterrors.WrongHealthCheck)
	if !ok {
		assert.Fail(t, "error is not a WrongHealthCheck", err)
	} else {
		assert.Equal(t, name, mf.Path)
	}
}

func assertBadRoutes(t *testing.T, name string, err error) {
	t.Helper()

	mf, ok := err.(linterrors.BadRoutesError)
	if !ok {
		assert.Fail(t, "error is not an BadRoutesError", err)
	} else {
		assert.Equal(t, name, mf.Path)
	}
}
