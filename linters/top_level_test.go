package linters

import (
	"testing"

	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestAllMissing(t *testing.T) {
	man := manifest.Manifest{}
	result := topLevelLinter{}.Lint(man)
	assert.Len(t, result.Errors, 3)
}

func TestTeamIsMissing(t *testing.T) {
	man := manifest.Manifest{}
	man.Pipeline = "yolo"
	man.Output = "concourse"

	result := topLevelLinter{}.Lint(man)
	assert.Len(t, result.Errors, 1)
	assertMissingField(t, "team", result.Errors[0])
}

func TestTeamIsUpperCase(t *testing.T) {
	man := manifest.Manifest{}
	man.Pipeline = "yolo"
	man.Team = "yoLo"
	man.Output = "concourse"

	result := topLevelLinter{}.Lint(man)
	assert.Len(t, result.Warnings, 1)
	assert.Len(t, result.Errors, 0)
	assertInvalidField(t, "team", result.Warnings[0])
}

func TestPipelineIsMissing(t *testing.T) {
	man := manifest.Manifest{}
	man.Team = "yolo"
	man.Output = "concourse"

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
		Output:   "actions",
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

func TestOutput(t *testing.T) {
	t.Run("set to action", func(t *testing.T) {
		man := manifest.Manifest{
			Pipeline: "kehe",
			Team:     "kehe",
			Output:   "actions",
		}
		result := topLevelLinter{}.Lint(man)
		assert.Empty(t, result.Errors)
		assert.Empty(t, result.Warnings)
	})

	t.Run("set to concourse", func(t *testing.T) {
		man := manifest.Manifest{
			Pipeline: "kehe",
			Team:     "kehe",
			Output:   "actions",
		}
		result := topLevelLinter{}.Lint(man)
		assert.Empty(t, result.Errors)
		assert.Empty(t, result.Warnings)
	})

	t.Run("set to travis", func(t *testing.T) {
		man := manifest.Manifest{
			Pipeline: "kehe",
			Team:     "kehe",
			Output:   "travis",
		}
		result := topLevelLinter{}.Lint(man)
		assert.Len(t, result.Errors, 1)
		assert.Empty(t, result.Warnings)
		assertInvalidFieldInErrors(t, "output", result.Errors)
	})
}
