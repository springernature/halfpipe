package linters

import (
	"fmt"
	"github.com/springernature/halfpipe/linters/result"
	"github.com/springernature/halfpipe/manifest"
)

var ErrNonSupportedFeature = func(feature string) error {
	if feature == "versioned" {
		return fmt.Errorf("feature '%s' is no longer supported. The same functionality is included in the 'update-pipeline' feature", feature)
	}
	return fmt.Errorf("feature '%s' is not supported", feature)
}

type featureToggleLinter struct {
	availableFeatures manifest.FeatureToggles
}

func NewFeatureToggleLinter(availableFeatures manifest.FeatureToggles) featureToggleLinter {
	return featureToggleLinter{
		availableFeatures: availableFeatures,
	}
}

func (f featureToggleLinter) Lint(manifest manifest.Manifest) (result result.LintResult) {
	result.Linter = "Feature Toggles"
	result.DocsURL = "https://ee.public.springernature.app/rel-eng/halfpipe/manifest/#feature_toggles"

	for _, feature := range manifest.FeatureToggles {
		if !f.featureInAvailableFeatures(feature) {
			result.AddError(ErrNonSupportedFeature(feature))
		}
	}
	return result
}

func (f featureToggleLinter) featureInAvailableFeatures(feature string) bool {
	for _, availableFeature := range f.availableFeatures {
		if feature == availableFeature {
			return true
		}
	}
	return false
}
