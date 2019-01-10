package manifest

type FeatureToggles []string

const FeatureVersioned = "versioned"
const FeatureUpdatePipeline = "update-pipeline"

var AvailableFeatureToggles = FeatureToggles{
	FeatureVersioned,
	FeatureUpdatePipeline,
}

func (f FeatureToggles) contains(aFeature string) bool {
	for _, feature := range f {
		if feature == aFeature {
			return true
		}
	}
	return false
}

func (f FeatureToggles) Versioned() bool {
	return f.contains(FeatureVersioned) || f.contains(FeatureUpdatePipeline)
}

func (f FeatureToggles) UpdatePipeline() bool {
	return f.contains(FeatureUpdatePipeline)
}
