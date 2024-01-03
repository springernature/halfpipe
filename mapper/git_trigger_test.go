package mapper

import (
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGitTriggerMapper_Apply(t *testing.T) {
	t.Run("does nothing on empty manifest", func(t *testing.T) {
		input := manifest.Manifest{}
		expected := manifest.Manifest{}

		updated, err := NewGitTriggerMapper().Apply(input)

		assert.NoError(t, err)
		assert.Equal(t, expected, updated)
	})

	t.Run("does nothing when there is no dot in the watched paths", func(t *testing.T) {
		input := manifest.Manifest{
			Triggers: manifest.TriggerList{
				manifest.GitTrigger{WatchedPaths: []string{"a", "b"}},
			},
		}
		expected := manifest.Manifest{
			Triggers: manifest.TriggerList{
				manifest.GitTrigger{WatchedPaths: []string{"a", "b"}},
			},
		}

		updated, err := NewGitTriggerMapper().Apply(input)

		assert.NoError(t, err)
		assert.Equal(t, expected, updated)
	})

	t.Run("converts dots to base path", func(t *testing.T) {
		input := manifest.Manifest{
			Triggers: manifest.TriggerList{
				manifest.GitTrigger{WatchedPaths: []string{"a", "."}, BasePath: "some/path/to/something"},
			},
		}
		expected := manifest.Manifest{
			Triggers: manifest.TriggerList{
				manifest.GitTrigger{WatchedPaths: []string{"a", "some/path/to/something"}, BasePath: "some/path/to/something"},
			},
		}

		updated, err := NewGitTriggerMapper().Apply(input)

		assert.NoError(t, err)
		assert.Equal(t, expected, updated)
	})

}
