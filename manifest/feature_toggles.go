package manifest

type FeatureToggles []string

const FeatureUpdatePipeline = "update-pipeline"
const FeatureUpdatePipelineAndTag = "update-pipeline-and-tag"

const FeatureDockerDecompose = "docker-decompose"
const FeatureDockerOciBuild = "docker-oci-build"

var AvailableFeatureToggles = FeatureToggles{
	FeatureUpdatePipeline,
	FeatureUpdatePipelineAndTag,
	FeatureDockerDecompose,
	FeatureDockerOciBuild,
}

func (f FeatureToggles) contains(aFeature string) bool {
	for _, feature := range f {
		if feature == aFeature {
			return true
		}
	}
	return false
}

func (f FeatureToggles) UpdatePipeline() bool {
	return f.contains(FeatureUpdatePipeline) || f.contains(FeatureUpdatePipelineAndTag)
}

func (f FeatureToggles) TagGitRepo() bool {
	return f.contains(FeatureUpdatePipelineAndTag)
}

func (f FeatureToggles) DockerDecompose() bool {
	return f.contains(FeatureDockerDecompose)
}

func (f FeatureToggles) DockerOciBuild() bool {
	return f.contains(FeatureDockerOciBuild)
}
