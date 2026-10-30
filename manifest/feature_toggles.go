package manifest

import "slices"

// FeatureToggles enable optional pipeline behaviours.
//
// | Toggle | Description |
// |--------|-------------|
// | `update-pipeline-and-tag` | Like update-pipeline, but also tags the git repo with `<PIPELINE_NAME>/v<BUILD_VERSION>`. |
// | `disable-update-pipeline` | Opts out of the `update-pipeline` feature being enabled by default. |
// | `github-statuses` | Updates GitHub commit statuses from Concourse job results (Actions does this by default). |
// | `disable-automatic-notifications` | Opts out of the default failure notification to the team's Teams channel. |
type FeatureToggles []string

const (
	FeatureUpdatePipelineAndTag          = "update-pipeline-and-tag"
	FeatureDisableUpdatePipeline         = "disable-update-pipeline"
	FeatureGithubStatuses                = "github-statuses"
	FeatureDisableAutomaticNotifications = "disable-automatic-notifications"
)

var AvailableFeatureToggles = FeatureToggles{
	FeatureUpdatePipelineAndTag,
	FeatureDisableUpdatePipeline,
	FeatureGithubStatuses,
	FeatureDisableAutomaticNotifications,
}

func (f FeatureToggles) UpdatePipeline() bool {
	return !slices.Contains(f, FeatureDisableUpdatePipeline)
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

func (f FeatureToggles) AutomaticNotifications() bool {
	return !slices.Contains(f, FeatureDisableAutomaticNotifications)
}
