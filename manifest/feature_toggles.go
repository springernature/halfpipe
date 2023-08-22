package manifest

type FeatureToggles []string

const (
	FeatureUpdatePipeline       = "update-pipeline"
	FeatureUpdatePipelineAndTag = "update-pipeline-and-tag"
	FeatureGithubStatuses       = "github-statuses"
	FeatureUpdateActions        = "update-actions"
)

var AvailableFeatureToggles = FeatureToggles{
	FeatureUpdatePipeline,
	FeatureUpdatePipelineAndTag,
	FeatureGithubStatuses,
	FeatureUpdateActions,
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

func (f FeatureToggles) UpdateActions() bool {
	return f.contains(FeatureUpdateActions)
}
