package linters

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

var emptyManifest = manifest.Manifest{}

func TestDockerPushTaskWithEmptyTask(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	errors := LintDockerPushTask(manifest.DockerPush{}, emptyManifest, fs)
	assertContainsError(t, errors, NewErrMissingField("image"))
}

func TestDockerPushTaskWithBadRepo(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	task := manifest.DockerPush{
		Username: "asd",
		Password: "asd",
		Image:    "asd",
	}

	errors := LintDockerPushTask(task, emptyManifest, fs)
	assertContainsError(t, errors, ErrInvalidField.WithValue("image"))
}

func TestDockerPushTaskWithoutTeamDirectoryInHalfpipeRepo(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("Dockerfile", []byte("FROM ubuntu"), 0777)

	task := manifest.DockerPush{
		Username:       "asd",
		Password:       "asd",
		Image:          "eu.gcr.io/halfpipe-io/asd",
		DockerfilePath: "Dockerfile",
	}

	errs := LintDockerPushTask(task, emptyManifest, fs)
	assertContainsError(t, errs, ErrInvalidField.WithValue("image"))
}

func TestDockerPushTaskWithTeamDirectoryInHalfpipeRepo(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("Dockerfile", []byte("FROM ubuntu"), 0777)

	task := manifest.DockerPush{
		Username:       "asd",
		Password:       "asd",
		Image:          "eu.gcr.io/halfpipe-io/team/asd",
		DockerfilePath: "Dockerfile",
	}

	errors := LintDockerPushTask(task, emptyManifest, fs)
	assert.Len(t, errors, 0)
}

func TestDockerPushTaskWithoutTeamDirectoryInGCRRepo(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("Dockerfile", []byte("FROM ubuntu"), 0777)

	task := manifest.DockerPush{
		Username:       "asd",
		Password:       "asd",
		Image:          "eu.gcr.io/repo/asd",
		DockerfilePath: "Dockerfile",
	}

	errors := LintDockerPushTask(task, emptyManifest, fs)
	assert.Len(t, errors, 0)
}

func TestDockerPushTaskWhenDockerfileIsMissing(t *testing.T) {
	t.Run("When FilePath is missing", func(t *testing.T) {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}

		task := manifest.DockerPush{
			Username:       "asd",
			Password:       "asd",
			Image:          "user/image",
			DockerfilePath: "Dockerfile",
		}

		errors := LintDockerPushTask(task, emptyManifest, fs)
		assertContainsError(t, errors, ErrFileNotFound)
	})

	t.Run("don't error when RestoreArtifacts is true", func(t *testing.T) {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}

		task := manifest.DockerPush{
			Username:         "asd",
			Password:         "asd",
			Image:            "user/image",
			DockerfilePath:   "Dockerfile",
			RestoreArtifacts: true,
		}

		errors := LintDockerPushTask(task, emptyManifest, fs)
		assertNotContainsError(t, errors, ErrFileNotFound)
	})
}

func TestDockerPushTaskWithCorrectData(t *testing.T) {

	t.Run("When FilePath is just Dockerfile", func(t *testing.T) {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}
		fs.WriteFile("Dockerfile", []byte("FROM ubuntu"), 0777)

		task := manifest.DockerPush{
			Username: "asd",
			Password: "asd",
			Image:    "asd/asd",
			Vars: map[string]string{
				"A": "a",
				"B": "b",
			},
			DockerfilePath: "Dockerfile",
		}

		errors := LintDockerPushTask(task, emptyManifest, fs)
		assert.Len(t, errors, 0)
	})

	t.Run("When FilePath is in a different path", func(t *testing.T) {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}
		fs.WriteFile("dockerfile/Dockerfile", []byte("FROM ubuntu"), 0777)

		task := manifest.DockerPush{
			Username: "asd",
			Password: "asd",
			Image:    "asd/asd",
			Vars: map[string]string{
				"A": "a",
				"B": "b",
			},
			DockerfilePath: "dockerfile/Dockerfile",
		}

		errors := LintDockerPushTask(task, emptyManifest, fs)
		assert.Len(t, errors, 0)
	})

	t.Run("When FilePath is in a different folder upwards", func(t *testing.T) {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}
		fs.WriteFile("../dockerfile/Dockerfile", []byte("FROM ubuntu"), 0777)

		task := manifest.DockerPush{
			Username: "asd",
			Password: "asd",
			Image:    "asd/asd",
			Vars: map[string]string{
				"A": "a",
				"B": "b",
			},
			DockerfilePath: "../dockerfile/Dockerfile",
		}

		errors := LintDockerPushTask(task, emptyManifest, fs)
		assert.Len(t, errors, 0)
	})

}

