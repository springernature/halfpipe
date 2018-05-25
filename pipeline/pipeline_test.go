package pipeline

import (
	"testing"

	cfManifest "code.cloudfoundry.org/cli/util/manifest"

	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func testPipeline() pipeline {
	return pipeline{
		rManifest: func(s string) ([]cfManifest.Application, error) {
			return []cfManifest.Application{
				{
					Name:   "test-name",
					Routes: []string{"test-route"},
				},
			}, nil
		},
	}
}

func TestRenderWithTriggerTrueAndPassedOnPreviousTask(t *testing.T) {
	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.Run{Script: "asd.sh"},
			manifest.DeployCF{ManualTrigger: true},
			manifest.DockerPush{},
		},
	}
	config := testPipeline().Render(man)

	assert.Nil(t, config.Jobs[0].Plan[0].Passed)
	assert.Equal(t, config.Jobs[0].Plan[0].Trigger, true)

	assert.Equal(t, config.Jobs[1].Plan[0].Passed[0], config.Jobs[0].Name)
	assert.Equal(t, config.Jobs[1].Plan[0].Trigger, false)

	assert.Equal(t, config.Jobs[2].Plan[0].Passed[0], config.Jobs[1].Name)
	assert.Equal(t, config.Jobs[2].Plan[0].Trigger, true)
}