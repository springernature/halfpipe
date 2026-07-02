package defaults

import (
	"testing"

	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestFeatureTogglesDefaulter_AddsUpdatePipelineWhenNoTogglesSet(t *testing.T) {
	man := manifest.Manifest{}

	updated := NewFeatureTogglesDefaulter().Apply(man)

	assert.Equal(t, manifest.FeatureToggles{manifest.FeatureUpdatePipeline}, updated.FeatureToggles)
}

func TestFeatureTogglesDefaulter_DoesNotDuplicateUpdatePipeline(t *testing.T) {
	man := manifest.Manifest{
		FeatureToggles: manifest.FeatureToggles{manifest.FeatureUpdatePipeline},
	}

	updated := NewFeatureTogglesDefaulter().Apply(man)

	assert.Equal(t, manifest.FeatureToggles{manifest.FeatureUpdatePipeline}, updated.FeatureToggles)
}

func TestFeatureTogglesDefaulter_DoesNotAddUpdatePipelineWhenUpdatePipelineAndTagSet(t *testing.T) {
	man := manifest.Manifest{
		FeatureToggles: manifest.FeatureToggles{manifest.FeatureUpdatePipelineAndTag},
	}

	updated := NewFeatureTogglesDefaulter().Apply(man)

	assert.Equal(t, manifest.FeatureToggles{manifest.FeatureUpdatePipelineAndTag}, updated.FeatureToggles)
}

func TestFeatureTogglesDefaulter_DoesNotAddUpdatePipelineWhenDisabled(t *testing.T) {
	man := manifest.Manifest{
		FeatureToggles: manifest.FeatureToggles{manifest.FeatureDisableUpdatePipeline},
	}

	updated := NewFeatureTogglesDefaulter().Apply(man)

	assert.Equal(t, manifest.FeatureToggles{manifest.FeatureDisableUpdatePipeline}, updated.FeatureToggles)
}

func TestFeatureTogglesDefaulter_DisableDoesNotAffectExplicitUpdatePipelineAndTag(t *testing.T) {
	man := manifest.Manifest{
		FeatureToggles: manifest.FeatureToggles{manifest.FeatureDisableUpdatePipeline, manifest.FeatureUpdatePipelineAndTag},
	}

	updated := NewFeatureTogglesDefaulter().Apply(man)

	assert.Equal(t, manifest.FeatureToggles{manifest.FeatureDisableUpdatePipeline, manifest.FeatureUpdatePipelineAndTag}, updated.FeatureToggles)
}
