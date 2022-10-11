package linters

import (
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDoesNothingIfNoFeaturesAvailable(t *testing.T) {
	assert.False(t, NewFeatureToggleLinter(manifest.FeatureToggles{}).Lint(manifest.Manifest{}).HasErrors())
}

func TestWarningIfUnknownFeatureToggle(t *testing.T) {
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
	assert.ErrorIs(t, result.Errors[0], ErrUnsupportedFeature.WithValue("featureb"))
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
