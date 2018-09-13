package tasks

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