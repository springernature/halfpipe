package linters

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func testTaskLinter() taskLinter {
	return taskLinter{
		Fs: afero.Afero{Fs: afero.NewMemMapFs()},
	}
}

func TestAtLeastOneTaskExists(t *testing.T) {
	man := manifest.Manifest{}
	taskLinter := testTaskLinter()

	result := taskLinter.Lint(man)
	assert.Len(t, result.Errors, 1)
	assertMissingField(t, "tasks", result.Errors[0])
}

func TestRunTaskWithoutScriptAndImage(t *testing.T) {
	man := manifest.Manifest{}
	taskLinter := testTaskLinter()

	man.Tasks = []manifest.Task{
		manifest.Run{},
	}

	result := taskLinter.Lint(man)
	assert.Len(t, result.Errors, 2)
	assertMissingField(t, "tasks[0] run.script", result.Errors[0])
	assertMissingField(t, "tasks[0] run.docker.image", result.Errors[1])
}

func TestRunTaskWithScriptAndImage(t *testing.T) {
	taskLinter := testTaskLinter()
	man := manifest.Manifest{}
	man.Tasks = []manifest.Task{
		manifest.Run{
			Script: "./build.sh",
			Docker: manifest.Docker{
				Image: "alpine",
			},
		},
	}

	result := taskLinter.Lint(man)
	if assert.Len(t, result.Warnings, 1) {
		assertFileError(t, "./build.sh", result.Warnings[0])
	}
}

func TestRunTaskWithScriptAndImageWithPasswordAndUsername(t *testing.T) {
	taskLinter := testTaskLinter()
	taskLinter.Fs.WriteFile("build.sh", []byte("foo"), 0777)
	man := manifest.Manifest{}
	man.Tasks = []manifest.Task{
		manifest.Run{
			Script: "./build.sh",
			Docker: manifest.Docker{
				Image:    "alpine",
				Password: "secret",
				Username: "Michiel",
			},
		},
	}

	result := taskLinter.Lint(man)
	assert.Len(t, result.Errors, 0)
}

func TestRunTaskWithScriptAndImageAndOnlyPassword(t *testing.T) {
	taskLinter := testTaskLinter()
	taskLinter.Fs.WriteFile("build.sh", []byte("foo"), 0777)
	man := manifest.Manifest{}
	man.Tasks = []manifest.Task{
		manifest.Run{
			Script: "./build.sh",
			Docker: manifest.Docker{
				Image:    "alpine",
				Password: "secret",
			},
		},
	}

	result := taskLinter.Lint(man)
	assert.Len(t, result.Errors, 1)
	assertMissingField(t, "tasks[0] run.docker.username", result.Errors[0])
}
func TestRunTaskWithScriptAndImageAndOnlyUsername(t *testing.T) {
	taskLinter := testTaskLinter()
	taskLinter.Fs.WriteFile("build.sh", []byte("foo"), 0777)
	man := manifest.Manifest{}
	man.Tasks = []manifest.Task{
		manifest.Run{
			Script: "./build.sh",
			Docker: manifest.Docker{
				Image:    "alpine",
				Username: "Michiel",
			},
		},
	}

	result := taskLinter.Lint(man)
	assert.Len(t, result.Errors, 1)
	assertMissingField(t, "tasks[0] run.docker.password", result.Errors[0])
}

func TestRunTaskScriptFileExists(t *testing.T) {
	taskLinter := testTaskLinter()
	taskLinter.Fs.WriteFile("build.sh", []byte("foo"), 0777)

	man := manifest.Manifest{}
	man.Tasks = []manifest.Task{
		manifest.Run{
			Script: "./build.sh",
			Docker: manifest.Docker{
				Image: "alpine",
			},
		},
	}

	result := taskLinter.Lint(man)
	assert.Len(t, result.Errors, 0)
}

func TestRunTaskScriptAcceptsArguments(t *testing.T) {
	taskLinter := testTaskLinter()
	taskLinter.Fs.WriteFile("build.sh", []byte("foo"), 0777)

	for _, script := range []string{"./build.sh", "build.sh", "./build.sh --arg 1", "build.sh some args"} {
		man := manifest.Manifest{}
		man.Tasks = []manifest.Task{
			manifest.Run{
				Script: script,
				Docker: manifest.Docker{
					Image: "alpine",
				},
			},
		}

		result := taskLinter.Lint(man)
		assert.Len(t, result.Errors, 0)
	}
}

func TestCFDeployTaskWithEmptyTask(t *testing.T) {
	taskLinter := testTaskLinter()
	man := manifest.Manifest{}
	man.Tasks = []manifest.Task{
		manifest.DeployCF{Manifest: "manifest.yml"},
	}

	result := taskLinter.Lint(man)
	assert.Len(t, result.Errors, 4)
	assertMissingField(t, "tasks[0] deploy-cf.api", result.Errors[0])
	assertMissingField(t, "tasks[0] deploy-cf.space", result.Errors[1])
	assertMissingField(t, "tasks[0] deploy-cf.org", result.Errors[2])
	assertFileError(t, "manifest.yml", result.Errors[3])
}