func TestDockerPushWithBuildPath(t *testing.T) {
	t.Run("errors when build path doesnt exist", func(t *testing.T) {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}
		fs.WriteFile("Dockerfile", []byte("FROM ubuntu"), 0777)
		task := manifest.DockerPush{
			Username: "asd",
			Password: "asd",
			Image:    "asd/asd",
			Vars: map[string]string{
				"A": "a",
				"B": "b",
			},
			DockerfilePath: "Dockerfile",
			BuildPath:      "buildPathDoesntExist",
		}

		errors := LintDockerPushTask(task, emptyManifest, fs)
		assert.Len(t, errors, 1)
		assertContainsError(t, errors, ErrInvalidField.WithValue("build_path"))
	})

	t.Run("errors when build path is a file", func(t *testing.T) {
		buildPath := "imAFileNotADir"

		fs := afero.Afero{Fs: afero.NewMemMapFs()}
		fs.WriteFile("Dockerfile", []byte("FROM ubuntu"), 0777)
		fs.WriteFile(buildPath, []byte("Im a file, not a dir"), 0777)

		task := manifest.DockerPush{
			Username: "asd",
			Password: "asd",
			Image:    "asd/asd",
			Vars: map[string]string{
				"A": "a",
				"B": "b",
			},
			DockerfilePath: "Dockerfile",
			BuildPath:      buildPath,
		}

		errors := LintDockerPushTask(task, emptyManifest, fs)
		assert.Len(t, errors, 1)
		assertContainsError(t, errors, ErrInvalidField.WithValue("build_path"))
	})

	t.Run("ok when build path is a directory", func(t *testing.T) {
		buildPath := "imADir"

		fs := afero.Afero{Fs: afero.NewMemMapFs()}
		fs.WriteFile("Dockerfile", []byte("FROM ubuntu"), 0777)
		fs.Mkdir(buildPath, 0777)

		task := manifest.DockerPush{
			Username: "asd",
			Password: "asd",
			Image:    "asd/asd",
			Vars: map[string]string{
				"A": "a",
				"B": "b",
			},
			DockerfilePath: "Dockerfile",
			BuildPath:      buildPath,
		}

		errors := LintDockerPushTask(task, emptyManifest, fs)
		assert.Len(t, errors, 0)
	})

	t.Run("ok when build path is a directory upwards", func(t *testing.T) {
		buildPath := "../../imADir"

		fs := afero.Afero{Fs: afero.NewMemMapFs()}
		fs.WriteFile("Dockerfile", []byte("FROM ubuntu"), 0777)
		fs.Mkdir(buildPath, 0777)

		task := manifest.DockerPush{
			Username: "asd",
			Password: "asd",
			Image:    "asd/asd",
			Vars: map[string]string{
				"A": "a",
				"B": "b",
			},
			DockerfilePath: "Dockerfile",
			BuildPath:      buildPath,
		}

		errors := LintDockerPushTask(task, emptyManifest, fs)
		assert.Len(t, errors, 0)
	})

	t.Run("don't error when RestoreArtifacts is true", func(t *testing.T) {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}
		fs.WriteFile("Dockerfile", []byte("FROM ubuntu"), 0777)
		task := manifest.DockerPush{
			Username:         "asd",
			Password:         "asd",
			Image:            "asd/asd",
			DockerfilePath:   "Dockerfile",
			BuildPath:        "buildPathDoesntExist",
			RestoreArtifacts: true,
		}

		errors := LintDockerPushTask(task, emptyManifest, fs)
		assert.Empty(t, errors)
	})

}

func TestDockerPushRetries(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("Dockerfile", []byte("FROM ubuntu"), 0777)

	task := manifest.DockerPush{
		Username: "asd",
		Password: "asd",
		Image:    "asd/asd",
		Vars: map[string]string{
			"A": "a",
			"B": "b",
		},
		DockerfilePath: "Dockerfile",
	}

	task.Retries = -1
	errors := LintDockerPushTask(task, emptyManifest, fs)
	assertContainsError(t, errors, ErrInvalidField.WithValue("retries"))

	task.Retries = 6
	errors = LintDockerPushTask(task, emptyManifest, fs)
	assertContainsError(t, errors, ErrInvalidField.WithValue("retries"))

	task.Retries = 4
	errors = LintDockerPushTask(task, emptyManifest, fs)
	assert.Len(t, errors, 0)
}

func TestDockerPushTag(t *testing.T) {
	t.Run("Alles ok without tag", func(t *testing.T) {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}
		fs.WriteFile("Dockerfile", []byte("FROM ubuntu"), 0777)

		task := manifest.DockerPush{
			Image:          "asd/asd",
			Username:       "asd",
			Password:       "asdf",
			DockerfilePath: "Dockerfile",
		}

		errors := LintDockerPushTask(task, emptyManifest, fs)
		assert.Empty(t, errors)
	})

	t.Run("Should warn about unused field", func(t *testing.T) {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}
		fs.WriteFile("Dockerfile", []byte("FROM ubuntu"), 0777)

		task := manifest.DockerPush{
			Image:          "asd/asd",
			Username:       "asd",
			Password:       "asdf",
			DockerfilePath: "Dockerfile",
			Tag:            "yolo",
		}

		errs := LintDockerPushTask(task, emptyManifest, fs)
		assertContainsError(t, errs, ErrDockerPushTag)
	})
}

func TestMultiplePlatforms(t *testing.T) {
	t.Run("Alles ok when actions", func(t *testing.T) {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}
		fs.WriteFile("Dockerfile", []byte("FROM ubuntu"), 0777)

		task := manifest.DockerPush{
			Image:          "asd/asd",
			Username:       "asd",
			Password:       "asdf",
			DockerfilePath: "Dockerfile",
			Platforms:      []string{"linux/arm64", "linux/amd64"},
		}

		m := manifest.Manifest{Platform: "actions"}

		errors := LintDockerPushTask(task, m, fs)
		assert.Empty(t, errors)
	})

	t.Run("errors when unknown platform in docker push", func(t *testing.T) {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}
		fs.WriteFile("Dockerfile", []byte("FROM ubuntu"), 0777)

		task := manifest.DockerPush{
			Image:          "asd/asd",
			Username:       "asd",
			Password:       "asdf",
			DockerfilePath: "Dockerfile",
			Platforms:      []string{"linux/ad64"},
		}

		m := manifest.Manifest{Platform: "actions"}

		errors := LintDockerPushTask(task, m, fs)
		assertContainsError(t, errors, ErrDockerPlatformUnknown)
	})
}
