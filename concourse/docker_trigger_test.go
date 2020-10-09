package concourse

import (
	"fmt"
	"testing"

	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestDockerTriggerSetAddsResource(t *testing.T) {
	trigger := manifest.DockerTrigger{
		Image: "myUser/ubuntu-with-somedeps",
	}
	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			trigger,
		},
		Tasks: []manifest.Task{
			manifest.Run{Script: "run.sh"},
		},
	}

	config := testPipeline().RenderAtcConfig(man)
	resource, found := config.Resources.Lookup(trigger.GetTriggerName())
	assert.True(t, found)
	assert.Equal(t, trigger.GetTriggerName(), resource.Name)
	assert.Equal(t, "docker-image", resource.Type)
	assert.Equal(t, man.Triggers[0].(manifest.DockerTrigger).Image, resource.Source["repository"])
}

func TestDockerTriggerSetWithCorrectPassedOnSecondJob(t *testing.T) {
	trigger := manifest.DockerTrigger{
		Image: "myUser/ubuntu-with-somedeps",
	}
	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			trigger,
		},
		Tasks: []manifest.Task{
			manifest.Run{Script: "s1.sh"},
			manifest.Run{Script: "s2.sh"},
		},
	}
	config := testPipeline().RenderAtcConfig(man)

	fmt.Println(ToString(config))

	t1 := config.Jobs[0].Plan
	t1InParallel := t1[0].InParallel.Steps

	assert.Len(t, t1, 2)
	assert.Equal(t, trigger.GetTriggerName(), t1InParallel[0].Name())
	assert.True(t, t1InParallel[0].Trigger)

	t2 := config.Jobs[1].Plan
	t2InParallel := t2[0].InParallel.Steps
	assert.Len(t, t2, 2)

	assert.Equal(t, trigger.GetTriggerName(), t2InParallel[0].Name())
	assert.Equal(t, []string{t1[1].Task}, t2InParallel[0].Passed)
}
