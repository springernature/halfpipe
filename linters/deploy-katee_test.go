package linters

import (
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
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

	t.Run("not set", func(t *testing.T) {
		errors := LintDeployKateeTask(task, emptyManifest, fs)
		assert.Len(t, errors, 0)
	})

	t.Run("gitref", func(t *testing.T) {
		task.Tag = "gitref"
		errors := LintDeployKateeTask(task, emptyManifest, fs)
		assert.Len(t, errors, 0)
	})

	t.Run("version with update-pipeline feature", func(t *testing.T) {
		manifestWithUpdate := manifest.Manifest{FeatureToggles: manifest.FeatureToggles{manifest.FeatureUpdatePipeline}}
		task.Tag = "version"
		errors := LintDeployKateeTask(task, manifestWithUpdate, fs)
		assert.Len(t, errors, 0)
	})

	t.Run("version without update-pipeline feature", func(t *testing.T) {
		manifestConcourse := manifest.Manifest{Platform: "concourse"}
		task.Tag = "version"
		errors := LintDeployKateeTask(task, manifestConcourse, fs)
		assertContainsError(t, errors, ErrInvalidField.WithValue("tag"))

	})

	t.Run("version without update-pipeline feature", func(t *testing.T) {
		manifestActions := manifest.Manifest{Platform: "actions"}
		task.Tag = "version"
		errors := LintDeployKateeTask(task, manifestActions, fs)
		assert.Len(t, errors, 0)
	})

	t.Run("invalid", func(t *testing.T) {
		task.Tag = "bananas"
		errors := LintDeployKateeTask(task, emptyManifest, fs)
		assertContainsError(t, errors, ErrInvalidField.WithValue("tag"))
	})
}
