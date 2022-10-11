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

	errors, _ := LintDockerPushTask(manifest.DockerPush{}, emptyManifest, fs)
	assertContainsError(t, errors, NewErrMissingField("image"))
}

func TestDockerPushTaskWithBadRepo(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	task := manifest.DockerPush{
		Username: "asd",
		Password: "asd",
		Image:    "asd",
	}

	errors, _ := LintDockerPushTask(task, emptyManifest, fs)
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

	errors, warnings := LintDockerPushTask(task, emptyManifest, fs)
	assert.Len(t, errors, 0)
	assert.Len(t, warnings, 1)

	assertContainsError(t, warnings, ErrInvalidField.WithValue("image"))
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

	errors, warnings := LintDockerPushTask(task, emptyManifest, fs)
	assert.Len(t, errors, 0)
	assert.Len(t, warnings, 0)
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

	errors, warnings := LintDockerPushTask(task, emptyManifest, fs)
	assert.Len(t, errors, 0)
	assert.Len(t, warnings, 0)
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

		errors, _ := LintDockerPushTask(task, emptyManifest, fs)

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

		errors, _ := LintDockerPushTask(task, emptyManifest, fs)
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

		errors, warnings := LintDockerPushTask(task, emptyManifest, fs)
		assert.Len(t, errors, 0)
		assert.Len(t, warnings, 0)
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

		errors, warnings := LintDockerPushTask(task, emptyManifest, fs)
		assert.Len(t, errors, 0)
		assert.Len(t, warnings, 0)
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

		errors, warnings := LintDockerPushTask(task, emptyManifest, fs)
		assert.Len(t, errors, 0)
		assert.Len(t, warnings, 0)
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

		errors, warnings := LintDockerPushTask(task, emptyManifest, fs)
		assert.Len(t, errors, 1)
		assert.Len(t, warnings, 0)
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

		errors, warnings := LintDockerPushTask(task, emptyManifest, fs)
		assert.Len(t, errors, 1)
		assert.Len(t, warnings, 0)
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

		errors, warnings := LintDockerPushTask(task, emptyManifest, fs)
		assert.Len(t, errors, 0)
		assert.Len(t, warnings, 0)
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

		errors, warnings := LintDockerPushTask(task, emptyManifest, fs)
		assert.Len(t, errors, 0)
		assert.Len(t, warnings, 0)
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

		errors, warnings := LintDockerPushTask(task, emptyManifest, fs)
		assert.Empty(t, errors)
		assert.Empty(t, warnings)
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
	errors, _ := LintDockerPushTask(task, emptyManifest, fs)
	assertContainsError(t, errors, ErrInvalidField.WithValue("retries"))

	task.Retries = 6
	errors, _ = LintDockerPushTask(task, emptyManifest, fs)
	assertContainsError(t, errors, ErrInvalidField.WithValue("retries"))

	task.Retries = 4
	errors, warnings := LintDockerPushTask(task, emptyManifest, fs)
	assert.Len(t, errors, 0)
	assert.Len(t, warnings, 0)
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

		errors, warnings := LintDockerPushTask(task, emptyManifest, fs)
		assert.Empty(t, errors)
		assert.Empty(t, warnings)
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

		_, warnings := LintDockerPushTask(task, emptyManifest, fs)
		assertContainsError(t, warnings, ErrDockerPushTag)
	})
}
