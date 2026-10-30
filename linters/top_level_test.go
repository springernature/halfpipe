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
	}

	result := topLevelLinter{}.Lint(man)
	assert.False(t, result.HasErrors())
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
