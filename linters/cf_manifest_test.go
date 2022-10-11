package linters

import (
	"code.cloudfoundry.org/cli/util/manifestparser"
	"errors"
	"fmt"
	"github.com/cloudfoundry/bosh-cli/director/template"
	"github.com/springernature/halfpipe/config"
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
	expectedErr := errors.New("invalid manifest error")
	errs := LintCfManifest(manifest.DeployCF{}, cfManifestReader("", expectedErr))

	assert.Len(t, errs, 1)
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
	errs := LintCfManifest(task, cfManifestReader(cfManifest, nil))
	assertContainsError(t, errs, ErrCFMultipleApps)
}

func TestWithoutARoute(t *testing.T) {
	cfManifest := `
applications:
- name: test
`

	errs := LintCfManifest(manifest.DeployCF{Manifest: "some-manifest.yml"}, cfManifestReader(cfManifest, nil))
	assertContainsError(t, errs, ErrCFMissingRoutes)
}

func TestWithoutAName(t *testing.T) {
	cfManifest := `
applications:
- routes:
  - route: test.com
`
	errs := LintCfManifest(manifest.DeployCF{Manifest: "some-manifest.yml"}, cfManifestReader(cfManifest, nil))
	assertContainsError(t, errs, ErrCFMissingName)
}

func TestWorkerAppGivesErrorIfHealthCheckIsNotProcess(t *testing.T) {
	cfManifest := `
applications:
- name: test
  no-route: true
`

	errs := LintCfManifest(manifest.DeployCF{Manifest: "some-manifest.yml"}, cfManifestReader(cfManifest, nil))
	assertContainsError(t, errs, ErrCFNoRouteHealthcheck)
}

func TestErrorsIfBothNoRouteAndRoutes(t *testing.T) {
	cfManifest := `
applications:
- name: test
  no-route: true
  routes:
  - route: test.com
`

	errs := LintCfManifest(manifest.DeployCF{Manifest: "some-manifest.yml"}, cfManifestReader(cfManifest, nil))
	assertContainsError(t, errs, ErrCFRoutesAndNoRoute)
}

func TestWorkerAppDoesNotNeedRoute(t *testing.T) {
	cfManifest := `
applications:
- name: test
  no-route: true
  health-check-type: process
`

	errs := LintCfManifest(manifest.DeployCF{}, cfManifestReader(cfManifest, nil))
	assertNotContainsError(t, errs, ErrCFMissingRoutes)
}

func TestDoesNotLintWhenManifestFromArtifacts(t *testing.T) {
	task := manifest.DeployCF{Manifest: "../artifacts/manifest.yml"}
	errs := LintCfManifest(task, cfManifestReader("", errors.New("manifest not found")))
	assert.Empty(t, errs)
}

func TestLintsBuildpackField(t *testing.T) {
	cfManifest := `
applications:
- name: test
  routes:
  - route: test.com
  buildpack: kehe
`

	errs := LintCfManifest(manifest.DeployCF{}, cfManifestReader(cfManifest, nil))
	assertContainsError(t, errs, ErrCFBuildpackDeprecated)
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

	errs := LintCfManifest(manifest.DeployCF{}, cfManifestReader(cfManifest, nil))
	assertContainsError(t, errs, ErrCFBuildpackUnversioned)
}

func TestLintMissingBuildpack(t *testing.T) {
	cfManifest := `
applications:
- name: test
  routes:
  - route: test.com
`
	errs := LintCfManifest(manifest.DeployCF{}, cfManifestReader(cfManifest, nil))
	assertContainsError(t, errs, ErrCFBuildpackMissing)
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

	errs := LintCfManifest(manifest.DeployCF{}, cfManifestReader(cfManifest, nil))
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

		errs := LintCfManifest(task, cfManifestReader(cfManifest, nil))
		assertContainsError(t, errs, ErrCFArtifactAndDocker)
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
		errs := LintCfManifest(manifest.DeployCF{}, cfManifestReader(cfManifest, nil))
		assertContainsError(t, errs, ErrUnsupportedRegistry)
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

		errs := LintCfManifest(manifest.DeployCF{}, cfManifestReader(cfManifest, nil))
		assert.Empty(t, errs)
	})

}
