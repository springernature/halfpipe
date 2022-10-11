package linters

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestRunTaskWithoutScriptAndImage(t *testing.T) {
	man := manifest.Manifest{}
	man.Tasks = []manifest.Task{
		manifest.Run{},
	}

	errors, _ := LintRunTask(manifest.Run{}, afero.Afero{}, "")
	assertContainsError(t, errors, NewErrMissingField("script"))
	assertContainsError(t, errors, NewErrMissingField("docker.image"))
}

func TestRunTaskWithScriptAndImageErrorsIfScriptIsNotThere(t *testing.T) {
	task := manifest.Run{
		Script: "./build.sh",
		Docker: manifest.Docker{
			Image: "alpine",
		},
	}

	errors, _ := LintRunTask(task, afero.Afero{Fs: afero.NewMemMapFs()}, "")
	assert.Len(t, errors, 1)
	assertContainsError(t, errors, ErrFileNotFound)
}

func TestRunTaskWithScriptAndImageWithPasswordAndUsername(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("build.sh", []byte("foo"), 0777)

	task := manifest.Run{
		Script: "./build.sh",
		Docker: manifest.Docker{
			Image:    "alpine",
			Password: "secret",
			Username: "Michiel",
		},
	}

	errors, warnings := LintRunTask(task, fs, "")
	assert.Len(t, errors, 0)
	assert.Len(t, warnings, 0)
}

func TestRunTaskWithScriptWithoutDotSlashAndImageWithPasswordAndUsername(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("build.sh", []byte("foo"), 0777)

	task := manifest.Run{
		Script: "build.sh",
		Docker: manifest.Docker{
			Image:    "alpine",
			Password: "secret",
			Username: "Michiel",
		},
	}

	errors, warnings := LintRunTask(task, fs, "")
	assert.Len(t, errors, 0)
	assert.Len(t, warnings, 0)
}

func TestRunTaskWithScriptAndImageAndOnlyPassword(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("build.sh", []byte("foo"), 0777)

	task := manifest.Run{
		Script: "./build.sh",
		Docker: manifest.Docker{
			Image:    "alpine",
			Password: "secret",
		},
	}

	errors, _ := LintRunTask(task, fs, "")
	assertContainsError(t, errors, NewErrMissingField("docker.username"))
}

func TestRunTaskWithScriptAndImageAndOnlyUsername(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("build.sh", []byte("foo"), 0777)

	task := manifest.Run{
		Script: "./build.sh",
		Docker: manifest.Docker{
			Image:    "alpine",
			Username: "username",
		},
	}

	errors, _ := LintRunTask(task, fs, "")
	assertContainsError(t, errors, NewErrMissingField("docker.password"))
}

func TestRunTaskScriptFileExists(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("build.sh", []byte("foo"), 0777)

	task := manifest.Run{
		Script: "./build.sh",
		Docker: manifest.Docker{
			Image: "alpine",
		},
	}

	errors, warnings := LintRunTask(task, fs, "")
	assert.Len(t, errors, 0)
	assert.Len(t, warnings, 0)
}

func TestRunTaskScriptAcceptsArguments(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("build.sh", []byte("foo"), 0777)

	for _, script := range []string{"./build.sh", "build.sh", "./build.sh --arg 1", "build.sh some args"} {
		task := manifest.Run{
			Script: script,
			Docker: manifest.Docker{
				Image: "alpine",
			},
		}

		errors, warnings := LintRunTask(task, fs, "")
		assert.Len(t, errors, 0)
		assert.Len(t, warnings, 0)
	}
}

func TestRunTaskWithScriptThatStartsWithBackSlackShouldNotError(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	task := manifest.Run{
		Script: `\make`,
		Docker: manifest.Docker{
			Image: "alpine",
		},
	}

	errors, warnings := LintRunTask(task, fs, "")
	assert.Len(t, errors, 0)
	assert.Len(t, warnings, 1)
	assert.Contains(t, warnings, WarnScriptMustExistInDockerImage("make"))

	taskWithArgs := manifest.Run{
		Script: `\ls -al`,
		Docker: manifest.Docker{
			Image: "alpine",
		},
	}

	errors2, warnings2 := LintRunTask(taskWithArgs, fs, "")
	assert.Len(t, errors2, 0)
	assert.Len(t, warnings2, 1)
	assert.Contains(t, warnings2, WarnScriptMustExistInDockerImage("ls"))
}

func TestRetries(t *testing.T) {
	task := manifest.Run{}

	task.Retries = -1
	errors, _ := LintRunTask(task, afero.Afero{Fs: afero.NewMemMapFs()}, "")
	assertContainsError(t, errors, ErrInvalidField.WithValue("retries"))

	task.Retries = 6
	errors, _ = LintRunTask(task, afero.Afero{Fs: afero.NewMemMapFs()}, "")
	assertContainsError(t, errors, ErrInvalidField.WithValue("retries"))
}

func TestShouldSkipExecutableTestAndProduceWarningIfRunningOnWindows(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("build.sh", []byte("foo"), 0444)

	task := manifest.Run{
		Script: "build.sh",
		Docker: manifest.Docker{
			Image:    "alpine",
			Password: "secret",
			Username: "Michiel",
		},
	}

	errors, warnings := LintRunTask(task, fs, "windows")
	assert.Len(t, errors, 0)
	assert.Len(t, warnings, 1)
	assert.Equal(t, WarnMakeSureScriptIsExecutable("build.sh"), warnings[0])
}
