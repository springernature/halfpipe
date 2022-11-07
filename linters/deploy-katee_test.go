package linters

import (
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/manifest"
	"testing"
)

func TestKateeDeployTaskWithEmptyTask(t *testing.T) {
	task := manifest.DeployKatee{VelaManifest: "vela.yml"}
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	errors := LintDeployKateeTask(task, emptyManifest, fs)
	assertContainsError(t, errors, NewErrMissingField("application_name"))
	assertContainsError(t, errors, NewErrMissingField("image"))
	assertContainsError(t, errors, ErrFileNotFound)
}

func TestKateeDeployRetries(t *testing.T) {
	task := manifest.DeployKatee{}

	task.Retries = -1
	errors := LintDeployKateeTask(task, emptyManifest, afero.Afero{Fs: afero.NewMemMapFs()})
	assertContainsError(t, errors, ErrInvalidField.WithValue("retries"))

	task.Retries = 6
	errors = LintDeployKateeTask(task, emptyManifest, afero.Afero{Fs: afero.NewMemMapFs()})
	assertContainsError(t, errors, ErrInvalidField.WithValue("retries"))
}

func TestKateeDeployTag(t *testing.T) {
	task := manifest.DeployKatee{ApplicationName: "app", VelaManifest: "vela.yml", Image: "my-image"}
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	_ = fs.WriteFile("vela.yml", []byte("foo"), 0777)
	expectedTagError := ErrInvalidField.WithValue("tag")

	t.Run("not set", func(t *testing.T) {
		errors := LintDeployKateeTask(task, emptyManifest, fs)
		assertNotContainsError(t, errors, expectedTagError)
	})

	t.Run("gitref", func(t *testing.T) {
		task.Tag = "gitref"
		errors := LintDeployKateeTask(task, emptyManifest, fs)
		assertNotContainsError(t, errors, expectedTagError)
	})

	t.Run("version with update-pipeline feature", func(t *testing.T) {
		manifestWithUpdate := manifest.Manifest{FeatureToggles: manifest.FeatureToggles{manifest.FeatureUpdatePipeline}}
		task.Tag = "version"
		errors := LintDeployKateeTask(task, manifestWithUpdate, fs)
		assertNotContainsError(t, errors, expectedTagError)
	})

	t.Run("version without update-pipeline feature", func(t *testing.T) {
		manifestConcourse := manifest.Manifest{Platform: "concourse"}
		task.Tag = "version"
		errors := LintDeployKateeTask(task, manifestConcourse, fs)
		assertContainsError(t, errors, expectedTagError)

	})

	t.Run("version without update-pipeline feature", func(t *testing.T) {
		manifestActions := manifest.Manifest{Platform: "actions"}
		task.Tag = "version"
		errors := LintDeployKateeTask(task, manifestActions, fs)
		assertNotContainsError(t, errors, expectedTagError)
	})

	t.Run("invalid", func(t *testing.T) {
		task.Tag = "bananas"
		errors := LintDeployKateeTask(task, emptyManifest, fs)
		assertContainsError(t, errors, expectedTagError)
	})
}

func TestLintIfVelaFileDoesNotExist(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	task := manifest.DeployKatee{VelaManifest: "vela.yaml"}

	errors := LintDeployKateeTask(task, emptyManifest, fs)
	assertContainsError(t, errors, ErrFileNotFound.WithFile("vela.yaml"))
}

func TestLintReturnsErrorIfVelaFileExistsButInvalid(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("invalid-vela.yaml", []byte("blah"), 0777)
	task := manifest.DeployKatee{VelaManifest: "invalid-vela.yaml"}

	errors := LintDeployKateeTask(task, emptyManifest, fs)
	assertContainsError(t, errors, ErrFileInvalid.WithFile("invalid-vela.yaml"))
}

func TestLintReturnsErrorIfEnvInKateeIsNotSetInHalfpipe(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("vela.yaml",
		[]byte(`apiVersion: core.oam.dev/v1beta1
kind: Application
metadata:
name: ${KATEE_APPLICATION_NAME}
namespace: katee-${KATEE_TEAM}
spec:
 components:
 - name: ${KATEE_APPLICATION_NAME}
   type: snstateless
   properties:
     image: ${KATEE_APPLICATION_IMAGE}
     env:
       - name: BLAH
         value: ${BLAH}
`), 0777)

	task := manifest.DeployKatee{VelaManifest: "vela.yaml"}
	errors := LintDeployKateeTask(task, emptyManifest, fs)
	assertContainsError(t, errors, ErrVelaVariableMissing.WithValue("BLAH"))
}

func TestLintReturnsNoErrorIfEnvInKateeIsSetInHalfpipe(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("vela.yaml",
		[]byte(`apiVersion: core.oam.dev/v1beta1
kind: Application
metadata:
name: ${KATEE_APPLICATION_NAME}
namespace: katee-${KATEE_TEAM}
spec:
components:
- name: ${KATEE_APPLICATION_NAME}
  type: snstateless
  properties:
    image: ${KATEE_APPLICATION_IMAGE}
    env:
      - name: haha
        value: ${BLAH}
`), 0777)

	task := manifest.DeployKatee{
		VelaManifest: "vela.yaml",
		Vars: map[string]string{
			"BLAH": "Simon",
		},
	}

	errors := LintDeployKateeTask(task, emptyManifest, fs)
	assertNotContainsError(t, errors, ErrVelaVariableMissing.WithValue("BLAH"))
}

func TestLintReturnsNoErrorIfEnvVarsInKateeAreBuildVersionOrGitRef(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("vela.yaml",
		[]byte(`apiVersion: core.oam.dev/v1beta1
kind: Application
metadata:
name: ${KATEE_APPLICATION_NAME}
namespace: katee-${KATEE_TEAM}
spec:
components:
- name: ${KATEE_APPLICATION_NAME}
  type: snstateless
  properties:
    image: ${KATEE_APPLICATION_IMAGE}
    env:
      - name: haha
        value: ${BUILD_VERSION}
`), 0777)

	task := manifest.DeployKatee{
		VelaManifest: "vela.yaml",
		Vars: map[string]string{
			"BLAH": "Simon",
		},
	}

	errors := LintDeployKateeTask(task, emptyManifest, fs)
	assertNotContainsError(t, errors, ErrVelaVariableMissing.WithValue("BLAH"))
}
