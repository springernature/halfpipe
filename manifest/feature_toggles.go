package manifest

type FeatureToggles []string

const FeatureUpdatePipeline = "update-pipeline"
const FeatureUpdatePipelineAndTag = "update-pipeline-and-tag"

const FeatureDockerOldBuild = "docker-old-build"

const FeatureGithubStatuses = "github-statuses"

var AvailableFeatureToggles = FeatureToggles{
	FeatureUpdatePipeline,
	FeatureUpdatePipelineAndTag,
	FeatureDockerOldBuild,
	FeatureGithubStatuses,
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

func (f FeatureToggles) DockerOldBuild() bool {
	return f.contains(FeatureDockerOldBuild)
}

func (f FeatureToggles) GithubStatuses() bool {
	return f.contains(FeatureGithubStatuses)
}
