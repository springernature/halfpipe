package linters

import (
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCurrentTaskDoesNotRequireArtifactAndThereAreNoPreviousTasks(t *testing.T) {
	currentTask := manifest.DockerPush{}
	errors, warnings := LintArtifacts(currentTask, nil)
	assert.Len(t, errors, 0)
	assert.Len(t, warnings, 0)
}

func TestCurrentTaskDoesNotRequireArtifactAndThereArePreviousTasks(t *testing.T) {
	currentTask := manifest.DockerPush{}
	errors, warnings := LintArtifacts(currentTask, []manifest.Task{manifest.DockerCompose{}, manifest.Run{}})
	assert.Len(t, errors, 0)
	assert.Len(t, warnings, 0)
}

func TestCurrentTaskRequiresArtifactAndThereArePreviousTasks(t *testing.T) {
	currentTask := manifest.DockerPush{RestoreArtifacts: true}
	errors, _ := LintArtifacts(currentTask, nil)
	assert.Len(t, errors, 1)

	assert.Equal(t, "reads from saved artifacts, but there are no previous tasks that saves any", errors[0].Error())
}

func TestCurrentTaskRequiresArtifactAndThereArePreviousTasksThatDoesntSaveAny(t *testing.T) {
	currentTask := manifest.DockerPush{RestoreArtifacts: true}
	errors, _ := LintArtifacts(currentTask, []manifest.Task{manifest.DockerCompose{}, manifest.Run{}})
	assert.Len(t, errors, 1)

	assert.Equal(t, "reads from saved artifacts, but there are no previous tasks that saves any", errors[0].Error())
}

func TestCurrentTaskRequiresArtifactAndThereIsAPreviousTasksThatSavesOne(t *testing.T) {
	currentTask := manifest.DockerPush{RestoreArtifacts: true}
	errors, _ := LintArtifacts(currentTask, []manifest.Task{manifest.DockerCompose{SaveArtifacts: []string{"path/to/artifact/to/save"}}, manifest.Run{}})
	assert.Len(t, errors, 0)
}

func TestThatUserDoesntUseEnvironmentVariables(t *testing.T) {
	t.Run("run", func(t *testing.T) {
		man := manifest.Run{
			SaveArtifacts: []string{
				"path/to/$BUILD_VERSION/blah",
				"this/is/ok",
				"this/$IS/not",
			},
		}

		errors, _ := LintArtifacts(man, []manifest.Task{})
		assert.Len(t, errors, 2)
		assertContainsError(t, errors, ErrInvalidField.WithValue("save_artifact"))
	})

	t.Run("docker-compose", func(t *testing.T) {
		man := manifest.DockerCompose{
			SaveArtifacts: []string{
				"path/to/$BUILD_VERSION/blah",
				"this/is/ok",
				"this/$IS/not",
			},
		}

		errors, _ := LintArtifacts(man, []manifest.Task{})
		assert.Len(t, errors, 2)
		assertContainsError(t, errors, ErrInvalidField.WithValue("save_artifact"))
	})

	t.Run("deploy-cf", func(t *testing.T) {
		man := manifest.DeployCF{
			DeployArtifact: "path/to/$BUILD_VERSION/blah",
		}

		errors, _ := LintArtifacts(man, []manifest.Task{manifest.Run{SaveArtifacts: []string{"."}}})
		assert.Len(t, errors, 1)
		assertContainsError(t, errors, ErrInvalidField.WithValue("deploy_artifact"))
	})
}
