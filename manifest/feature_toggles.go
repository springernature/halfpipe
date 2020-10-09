package manifest

type FeatureToggles []string

const FeatureUpdatePipeline = "update-pipeline"
const FeatureDockerDecompose = "docker-decompose"
const FeatureOldDeployResource = "old-deploy-resource"
const FeatureGithubActions = "github-actions"

var AvailableFeatureToggles = FeatureToggles{
	FeatureUpdatePipeline,
	FeatureDockerDecompose,
	FeatureOldDeployResource,
	FeatureGithubActions,
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

func (f FeatureToggles) OldDeployResource() bool {
	return f.contains(FeatureOldDeployResource)
}

func (f FeatureToggles) UpdatePipeline() bool {
	return f.contains(FeatureUpdatePipeline)
}

func (f FeatureToggles) DockerDecompose() bool {
	return f.contains(FeatureDockerDecompose)
}

func (f FeatureToggles) GithubActions() bool {
	return f.contains(FeatureGithubActions)
}
