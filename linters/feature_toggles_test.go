package linters

import (
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDoesNothingIfNoFeaturesAvailable(t *testing.T) {
	assert.False(t, NewFeatureToggleLinter(manifest.FeatureToggles{}).Lint(manifest.Manifest{}).HasErrors())
}

func TestErrorsIfUnknownFeatureToggle(t *testing.T) {
	availableFeatures := manifest.FeatureToggles{
		"featurea",
	}

	man := manifest.Manifest{
		FeatureToggles: manifest.FeatureToggles{
			"featurea",
			"featureb",
		},
	}

	result := NewFeatureToggleLinter(availableFeatures).Lint(man)
	assert.True(t, result.HasErrors())
	assert.Equal(t, ErrNonSupportedFeature("featureb"), result.Errors[0])
}

func TestDoesNothingIfAllFeaturesAreAvailable(t *testing.T) {
	availableFeatures := manifest.FeatureToggles{
		"featurea",
		"featureb",
		"featurec",
		"featured",
	}

	man := manifest.Manifest{
		FeatureToggles: availableFeatures,
	}

	result := NewFeatureToggleLinter(availableFeatures).Lint(man)
	assert.False(t, result.HasErrors())

}
