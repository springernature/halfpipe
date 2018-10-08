package linters

import (
	"testing"

	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestAllMissing(t *testing.T) {
	man := manifest.Manifest{}
	result := manifestLinter{}.Lint(man)
	assert.Len(t, result.Errors, 2)
}

func TestTeamIsMissing(t *testing.T) {
	man := manifest.Manifest{}
	man.Pipeline = "yolo"

	result := manifestLinter{}.Lint(man)
	assert.Len(t, result.Errors, 1)
	assertMissingField(t, "team", result.Errors[0])
}

func TestTeamIsUpperCase(t *testing.T) {
	man := manifest.Manifest{}
	man.Pipeline = "yolo"
	man.Team = "yoLo"

	result := manifestLinter{}.Lint(man)
	assert.Len(t, result.Warnings, 1)
	assert.Len(t, result.Errors, 0)
	assertInvalidField(t, "team", result.Warnings[0])
}

func TestPipelineIsMissing(t *testing.T) {
	man := manifest.Manifest{}
	man.Team = "yolo"

	result := manifestLinter{}.Lint(man)
	assert.Len(t, result.Errors, 1)
	assertMissingField(t, "pipeline", result.Errors[0])
}

func TestPipelineIsValid(t *testing.T) {
	man := manifest.Manifest{}
	man.Team = "yolo"
	man.Pipeline = "Something with spaces"

	result := manifestLinter{}.Lint(man)
	assert.True(t, result.HasErrors())
}

func TestHappyPath(t *testing.T) {
	man := manifest.Manifest{
		Team:     "yolo",
		Pipeline: "alles-gut",
		ArtifactConfig: manifest.ArtifactConfig{
			Bucket:  "someBucket",
			JsonKey: "someKey",
		},
	}

	result := manifestLinter{}.Lint(man)
	assert.False(t, result.HasErrors())
}

func TestMissingFieldInArtifactConfig(t *testing.T) {
	missingJsonKey := manifest.Manifest{
		Team:     "team",
		Pipeline: "pipeline",
		ArtifactConfig: manifest.ArtifactConfig{
			Bucket: "notEmpty",
		},
	}

	result := manifestLinter{}.Lint(missingJsonKey)
	assert.True(t, result.HasErrors())
	assertInvalidFieldInErrors(t, "artifact_config", result.Errors)

	missingBucket := manifest.Manifest{
		Team:     "team",
		Pipeline: "pipeline",
		ArtifactConfig: manifest.ArtifactConfig{
			JsonKey: "notEmpty",
		},
	}

	result2 := manifestLinter{}.Lint(missingBucket)
	assert.True(t, result2.HasErrors())
	assertInvalidFieldInErrors(t, "artifact_config", result2.Errors)
}
