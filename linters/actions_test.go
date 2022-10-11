package linters

import (
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

var lint = ActionsLinter{}.Lint

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

	errs := lint(man).Issues
	assertContainsError(t, errs, ErrUnsupportedPipelineTrigger)
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

	errs := lint(man).Issues
	assertContainsError(t, errs, ErrUnsupportedGitPrivateKey)
	assertContainsError(t, errs, ErrUnsupportedGitUri)
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
	errs := lint(man).Issues

	if assert.Len(t, errs, 4) {
		assert.Contains(t, errs[0].Error(), "manual_trigger")
		assert.Contains(t, errs[1].Error(), "manual_trigger")
		assert.Contains(t, errs[2].Error(), "manual_trigger")
		assert.Contains(t, errs[3].Error(), "rolling")
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

	errs := lint(man).Issues
	assertContainsError(t, errs, ErrDockerTriggerLoop)
}

func TestActionsFeatures_WarnAboutUpdatePipelineNotImplemented(t *testing.T) {
	tests := map[string]manifest.FeatureToggles{
		"all features": {
			manifest.FeatureUpdatePipeline,
			manifest.FeatureUpdatePipelineAndTag,
		},
		"update-pipeline":         {manifest.FeatureUpdatePipeline},
		"update-pipeline-and-tag": {manifest.FeatureUpdatePipelineAndTag},
	}

	for name, features := range tests {
		t.Run(name, func(t *testing.T) {
			errs := lint(manifest.Manifest{Platform: "actions", FeatureToggles: features}).Issues
			assertContainsError(t, errs, ErrUnsupportedUpdatePipeline)
		})
	}

}

func TestActionsLinter_UnsupportedUseCovenant(t *testing.T) {
	man := manifest.Manifest{
		Platform: "actions",
		Tasks: manifest.TaskList{
			manifest.ConsumerIntegrationTest{
				UseCovenant: true,
			},
		},
	}
	errs := lint(man).Issues
	assertContainsError(t, errs, ErrUnsupportedCovenant)
}
