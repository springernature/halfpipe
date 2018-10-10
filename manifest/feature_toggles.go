package manifest

type FeatureToggles []string

const FeatureVersioned = "versioned"

var AvailableFeatureToggles = FeatureToggles{
	FeatureVersioned,
}

func (f FeatureToggles) Versioned() bool {
	for _, feature := range f {
		if feature == FeatureVersioned {
			return true
		}
	}
	return false
}
