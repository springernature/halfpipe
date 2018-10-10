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
	availibleFeatures := manifest.FeatureToggles{
		"featurea",
	}

	man := manifest.Manifest{
		FeatureToggles: manifest.FeatureToggles{
			"featurea",
			"featureb",
		},
	}

	result := NewFeatureToggleLinter(availibleFeatures).Lint(man)
	assert.True(t, result.HasErrors())
	assert.Equal(t, ErrNonSupportedFeature("featureb"), result.Errors[0])
}

func TestDoesNothingIfAllFeaturesAreAvailable(t *testing.T) {
	availibleFeatures := manifest.FeatureToggles{
		"featurea",
		"featureb",
		"featurec",
		"featured",
	}

	man := manifest.Manifest{
		FeatureToggles: availibleFeatures,
	}

	result := NewFeatureToggleLinter(availibleFeatures).Lint(man)
	assert.False(t, result.HasErrors())

}
