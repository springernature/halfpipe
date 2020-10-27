package linters_test

import (
	linters "github.com/springernature/halfpipe/linters/actions"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

var lint = linters.ActionsLinter{}.Lint

func TestActionsLinter_UnsupportedTasks(t *testing.T) {
	man := manifest.Manifest{
		Tasks: manifest.TaskList{
			manifest.ConsumerIntegrationTest{},
			manifest.DeployCF{},
			manifest.DeployMLModules{},
			manifest.DeployMLZip{},
			manifest.DockerCompose{},
			manifest.DockerPush{},
			manifest.Parallel{},
			manifest.Run{},
			manifest.Sequence{},
		},
	}

	actual := lint(man)
	assert.Empty(t, actual.Errors)
	assert.Len(t, actual.Warnings, 8)
}

func TestActionsLinter_UnsupportedTriggers(t *testing.T) {
	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.DockerTrigger{},
			manifest.GitTrigger{},
			manifest.PipelineTrigger{},
			manifest.TimerTrigger{},
		},
	}

	actual := lint(man)
	assert.Empty(t, actual.Errors)
	assert.Len(t, actual.Warnings, 2)
}

func TestActionsLinter_UnsupportedGitTriggerOptions(t *testing.T) {
	man := manifest.Manifest{
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
