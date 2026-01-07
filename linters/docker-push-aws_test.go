package linters

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/manifest"
)

func TestDockerPushAWSErrorsWhenPlatformIsNotActions(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("Dockerfile", []byte("FROM ubuntu"), 0777)

	task := manifest.DockerPushAWS{
		Repository:     "my-repo",
		DockerfilePath: "Dockerfile",
	}

	t.Run("errors when platform is concourse", func(t *testing.T) {
		man := manifest.Manifest{Platform: "concourse"}
		errs := LintDockerPushAWSTask(task, man, fs)
		assertContainsError(t, errs, ErrDockerPushAWSActionsOnly)
	})

	t.Run("no error when platform is actions", func(t *testing.T) {
		man := manifest.Manifest{Platform: "actions"}
		errs := LintDockerPushAWSTask(task, man, fs)
		assertNotContainsError(t, errs, ErrDockerPushAWSActionsOnly)
	})
}

func TestDockerPushAWSErrorsWhenRepositoryIsEmpty(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("Dockerfile", []byte("FROM ubuntu"), 0777)

	task := manifest.DockerPushAWS{
		Repository:     "",
		DockerfilePath: "Dockerfile",
	}
	man := manifest.Manifest{Platform: "actions"}

	errs := LintDockerPushAWSTask(task, man, fs)
	assertContainsError(t, errs, ErrMissingField.WithValue("repository"))
}

func TestDockerPushAWSAlwaysWarnsExperimental(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("Dockerfile", []byte("FROM ubuntu"), 0777)

	task := manifest.DockerPushAWS{
		Repository:     "my-repo",
		DockerfilePath: "Dockerfile",
	}
	man := manifest.Manifest{Platform: "actions"}

	errs := LintDockerPushAWSTask(task, man, fs)
	assertContainsError(t, errs, ErrDockerPushAWSExperimental)
}

func TestDockerPushAWSDockerfileExists(t *testing.T) {
	t.Run("errors when Dockerfile does not exist", func(t *testing.T) {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}

		task := manifest.DockerPushAWS{
			Repository:     "my-repo",
			DockerfilePath: "Dockerfile",
		}
		man := manifest.Manifest{Platform: "actions"}

		errs := LintDockerPushAWSTask(task, man, fs)
		assertContainsError(t, errs, ErrFileNotFound)
	})

	t.Run("no error when Dockerfile exists", func(t *testing.T) {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}
		fs.WriteFile("Dockerfile", []byte("FROM ubuntu"), 0777)

		task := manifest.DockerPushAWS{
			Repository:     "my-repo",
			DockerfilePath: "Dockerfile",
		}
		man := manifest.Manifest{Platform: "actions"}

		errs := LintDockerPushAWSTask(task, man, fs)
		assertNotContainsError(t, errs, ErrFileNotFound)
	})

	t.Run("errors when custom Dockerfile path does not exist", func(t *testing.T) {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}

		task := manifest.DockerPushAWS{
			Repository:     "my-repo",
			DockerfilePath: "docker/Dockerfile.prod",
		}
		man := manifest.Manifest{Platform: "actions"}

		errs := LintDockerPushAWSTask(task, man, fs)
		assertContainsError(t, errs, ErrFileNotFound)
	})

	t.Run("no error when custom Dockerfile path exists", func(t *testing.T) {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}
		fs.WriteFile("docker/Dockerfile.prod", []byte("FROM ubuntu"), 0777)

		task := manifest.DockerPushAWS{
			Repository:     "my-repo",
			DockerfilePath: "docker/Dockerfile.prod",
		}
		man := manifest.Manifest{Platform: "actions"}

		errs := LintDockerPushAWSTask(task, man, fs)
		assertNotContainsError(t, errs, ErrFileNotFound)
	})

	t.Run("no error when RestoreArtifacts is true and Dockerfile does not exist", func(t *testing.T) {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}

		task := manifest.DockerPushAWS{
			Repository:       "my-repo",
			DockerfilePath:   "Dockerfile",
			RestoreArtifacts: true,
		}
		man := manifest.Manifest{Platform: "actions"}

		errs := LintDockerPushAWSTask(task, man, fs)
		assertNotContainsError(t, errs, ErrFileNotFound)
	})
}
