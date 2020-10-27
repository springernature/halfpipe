package linters_test

import (
	linters "github.com/springernature/halfpipe/linters/actions"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

var linter = linters.ActionsLinter{}.Lint

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

	actual := linter(man)
	assert.Empty(t, actual.Errors)
	assert.Len(t, actual.Warnings, 8)
}
