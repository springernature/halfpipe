package tasks

import (
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/helpers"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRunTaskWithoutScriptAndImage(t *testing.T) {
	man := manifest.Manifest{}
	man.Tasks = []manifest.Task{
		manifest.Run{},
	}

	errors, warnings := LintRunTask(manifest.Run{}, afero.Afero{})
	assert.Len(t, errors, 2)
	assert.Len(t, warnings, 0)

	helpers.AssertMissingField(t, "script", errors[0])
	helpers.AssertMissingField(t, "docker.image", errors[1])
}

func TestRunTaskWithScriptAndImageErrorsIfScriptIsNotThere(t *testing.T) {
	task := manifest.Run{
		Script: "./build.sh",
		Docker: manifest.Docker{
			Image: "alpine",
		},
	}

	errors, warnings := LintRunTask(task, afero.Afero{Fs: afero.NewMemMapFs()})
	assert.Len(t, errors, 1)
	assert.Len(t, warnings, 0)
	helpers.AssertFileErrorInErrors(t, "./build.sh", errors)
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

	errors, warnings := LintRunTask(task, fs)
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

	errors, warnings := LintRunTask(task, fs)
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

	errors, warnings := LintRunTask(task, fs)
	assert.Len(t, errors, 1)
	assert.Len(t, warnings, 0)
	helpers.AssertMissingField(t, "docker.username", errors[0])
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

	errors, warnings := LintRunTask(task, fs)
	assert.Len(t, errors, 1)
	assert.Len(t, warnings, 0)
	helpers.AssertMissingField(t, "docker.password", errors[0])
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

	errors, warnings := LintRunTask(task, fs)
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

		errors, warnings := LintRunTask(task, fs)
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

	errors, warnings := LintRunTask(task, fs)
	assert.Len(t, errors, 0)
	assert.Len(t, warnings, 1)
	assert.Contains(t, warnings, WarnScriptMustExistInDockerImage("make"))

	taskWithArgs := manifest.Run{
		Script: `\ls -al`,
		Docker: manifest.Docker{
			Image: "alpine",
		},
	}

	errors2, warnings2 := LintRunTask(taskWithArgs, fs)
	assert.Len(t, errors2, 0)
	assert.Len(t, warnings2, 1)
	assert.Contains(t, warnings2, WarnScriptMustExistInDockerImage("ls"))
}

func TestRetries(t *testing.T) {
	task := manifest.Run{}

	task.Retries = -1
	errors, _ := LintRunTask(task, afero.Afero{Fs: afero.NewMemMapFs()})
	helpers.AssertInvalidFieldInErrors(t, "retries", errors)

	task.Retries = 6
	errors, _ = LintRunTask(task, afero.Afero{Fs: afero.NewMemMapFs()})
	helpers.AssertInvalidFieldInErrors(t, "retries", errors)
}
