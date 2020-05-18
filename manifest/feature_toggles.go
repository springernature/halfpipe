package manifest

type FeatureToggles []string

const FeatureUpdatePipeline = "update-pipeline"
const FeatureDockerDecompose = "docker-decompose"
const FeatureNewDeployResource = "new-deploy-resource"
const FeatureDisableDeprecatedDockerRegistryError = "im-aware-that-old-docker-registries-will-stop-working-on-24-august-2020"
const FeatureDisableDeprecatedNexusRepositoryError = "im-aware-that-repo-dot-tools-will-stop-working-on-24-august-2020"

var AvailableFeatureToggles = FeatureToggles{
	FeatureUpdatePipeline,
	FeatureDockerDecompose,
	FeatureNewDeployResource,
	FeatureDisableDeprecatedDockerRegistryError,
	FeatureDisableDeprecatedNexusRepositoryError,
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
	return f.UpdatePipeline()
}

func (f FeatureToggles) NewDeployResource() bool {
	return f.contains(FeatureNewDeployResource)
}

func (f FeatureToggles) UpdatePipeline() bool {
	return f.contains(FeatureUpdatePipeline)
}

func (f FeatureToggles) DockerDecompose() bool {
	return f.contains(FeatureDockerDecompose)
}

func (f FeatureToggles) DisableDeprecatedDockerRegistryError() bool {
	return f.contains(FeatureDisableDeprecatedDockerRegistryError)
}

func (f FeatureToggles) DisableDeprecatedNexusRepositoryError() bool {
	return f.contains(FeatureDisableDeprecatedNexusRepositoryError)
}
