package linters

import (
	"testing"

	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestAllMissing(t *testing.T) {
	man := manifest.Manifest{}
	result := topLevelLinter{}.Lint(man)
	assert.Len(t, result.Errors, 2)
}

func TestTeamIsMissing(t *testing.T) {
	man := manifest.Manifest{}
	man.Pipeline = "yolo"

	result := topLevelLinter{}.Lint(man)
	assert.Len(t, result.Errors, 1)
	assertMissingField(t, "team", result.Errors[0])
}

func TestTeamIsUpperCase(t *testing.T) {
	man := manifest.Manifest{}
	man.Pipeline = "yolo"
	man.Team = "yoLo"

	result := topLevelLinter{}.Lint(man)
	assert.Len(t, result.Warnings, 1)
	assert.Len(t, result.Errors, 0)
	assertInvalidField(t, "team", result.Warnings[0])
}

func TestPipelineIsMissing(t *testing.T) {
	man := manifest.Manifest{}
	man.Team = "yolo"

	result := topLevelLinter{}.Lint(man)
	assert.Len(t, result.Errors, 1)
	assertMissingField(t, "pipeline", result.Errors[0])
}

func TestPipelineIsValid(t *testing.T) {
	man := manifest.Manifest{}
	man.Team = "yolo"
	man.Pipeline = "Something with spaces"

	result := topLevelLinter{}.Lint(man)
	assert.True(t, result.HasErrors())
}

func TestHappyPath(t *testing.T) {
	man := manifest.Manifest{
		Team:     "yolo",
		Pipeline: "alles-gut",
		ArtifactConfig: manifest.ArtifactConfig{
			Bucket:  "someBucket",
			JSONKey: "someKey",
		},
	}

	result := topLevelLinter{}.Lint(man)
	assert.False(t, result.HasErrors())
}

func TestMissingFieldInArtifactConfig(t *testing.T) {
	missingJSONKey := manifest.Manifest{
		Team:     "team",
		Pipeline: "pipeline",
		ArtifactConfig: manifest.ArtifactConfig{
			Bucket: "notEmpty",
		},
	}

	result := topLevelLinter{}.Lint(missingJSONKey)
	assert.True(t, result.HasErrors())
	assertInvalidFieldInErrors(t, "artifact_config", result.Errors)

	missingBucket := manifest.Manifest{
		Team:     "team",
		Pipeline: "pipeline",
		ArtifactConfig: manifest.ArtifactConfig{
			JSONKey: "notEmpty",
		},
	}

	result2 := topLevelLinter{}.Lint(missingBucket)
	assert.True(t, result2.HasErrors())
	assertInvalidFieldInErrors(t, "artifact_config", result2.Errors)
}

func TestNotifyOnSuccess(t *testing.T) {
	t.Run("when its a top level task", func(t *testing.T) {
		t.Run("and slack channel is not set it gives an error", func(t *testing.T) {
			man := manifest.Manifest{
				Team:     "team",
				Pipeline: "pipeline",
				Tasks: manifest.TaskList{
					manifest.Run{},
					manifest.DockerCompose{NotifyOnSuccess: true},
					manifest.Run{},
				},
			}

			result := NewTopLevelLinter().Lint(man)
			assertInvalidFieldInErrors(t, "slack_channel", result.Errors)
		})

		t.Run("and slack channel is set it gives no error", func(t *testing.T) {
			man := manifest.Manifest{
				Team:         "team",
				Pipeline:     "pipeline",
				SlackChannel: "#meehp",
				Tasks: manifest.TaskList{
					manifest.Run{},
					manifest.DockerCompose{NotifyOnSuccess: true},
					manifest.Run{},
				},
			}

			result := NewTopLevelLinter().Lint(man)
			assert.Len(t, result.Errors, 0)
		})
	})

	t.Run("when its a pre promote task", func(t *testing.T) {
		t.Run("and slack channel is not set it gives an error", func(t *testing.T) {
			man := manifest.Manifest{
				Team:     "team",
				Pipeline: "pipeline",
				Tasks: manifest.TaskList{
					manifest.Run{},
					manifest.DeployCF{
						PrePromote: manifest.TaskList{
							manifest.DockerCompose{NotifyOnSuccess: true},
						},
					},
					manifest.Run{},
				},
			}

			result := NewTopLevelLinter().Lint(man)
			assertInvalidFieldInErrors(t, "slack_channel", result.Errors)
		})

		t.Run("and slack channel is set it gives no error", func(t *testing.T) {
			man := manifest.Manifest{
				Team:         "team",
				Pipeline:     "pipeline",
				SlackChannel: "#meehp",
				Tasks: manifest.TaskList{
					manifest.Run{},
					manifest.DeployCF{
						PrePromote: manifest.TaskList{
							manifest.DockerCompose{NotifyOnSuccess: true},
						},
					},
					manifest.Run{},
				},
			}

			result := NewTopLevelLinter().Lint(man)
			assert.Len(t, result.Errors, 0)
		})
	})
}
