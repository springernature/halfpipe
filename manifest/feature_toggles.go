package manifest

type FeatureToggles []string

const FeatureUpdatePipeline = "update-pipeline"
const FeatureUpdatePipelineAndTag = "update-pipeline-and-tag"
const FeatureGithubStatuses = "github-statuses"

var AvailableFeatureToggles = FeatureToggles{
	FeatureUpdatePipeline,
	FeatureUpdatePipelineAndTag,
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

func (f FeatureToggles) GithubStatuses() bool {
	return f.contains(FeatureGithubStatuses)
}