func TestDockerPushTaskWithEmptyTask(t *testing.T) {
	taskLinter := testTaskLinter()
	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DockerPush{},
		},
	}

	result := taskLinter.Lint(man)
	assert.Len(t, result.Errors, 4)
	assertMissingField(t, "tasks[0] docker-push.username", result.Errors[0])
	assertMissingField(t, "tasks[0] docker-push.password", result.Errors[1])
	assertMissingField(t, "tasks[0] docker-push.image", result.Errors[2])
	assertFileError(t, "Dockerfile", result.Errors[3])

}

func TestDockerPushTaskWithBadRepo(t *testing.T) {
	taskLinter := testTaskLinter()
	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DockerPush{
				Username: "asd",
				Password: "asd",
				Image:    "asd",
			},
		},
	}

	result := taskLinter.Lint(man)
	assert.Len(t, result.Errors, 2)
	assertInvalidField(t, "tasks[0] docker-push.image", result.Errors[0])
	assertFileError(t, "Dockerfile", result.Errors[1])

}

func TestDockerPushTaskWhenDockerfileIsMissing(t *testing.T) {
	taskLinter := testTaskLinter()
	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DockerPush{
				Username: "asd",
				Password: "asd",
				Image:    "asd/asd",
			},
		},
	}

	result := taskLinter.Lint(man)
	assert.Len(t, result.Errors, 1)
	assertFileError(t, "Dockerfile", result.Errors[0])
}

func TestDockerPushTaskWithCorrectData(t *testing.T) {
	taskLinter := testTaskLinter()
	taskLinter.Fs.WriteFile("Dockerfile", []byte("FROM ubuntu"), 0777)

	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DockerPush{
				Username: "asd",
				Password: "asd",
				Image:    "asd/asd",
				Vars: map[string]string{
					"A": "a",
					"B": "b",
				},
			},
		},
	}

	result := taskLinter.Lint(man)
	assert.Len(t, result.Errors, 0)
}

func TestEnvVarsMustBeUpperCase(t *testing.T) {
	taskLinter := testTaskLinter()

	badKey1 := "KeHe"
	badKey2 := "b"
	badKey3 := "AAAAa"

	goodKey1 := "YO"
	goodKey2 := "A"
	goodKey3 := "AOIJASOID"

	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.Run{
				Vars: map[string]string{
					badKey1:  "a",
					goodKey1: "sup",
				},
			},

			manifest.DockerPush{
				Vars: map[string]string{
					goodKey2: "a",
					badKey2:  "B",
				},
			},

			manifest.DeployCF{
				Vars: map[string]string{
					badKey3:  "asd",
					goodKey3: "asd",
				},
			},
		},
	}

	result := taskLinter.Lint(man)
	assertInvalidFieldInErrors(t, badKey1, result.Errors)
	assertInvalidFieldInErrors(t, badKey2, result.Errors)
	assertInvalidFieldInErrors(t, badKey3, result.Errors)

	assertInvalidFieldShouldNotBeInErrors(t, goodKey1, result.Errors)
	assertInvalidFieldShouldNotBeInErrors(t, goodKey2, result.Errors)
	assertInvalidFieldShouldNotBeInErrors(t, goodKey3, result.Errors)
}

func TestDockerCompose_Happy(t *testing.T) {
	taskLinter := testTaskLinter()
	taskLinter.Fs.WriteFile("docker-compose.yml", []byte("something"), 0777)
	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DockerCompose{}, // empty is ok, everything is optional
			manifest.DockerCompose{
				Name: "run docker compose",
				Vars: manifest.Vars{
					"A": "a",
					"B": "b",
				},
			},
		},
	}

	result := taskLinter.Lint(man)
	assert.Len(t, result.Errors, 0)
}

func TestDockerCompose_MissingFile(t *testing.T) {
	taskLinter := testTaskLinter()
	man := manifest.Manifest{
		Tasks: []manifest.Task{manifest.DockerCompose{}},
	}

	result := taskLinter.Lint(man)
	assert.Len(t, result.Errors, 1)
	assertFileError(t, "docker-compose.yml", result.Errors[0])
}

func TestDockerCompose_InvalidVar(t *testing.T) {
	taskLinter := testTaskLinter()
	taskLinter.Fs.WriteFile("docker-compose.yml", []byte("something"), 0777)
	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DockerCompose{
				Name: "run docker compose",
				Vars: manifest.Vars{
					"a": "a",
					"B": "b",
				},
			},
		},
	}

	result := taskLinter.Lint(man)
	assert.Len(t, result.Errors, 1)
	assertInvalidFieldInErrors(t, "tasks[0] a", result.Errors)
}

func TestLintsSubTasksInDeployCF(t *testing.T) {

	man := manifest.Manifest{}
	taskLinter := testTaskLinter()

	man.Tasks = []manifest.Task{
		manifest.DeployCF{
			API:   "api",
			Org:   "org",
			Space: "space",
			PrePromote: []manifest.Task{
				manifest.Run{},
				manifest.DockerCompose{},
			},
		},
	}

	result := taskLinter.Lint(man)
	assert.Len(t, result.Errors, 3)
	assertMissingField(t, "tasks[0].pre_promote[0] run.script", result.Errors[0])
	assertMissingField(t, "tasks[0].pre_promote[0] run.docker.image", result.Errors[1])
	assertFileError(t, "docker-compose.yml", result.Errors[2])
}
