package tasks

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/linters/linterrors"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

var emptyManifest = manifest.Manifest{}

func TestDockerPushTaskWithEmptyTask(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	errors, _ := LintDockerPushTask(manifest.DockerPush{}, emptyManifest, fs)
	linterrors.AssertMissingFieldInErrors(t, "image", errors)
}

func TestDockerPushTaskWithBadRepo(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	task := manifest.DockerPush{
		Username: "asd",
		Password: "asd",
		Image:    "asd",
	}

	errors, _ := LintDockerPushTask(task, emptyManifest, fs)
	linterrors.AssertInvalidFieldInErrors(t, "image", errors)
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

	linterrors.AssertInvalidFieldInErrors(t, "image", warnings)
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
	t.Run("When FilePath is just Dockerfile", func(t *testing.T) {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}

		task := manifest.DockerPush{
			Username:       "asd",
			Password:       "asd",
			Image:          "user/image",
			DockerfilePath: "Dockerfile",
		}

		errors, _ := LintDockerPushTask(task, emptyManifest, fs)

		linterrors.AssertFileErrorInErrors(t, "Dockerfile", errors)
	})

	t.Run("When FilePath is in a different folder", func(t *testing.T) {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}

		task := manifest.DockerPush{
			Username:       "asd",
			Password:       "asd",
			Image:          "user/image",
			DockerfilePath: "dockerfiles/Dockerfile",
		}

		errors, _ := LintDockerPushTask(task, emptyManifest, fs)

		linterrors.AssertFileErrorInErrors(t, "dockerfiles/Dockerfile", errors)
	})

	t.Run("When FilePath is in a different folder upwards", func(t *testing.T) {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}

		task := manifest.DockerPush{
			Username:       "asd",
			Password:       "asd",
			Image:          "user/image",
			DockerfilePath: "../dockerfiles/Dockerfile",
		}

		errors, _ := LintDockerPushTask(task, emptyManifest, fs)

		linterrors.AssertFileErrorInErrors(t, "../dockerfiles/Dockerfile", errors)
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
		linterrors.AssertInvalidFieldInErrors(t, "build_path", errors)
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
		linterrors.AssertInvalidFieldInErrors(t, "build_path", errors)
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
	linterrors.AssertInvalidFieldInErrors(t, "retries", errors)

	task.Retries = 6
	errors, _ = LintDockerPushTask(task, emptyManifest, fs)
	linterrors.AssertInvalidFieldInErrors(t, "retries", errors)

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

	t.Run("Should warn about deprecated field", func(t *testing.T) {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}
		fs.WriteFile("Dockerfile", []byte("FROM ubuntu"), 0777)

		task := manifest.DockerPush{
			Image:          "asd/asd",
			Username:       "asd",
			Password:       "asdf",
			DockerfilePath: "Dockerfile",
			Tag:            "yolo",
		}

		errors, warnings := LintDockerPushTask(task, emptyManifest, fs)

		linterrors.AssertDeprecatedFieldErrorInErrors(t, "tag", warnings)
		assert.Empty(t, errors)
		assert.Len(t, warnings, 1)
	})
}
