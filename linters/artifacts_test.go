package linters

import (
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCurrentTaskDoesNotRequireArtifactAndThereAreNoPreviousTasks(t *testing.T) {
	currentTask := manifest.DockerPush{}
	errs := LintArtifacts(currentTask, nil)
	assert.Len(t, errs, 0)
}

func TestCurrentTaskDoesNotRequireArtifactAndThereArePreviousTasks(t *testing.T) {
	currentTask := manifest.DockerPush{}
	errors := LintArtifacts(currentTask, []manifest.Task{manifest.DockerCompose{}, manifest.Run{}})
	assert.Len(t, errors, 0)
}

func TestCurrentTaskRequiresArtifactAndThereArePreviousTasks(t *testing.T) {
	currentTask := manifest.DockerPush{RestoreArtifacts: true}
	errs := LintArtifacts(currentTask, nil)
	assertContainsError(t, errs, ErrReadsFromSavedArtifacts)
}

func TestCurrentTaskRequiresArtifactAndThereArePreviousTasksThatDoesntSaveAny(t *testing.T) {
	currentTask := manifest.DockerPush{RestoreArtifacts: true}
	errs := LintArtifacts(currentTask, []manifest.Task{manifest.DockerCompose{}, manifest.Run{}})
	assertContainsError(t, errs, ErrReadsFromSavedArtifacts)
}

func TestCurrentTaskRequiresArtifactAndThereIsAPreviousTasksThatSavesOne(t *testing.T) {
	currentTask := manifest.DockerPush{RestoreArtifacts: true}
	errs := LintArtifacts(currentTask, []manifest.Task{manifest.DockerCompose{SaveArtifacts: []string{"path/to/artifact/to/save"}}, manifest.Run{}})
	assert.Len(t, errs, 0)
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

		errs := LintArtifacts(man, []manifest.Task{})
		assert.Len(t, errs, 2)
		assert.ErrorIs(t, errs[0], ErrInvalidField.WithValue("save_artifact"))
		assert.ErrorIs(t, errs[1], ErrInvalidField.WithValue("save_artifact"))
	})

	t.Run("docker-compose", func(t *testing.T) {
		man := manifest.DockerCompose{
			SaveArtifacts: []string{
				"path/to/$BUILD_VERSION/blah",
				"this/is/ok",
				"this/$IS/not",
			},
		}

		errors := LintArtifacts(man, []manifest.Task{})
		assert.Len(t, errors, 2)
		assertContainsError(t, errors, ErrInvalidField.WithValue("save_artifact"))
	})

	t.Run("deploy-cf", func(t *testing.T) {
		man := manifest.DeployCF{
			DeployArtifact: "path/to/$BUILD_VERSION/blah",
		}

		errors := LintArtifacts(man, []manifest.Task{manifest.Run{SaveArtifacts: []string{"."}}})
		assert.Len(t, errors, 1)
		assertContainsError(t, errors, ErrInvalidField.WithValue("deploy_artifact"))
	})
}
