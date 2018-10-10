package linters

import (
	"fmt"
	"github.com/springernature/halfpipe/linters/result"
	"github.com/springernature/halfpipe/manifest"
)

var ErrNonSupportedFeature = func(feature string) error {
	return fmt.Errorf("feature '%s' is not supported", feature)
}

type featureToggleLinter struct {
	availibleFeatures manifest.FeatureToggles
}

func NewFeatureToggleLinter(availibleFeatures manifest.FeatureToggles) featureToggleLinter {
	return featureToggleLinter{
		availibleFeatures: availibleFeatures,
	}
}

func (f featureToggleLinter) Lint(manifest manifest.Manifest) (result result.LintResult) {
	result.Linter = "Feature Toggles Linter"
	result.DocsURL = "https://docs.halfpipe.io/manifest/#feature-toggles"

	for _, feature := range manifest.FeatureToggles {
		if !f.featureInAvailableFeatures(feature) {
			result.AddError(ErrNonSupportedFeature(feature))
		}
	}
	return
}

func (f featureToggleLinter) featureInAvailableFeatures(feature string) bool {
	for _, availibleFeature := range f.availibleFeatures {
		if feature == availibleFeature {
			return true
		}
	}
	return false
}
