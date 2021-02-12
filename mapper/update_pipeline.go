package mapper

import (
	"github.com/springernature/halfpipe/manifest"
)

type updatePipeline struct{}

func NewUpdatePipelineMapper() updatePipeline {
	return updatePipeline{}
}

func (up updatePipeline) Apply(original manifest.Manifest) (updated manifest.Manifest, err error) {
	updated = original

	updateTask := manifest.Update{
		TagRepo: original.FeatureToggles.TagGitRepo(),
	}

	if (original.Platform.IsConcourse() && original.FeatureToggles.UpdatePipeline()) ||
		(original.Platform.IsActions() && original.FeatureToggles.TagGitRepo()) {
		updated.Tasks = append(manifest.TaskList{updateTask}, updated.Tasks...)
	}

	return updated, nil
}
