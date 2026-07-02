package defaults

import "github.com/springernature/halfpipe/manifest"

type FeatureTogglesDefaulter interface {
	Apply(original manifest.Manifest) (updated manifest.Manifest)
}

type featureTogglesDefaulter struct {
}

func (f featureTogglesDefaulter) Apply(original manifest.Manifest) (updated manifest.Manifest) {
	updated = original

	if !updated.FeatureToggles.DisableUpdatePipeline() && !updated.FeatureToggles.UpdatePipeline() {
		updated.FeatureToggles = append(updated.FeatureToggles, manifest.FeatureUpdatePipeline)
	}

	return
}

func NewFeatureTogglesDefaulter() FeatureTogglesDefaulter {
	return featureTogglesDefaulter{}
}
