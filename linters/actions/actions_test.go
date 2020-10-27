package linters_test

import (
	linters "github.com/springernature/halfpipe/linters/actions"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"strings"
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
	assert.Len(t, actual.Warnings, 3)
}

func TestActionsLinter_ManualGitTrigger(t *testing.T) {
	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.GitTrigger{
				ManualTrigger: true,
			},
		},
	}

	actual := lint(man)
	assert.Empty(t, actual.Errors)
	assert.Len(t, actual.Warnings, 1)
	assert.True(t, strings.Contains(actual.Warnings[0].Error(), "manual_trigger"))
}
