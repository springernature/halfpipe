package defaults

import "github.com/springernature/halfpipe/manifest"

type topLevelDefaulter struct{}

func (o topLevelDefaulter) Apply(original manifest.Manifest) (updated manifest.Manifest) {
	updated = original

	if updated.Platform == "" {
		updated.Platform = "concourse"
	}

	if updated.PipelineId == "" {
		updated.PipelineId = updated.Pipeline
	}

	return
}

func newTopLevelDefaulter() topLevelDefaulter {
	return topLevelDefaulter{}
}
