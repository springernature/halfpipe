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

	errors, warnings := LintRunTask(manifest.Run{}, "", afero.Afero{})
	assert.Len(t, errors, 2)
	assert.Len(t, warnings, 0)

	helpers.AssertMissingField(t, " run.script", errors[0])
	helpers.AssertMissingField(t, " run.docker.image", errors[1])
}

func TestRunTaskWithScriptAndImage(t *testing.T) {
	task := manifest.Run{
		Script: "./build.sh",
		Docker: manifest.Docker{
			Image: "alpine",
		},
	}

	errors, warnings := LintRunTask(task, "", afero.Afero{Fs: afero.NewMemMapFs()})
	assert.Len(t, errors, 0)
	assert.Len(t, warnings, 1)
	helpers.AssertFileError(t, "./build.sh", warnings[0])
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

	errors, warnings := LintRunTask(task, "", fs)
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

	errors, warnings := LintRunTask(task, "", fs)
	assert.Len(t, errors, 1)
	assert.Len(t, warnings, 0)
	helpers.AssertMissingField(t, " run.docker.username", errors[0])
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

	errors, warnings := LintRunTask(task, "", fs)
	assert.Len(t, errors, 1)
	assert.Len(t, warnings, 0)
	helpers.AssertMissingField(t, " run.docker.password", errors[0])
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

	errors, warnings := LintRunTask(task, "", fs)
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

		errors, warnings := LintRunTask(task, "", fs)
		assert.Len(t, errors, 0)
		assert.Len(t, warnings, 0)
	}
}

func TestRetries(t *testing.T) {
	task := manifest.Run{}

	task.Retries = -1
	errors, _ := LintRunTask(task, "task", afero.Afero{Fs: afero.NewMemMapFs()})
	helpers.AssertInvalidFieldInErrors(t, "run.retries", errors)

	task.Retries = 6
	errors, _ = LintRunTask(task, "task", afero.Afero{Fs: afero.NewMemMapFs()})
	helpers.AssertInvalidFieldInErrors(t, "run.retries", errors)
}
