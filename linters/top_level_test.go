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
	man.Platform = "concourse"

	result := topLevelLinter{}.Lint(man)
	AssertContainsError(t, result.Errors, NewErrMissingField("team"))
}

func TestTeamIsUpperCase(t *testing.T) {
	man := manifest.Manifest{}
	man.Pipeline = "yolo"
	man.Team = "yoLo"
	man.Platform = "concourse"

	result := topLevelLinter{}.Lint(man)
	AssertContainsError(t, result.Warnings, ErrInvalidField.WithValue("team"))
}

func TestPipelineIsMissing(t *testing.T) {
	man := manifest.Manifest{}
	man.Team = "yolo"
	man.Platform = "concourse"

	result := topLevelLinter{}.Lint(man)
	AssertContainsError(t, result.Errors, NewErrMissingField("pipeline"))
}

func TestHappyPath(t *testing.T) {
	man := manifest.Manifest{
		Team:     "yolo",
		Pipeline: "alles-gut",
		Platform: "actions",
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
	AssertContainsError(t, result.Errors, ErrInvalidField.WithValue("artifact_config"))

	missingBucket := manifest.Manifest{
		Team:     "team",
		Pipeline: "pipeline",
		ArtifactConfig: manifest.ArtifactConfig{
			JSONKey: "notEmpty",
		},
	}

	result2 := topLevelLinter{}.Lint(missingBucket)
	AssertContainsError(t, result2.Errors, ErrInvalidField.WithValue("artifact_config"))
}

func TestOutput(t *testing.T) {
	t.Run("set to action", func(t *testing.T) {
		man := manifest.Manifest{
			Pipeline: "kehe",
			Team:     "kehe",
			Platform: "actions",
		}
		result := topLevelLinter{}.Lint(man)
		assert.Empty(t, result.Errors)
		assert.Empty(t, result.Warnings)
	})

	t.Run("set to concourse", func(t *testing.T) {
		man := manifest.Manifest{
			Pipeline: "kehe",
			Team:     "kehe",
			Platform: "actions",
		}
		result := topLevelLinter{}.Lint(man)
		assert.Empty(t, result.Errors)
		assert.Empty(t, result.Warnings)
	})

	t.Run("set to travis", func(t *testing.T) {
		man := manifest.Manifest{
			Pipeline: "kehe",
			Team:     "kehe",
			Platform: "travis",
		}
		result := topLevelLinter{}.Lint(man)
		AssertContainsError(t, result.Errors, ErrInvalidField.WithValue("platform"))
	})
}
