package linters_test

import (
	"github.com/springernature/halfpipe/linters"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

var lint = linters.ActionsLinter{}.Lint

func TestActionsLinter_UnsupportedTriggers(t *testing.T) {
	man := manifest.Manifest{
		Platform: "actions",
		Triggers: manifest.TriggerList{
			manifest.DockerTrigger{},
			manifest.GitTrigger{},
			manifest.PipelineTrigger{},
			manifest.TimerTrigger{},
		},
	}

	actual := lint(man)
	assert.Empty(t, actual.Errors)
	assert.Len(t, actual.Warnings, 1)
	assert.Contains(t, actual.Warnings[0].Error(), "PipelineTrigger")
}

func TestActionsLinter_UnsupportedGitTriggerOptions(t *testing.T) {
	man := manifest.Manifest{
		Platform: "actions",
		Triggers: manifest.TriggerList{
			manifest.GitTrigger{
				URI:           "uri",
				PrivateKey:    "key",
				WatchedPaths:  []string{"watch"},
				IgnoredPaths:  []string{"ignore"},
				GitCryptKey:   "key",
				Branch:        "branch",
				Shallow:       false,
				ManualTrigger: true,
			},
		},
	}

	actual := lint(man)
	assert.Empty(t, actual.Errors)
	assert.Len(t, actual.Warnings, 3)
}

func TestActionsLinter_UnsupportedTaskOptions(t *testing.T) {
	man := manifest.Manifest{
		Platform: "actions",
		Tasks: manifest.TaskList{
			manifest.DockerPush{ManualTrigger: true},
			manifest.Run{ManualTrigger: true},
			manifest.DeployCF{
				ManualTrigger: true,
				PrePromote:    manifest.TaskList{manifest.Run{}},
				Rolling:       true,
			},
		},
	}
	actual := lint(man)
	assert.Empty(t, actual.Errors)
	if assert.Len(t, actual.Warnings, 4) {
		assert.Contains(t, actual.Warnings[0].Error(), "manual_trigger")
		assert.Contains(t, actual.Warnings[1].Error(), "manual_trigger")
		assert.Contains(t, actual.Warnings[2].Error(), "manual_trigger")
		assert.Contains(t, actual.Warnings[3].Error(), "rolling")
	}
}

func TestActionsLinter_PreventCircularTriggers(t *testing.T) {
	man := manifest.Manifest{
		Platform: "actions",
		Triggers: manifest.TriggerList{
			manifest.DockerTrigger{
				Image: "the-same-image",
			},
		},
		Tasks: manifest.TaskList{
			manifest.DockerPush{
				Image: "the-same-image",
			},
		},
	}

	actual := lint(man)
	assert.Empty(t, actual.Errors)
	assert.Len(t, actual.Warnings, 1)
	assert.Contains(t, actual.Warnings[0].Error(), "loop")
}

func TestActionsFeatures_WarnAboutUpdatePipelineNotImplemented(t *testing.T) {
	tests := map[string]manifest.FeatureToggles{
		"all features": {
			manifest.FeatureUpdatePipeline,
			manifest.FeatureUpdatePipelineAndTag,
			manifest.FeatureDockerDecompose,
		},
		"update-pipeline":         {manifest.FeatureUpdatePipeline},
		"update-pipeline-and-tag": {manifest.FeatureUpdatePipelineAndTag},
	}

	for name, features := range tests {
		t.Run(name, func(t *testing.T) {
			actual := lint(manifest.Manifest{Platform: "actions", FeatureToggles: features})
			assert.Empty(t, actual.Errors)
			assert.Len(t, actual.Warnings, 1)
			assert.Equal(t, actual.Warnings[0], linters.ErrUpdatePiplineNotImplemented)
		})
	}

}
