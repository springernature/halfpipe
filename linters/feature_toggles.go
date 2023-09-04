package linters

import (
	"github.com/springernature/halfpipe/manifest"
	"golang.org/x/exp/slices"
)

var (
	ErrUnsupportedFeature          = newError("unsupported feature")
	ErrUnsupportedFeatureVersioned = newError("feature 'versioned' is no longer supported. The same functionality is included in the 'update-pipeline' feature")
	ErrUnsupportedDockerDecompose  = newError("feature 'docker-decompose' is no longer supported. docker-compose tasks will not be modified")
)

type featureToggleLinter struct {
	availableFeatures manifest.FeatureToggles
}

func NewFeatureToggleLinter(availableFeatures manifest.FeatureToggles) featureToggleLinter {
	return featureToggleLinter{
		availableFeatures: availableFeatures,
	}
}

func (f featureToggleLinter) Lint(manifest manifest.Manifest) (result LintResult) {
	result.Linter = "Feature Toggles"
	result.DocsURL = "https://ee.public.springernature.app/rel-eng/halfpipe/manifest/#feature_toggles"

	for _, feature := range manifest.FeatureToggles {
		if !f.featureInAvailableFeatures(feature) {
			if feature == "versioned" {
				result.Add(ErrUnsupportedFeatureVersioned.AsWarning())
			} else if feature == "docker-decompose" {
				result.Add(ErrUnsupportedDockerDecompose.AsWarning())
			} else {
				result.Add(ErrUnsupportedFeature.WithValue(feature).AsWarning())
			}
		}
	}
	return result
}

func (f featureToggleLinter) featureInAvailableFeatures(feature string) bool {
	return slices.ContainsFunc(f.availableFeatures, func(availableFeature string) bool { return availableFeature == feature })
}
