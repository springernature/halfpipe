package manifest

import "slices"

// FeatureToggles enable optional pipeline behaviours.
//
// | Toggle | Description |
// |--------|-------------|
// | `update-pipeline` | Inserts a job that keeps the pipeline/workflow in sync with the halfpipe manifest. Sets BUILD_VERSION. |
// | `update-pipeline-and-tag` | Like update-pipeline, but also tags the git repo with `<PIPELINE_NAME>/v<BUILD_VERSION>`. |
// | `github-statuses` | Updates GitHub commit statuses from Concourse job results (Actions does this by default). |
// | `ghas` | Enables GitHub Advanced Security scanning on docker-push tasks. |
type FeatureToggles []string

const (
	FeatureUpdatePipeline       = "update-pipeline"
	FeatureUpdatePipelineAndTag = "update-pipeline-and-tag"
	FeatureGithubStatuses       = "github-statuses"
	FeatureGhas                 = "ghas"
)

var AvailableFeatureToggles = FeatureToggles{
	FeatureUpdatePipeline,
	FeatureUpdatePipelineAndTag,
	FeatureGithubStatuses,
	FeatureGhas,
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

func (f FeatureToggles) Ghas() bool {
	return slices.Contains(f, FeatureGhas)
}
