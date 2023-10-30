package manifest

import "golang.org/x/exp/slices"

type FeatureToggles []string

const (
	FeatureUpdatePipeline       = "update-pipeline"
	FeatureUpdatePipelineAndTag = "update-pipeline-and-tag"
	FeatureGithubStatuses       = "github-statuses"
)

var AvailableFeatureToggles = FeatureToggles{
	FeatureUpdatePipeline,
	FeatureUpdatePipelineAndTag,
	FeatureGithubStatuses,
}

func (f FeatureToggles) UpdatePipeline() bool {
	return slices.Contains(f, FeatureUpdatePipeline) || slices.Contains(f, FeatureUpdatePipelineAndTag)
}

func (f FeatureToggles) TagGitRepo() bool {
	return slices.Contains(f, FeatureUpdatePipelineAndTag)
}

func (f FeatureToggles) GithubStatuses() bool {
	return slices.Contains(f, FeatureGithubStatuses)
}
