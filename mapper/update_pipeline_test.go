package mapper

import (
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUpdatePipeline(t *testing.T) {
	mapper := NewUpdatePipelineMapper()

	originalManifest := manifest.Manifest{
		Tasks: manifest.TaskList{
			manifest.Run{},
			manifest.DeployCF{},
		},
	}

	t.Run("adds update job as first job", func(t *testing.T) {
		man := originalManifest

		expectedTasks := manifest.TaskList{
			manifest.Update{},
			manifest.Run{},
			manifest.DeployCF{},
		}

		updated, err := mapper.Apply(man)
		assert.NoError(t, err)
		assert.Equal(t, expectedTasks, updated.Tasks)
	})

	t.Run("adds update job as first job if update-pipeline-and-tag feature is enabled", func(t *testing.T) {
		man := originalManifest
		man.FeatureToggles = manifest.FeatureToggles{manifest.FeatureUpdatePipelineAndTag}

		expectedTasks := manifest.TaskList{
			manifest.Update{TagRepo: true},
			manifest.Run{},
			manifest.DeployCF{},
		}

		updated, err := mapper.Apply(man)
		assert.NoError(t, err)
		assert.Equal(t, expectedTasks, updated.Tasks)
	})
}
