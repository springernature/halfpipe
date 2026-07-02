package manifest

import "slices"

// FeatureToggles enable optional pipeline behaviours.
//
// | Toggle | Description |
// |--------|-------------|
// | `update-pipeline` | Inserts a job that keeps the pipeline/workflow in sync with the halfpipe manifest. Sets BUILD_VERSION. |
// | `update-pipeline-and-tag` | Like update-pipeline, but also tags the git repo with `<PIPELINE_NAME>/v<BUILD_VERSION>`. |
// | `disable-update-pipeline` | Opts out of the `update-pipeline` feature being enabled by default. |
// | `github-statuses` | Updates GitHub commit statuses from Concourse job results (Actions does this by default). |
type FeatureToggles []string

const (
	FeatureUpdatePipeline        = "update-pipeline"
	FeatureUpdatePipelineAndTag  = "update-pipeline-and-tag"
	FeatureDisableUpdatePipeline = "disable-update-pipeline"
	FeatureGithubStatuses        = "github-statuses"
)

var AvailableFeatureToggles = FeatureToggles{
	FeatureUpdatePipeline,
	FeatureUpdatePipelineAndTag,
	FeatureDisableUpdatePipeline,
	FeatureGithubStatuses,
}

func (f FeatureToggles) UpdatePipeline() bool {
	return slices.Contains(f, FeatureUpdatePipeline) || slices.Contains(f, FeatureUpdatePipelineAndTag)
}

func (f FeatureToggles) DisableUpdatePipeline() bool {
	return slices.Contains(f, FeatureDisableUpdatePipeline)
}

func (f FeatureToggles) TagGitRepo() bool {
	return slices.Contains(f, FeatureUpdatePipelineAndTag)
}

func (f FeatureToggles) GithubStatuses() bool {
	return slices.Contains(f, FeatureGithubStatuses)
}
