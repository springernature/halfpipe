package manifest

type FeatureToggles []string

const FeatureVersioned = "versioned"
const FeatureAutoUpdate = "auto-update"

var AvailableFeatureToggles = FeatureToggles{
	FeatureVersioned,
	FeatureAutoUpdate,
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
	return f.contains(FeatureVersioned) || f.contains(FeatureAutoUpdate)
}

func (f FeatureToggles) AutoUpdate() bool {
	return f.contains(FeatureAutoUpdate)
}
