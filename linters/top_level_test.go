package linters

import (
	"testing"

	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestAllMissing(t *testing.T) {
	man := manifest.Manifest{}
	result := topLevelLinter{}.Lint(man)
	assert.Len(t, result.Issues, 3)
}

func TestTeamIsMissing(t *testing.T) {
	man := manifest.Manifest{}
	man.Pipeline = "yolo"
	man.Platform = "concourse"

	result := topLevelLinter{}.Lint(man)
	assertContainsError(t, result.Issues, NewErrMissingField("team"))
}

func TestTeamIsUpperCase(t *testing.T) {
	man := manifest.Manifest{}
	man.Pipeline = "yolo"
	man.Team = "yoLo"
	man.Platform = "concourse"

	result := topLevelLinter{}.Lint(man)
	assertContainsError(t, result.Issues, ErrInvalidField.WithValue("team"))
}

func TestPipelineIsMissing(t *testing.T) {
	man := manifest.Manifest{}
	man.Team = "yolo"
	man.Platform = "concourse"

	result := topLevelLinter{}.Lint(man)
	assertContainsError(t, result.Issues, NewErrMissingField("pipeline"))
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
	assertContainsError(t, result.Issues, ErrInvalidField.WithValue("artifact_config"))

	missingBucket := manifest.Manifest{
		Team:     "team",
		Pipeline: "pipeline",
		ArtifactConfig: manifest.ArtifactConfig{
			JSONKey: "notEmpty",
		},
	}

	result2 := topLevelLinter{}.Lint(missingBucket)
	assertContainsError(t, result2.Issues, ErrInvalidField.WithValue("artifact_config"))
}

func TestOutput(t *testing.T) {
	t.Run("set to action", func(t *testing.T) {
		man := manifest.Manifest{
			Pipeline: "kehe",
			Team:     "kehe",
			Platform: "actions",
		}
		result := topLevelLinter{}.Lint(man)
		assert.Empty(t, result.Issues)
	})

	t.Run("set to concourse", func(t *testing.T) {
		man := manifest.Manifest{
			Pipeline: "kehe",
			Team:     "kehe",
			Platform: "actions",
		}
		result := topLevelLinter{}.Lint(man)
		assert.Empty(t, result.Issues)
	})

	t.Run("set to travis", func(t *testing.T) {
		man := manifest.Manifest{
			Pipeline: "kehe",
			Team:     "kehe",
			Platform: "travis",
		}
		result := topLevelLinter{}.Lint(man)
		assertContainsError(t, result.Issues, ErrInvalidField.WithValue("platform"))
	})
}

func TestNotifications(t *testing.T) {
	t.Run("deprecated fields", func(t *testing.T) {
		man := manifest.Manifest{
			Team:                "team",
			Pipeline:            "pipeline",
			Platform:            "concourse",
			SlackSuccessMessage: "Blah",
			SlackFailureMessage: "Bluh",
		}

		result := NewTopLevelLinter().Lint(man)
		assert.True(t, result.HasWarnings())
		assert.False(t, result.HasErrors())
		assert.Len(t, result.Issues, 2)
		assertContainsError(t, result.Issues, ErrSlackSuccessMessageFieldDeprecated)
		assertContainsError(t, result.Issues, ErrSlackFailureMessageFieldDeprecated)
	})

	t.Run("both slack and teams", func(t *testing.T) {
		man := manifest.Manifest{
			Team:     "team",
			Pipeline: "pipeline",
			Platform: "concourse",
			Notifications: manifest.Notifications{
				Success: manifest.NotificationChannels{
					{Slack: "a", Teams: "b"},
				},
				Failure: manifest.NotificationChannels{
					{Slack: "a", Teams: "b"},
				},
			},
		}

		result := NewTopLevelLinter().Lint(man)
		assert.True(t, result.HasErrors())
		assert.Len(t, result.Issues, 2)
	})
}
