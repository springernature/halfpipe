package pipeline

import (
	"fmt"
	"testing"

	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestTriggerName(t *testing.T) {
	t.Run("without tag", func(t *testing.T) {
		t.Run("with hostname and simple chars", func(t *testing.T) {
			expectedName := "my-cool-image"
			trigger := manifest.DockerTrigger{
				Image: "eu.gcr.io/halfpipe-io/my-cool-image",
			}
			assert.Equal(t, expectedName, trigger.GetTriggerName())
		})

		t.Run("with org and simple chars", func(t *testing.T) {
			expectedName := "my-cool-image"
			trigger := manifest.DockerTrigger{
				Image: "springernature/my-cool-image",
			}
			assert.Equal(t, expectedName, trigger.GetTriggerName())
		})

		t.Run("with some underscores", func(t *testing.T) {
			expectedName := "my-mega-cool-image"
			trigger := manifest.DockerTrigger{
				Image: "my_mega_cool-image",
			}
			assert.Equal(t, expectedName, trigger.GetTriggerName())
		})
	})

	t.Run("with tag", func(t *testing.T) {
		t.Run("with hostname and simple chars", func(t *testing.T) {
			expectedName := "my-cool-image.my-tag"
			trigger := manifest.DockerTrigger{
				Image: "eu.gcr.io/halfpipe-io/my-cool-image:my-tag",
			}
			assert.Equal(t, expectedName, trigger.GetTriggerName())
		})

		t.Run("with org and simple chars", func(t *testing.T) {
			expectedName := "my-cool-image.my-tag"
			trigger := manifest.DockerTrigger{
				Image: "springernature/my-cool-image.my-tag",
			}
			assert.Equal(t, expectedName, trigger.GetTriggerName())
		})

		t.Run("with some underscores", func(t *testing.T) {
			expectedName := "my-mega-cool-image.my-tag-yeah"
			trigger := manifest.DockerTrigger{
				Image: "my_mega_cool-image:my_tag-yeah",
			}
			assert.Equal(t, expectedName, trigger.GetTriggerName())
		})
	})
}

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

	config := testPipeline().Render(man)
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
	config := testPipeline().Render(man)

	fmt.Println(ToString(config))

	t1 := config.Jobs[0]
	t1InParallel := t1.Plan[0].InParallel.Steps

	assert.Len(t, t1.Plan, 2)
	assert.Equal(t, trigger.GetTriggerName(), t1InParallel[0].Name())
	assert.True(t, t1InParallel[0].Trigger)

	t2 := config.Jobs[1]
	t2InParallel := t2.Plan[0].InParallel.Steps
	assert.Len(t, t2.Plan, 2)

	assert.Equal(t, trigger.GetTriggerName(), t2InParallel[0].Name())
	assert.Equal(t, []string{t1.Name}, t2InParallel[0].Passed)
}
