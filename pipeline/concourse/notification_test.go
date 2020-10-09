package concourse

import (
	"testing"

	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestRendersSlackOnFailurePlan(t *testing.T) {
	slackChannel := "#ee-re"

	man := manifest.Manifest{}
	man.Tasks = []manifest.Task{
		manifest.DeployCF{
			Notifications: manifest.Notifications{
				OnFailure: []string{slackChannel},
			},
		},
		manifest.Run{
			Notifications: manifest.Notifications{
				OnFailure: []string{slackChannel},
			},
		},
	}
	pipeline := testPipeline().Render(man)

	channel := (pipeline.Jobs[0].Failure.InParallel.Steps)[0].Params["channel"]
	channel1 := (pipeline.Jobs[1].Failure.InParallel.Steps)[0].Params["channel"]

	assert.Equal(t, slackChannel, channel)
	assert.Equal(t, slackChannel, channel1)
}

func TestRendersSlackOnFailurePlanWithArtifactOnFailure(t *testing.T) {
	slackChannel := "#ee-re"

	man := manifest.Manifest{}
	man.Tasks = []manifest.Task{
		manifest.DeployCF{
			Notifications: manifest.Notifications{
				OnFailure: []string{slackChannel},
			},
		},
		manifest.Run{
			SaveArtifactsOnFailure: []string{"test-reports"},
			Notifications: manifest.Notifications{
				OnFailure: []string{slackChannel},
			},
		},
	}
	pipeline := testPipeline().Render(man)

	assert.Equal(t, slackResourceName, (pipeline.Jobs[0].Failure.InParallel.Steps)[0].Put)
	assert.Equal(t, slackChannel, (pipeline.Jobs[0].Failure.InParallel.Steps)[0].Params["channel"])

	assert.Equal(t, artifactsOnFailureName, (pipeline.Jobs[1].Failure.InParallel.Steps)[0].Put)
	assert.Equal(t, slackResourceName, (pipeline.Jobs[1].Failure.InParallel.Steps)[1].Put)
	assert.Equal(t, slackChannel, (pipeline.Jobs[1].Failure.InParallel.Steps)[1].Params["channel"])
}

func TestDoesntRenderWhenNotSet(t *testing.T) {
	man := manifest.Manifest{}
	man.SlackChannel = ""

	pipeline := testPipeline().Render(man)
	_, foundResource := pipeline.Resources.Lookup(slackResourceName)
	assert.False(t, foundResource)

	_, foundResourceType := pipeline.ResourceTypes.Lookup(slackResourceName)
	assert.False(t, foundResourceType)
}

func TestAddsSlackNotificationOnSuccess(t *testing.T) {
	slackChannel := "#yay"
	taskName1 := "task1"
	taskName2 := "task2"
	taskName3 := "task3"

	withoutSuccess := manifest.Notifications{
		OnFailure: []string{slackChannel},
	}
	withSuccess := manifest.Notifications{
		OnFailure: []string{slackChannel},
		OnSuccess: []string{slackChannel},
	}

	t.Run("top level task", func(t *testing.T) {
		man := manifest.Manifest{
			Tasks: manifest.TaskList{
				manifest.Run{Name: taskName1, NotifyOnSuccess: true, Notifications: withSuccess},
				manifest.DockerCompose{Name: taskName2, Notifications: withoutSuccess},
				manifest.DockerPush{Name: taskName3, NotifyOnSuccess: true, Notifications: withSuccess},
			},
		}

		pipeline := testPipeline().Render(man)

		firstTask, _ := pipeline.Jobs.Lookup(taskName1)
		assert.Equal(t, (firstTask.Success.InParallel.Steps)[0], slackOnSuccessPlan(slackChannel, ""))

		secondTask, _ := pipeline.Jobs.Lookup(taskName2)
		assert.Nil(t, secondTask.Success)

		thirdTask, _ := pipeline.Jobs.Lookup(taskName3)
		assert.Equal(t, (thirdTask.Success.InParallel.Steps)[0], slackOnSuccessPlan(slackChannel, ""))
	})
}
