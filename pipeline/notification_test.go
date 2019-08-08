package pipeline

import (
	"testing"

	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestRendersSlackResourceWithoutOnFailureTask(t *testing.T) {
	slackChannel := "#ee-re"

	man := manifest.Manifest{}
	man.SlackChannel = slackChannel

	pipeline := testPipeline().Render(man)

	resourceType, found := pipeline.ResourceTypes.Lookup(slackResourceName)
	assert.True(t, found)

	resource, found := pipeline.Resources.Lookup(slackResourceName)
	assert.True(t, found)

	assert.Equal(t, slackResourceName, resource.Name)
	assert.Equal(t, config.SlackWebhook, resource.Source["url"])
	assert.Equal(t, "slack", resource.Type)
	assert.Equal(t, "registry-image", resourceType.Type)
	assert.Equal(t, "cfcommunity/slack-notification-resource", resourceType.Source["repository"])
	assert.Equal(t, "slack", resourceType.Name)
}

func TestRendersSlackOnFailurePlan(t *testing.T) {
	slackChannel := "#ee-re"

	man := manifest.Manifest{Repo: manifest.Repo{URI: "git@github.com:foo/reponame"}}
	man.SlackChannel = slackChannel
	man.Tasks = []manifest.Task{
		manifest.DeployCF{},
		manifest.Run{},
	}
	pipeline := testPipeline().Render(man)

	channel := (pipeline.Jobs[0].Failure.InParallel.Steps)[0].Params["channel"]
	channel1 := (pipeline.Jobs[1].Failure.InParallel.Steps)[0].Params["channel"]

	assert.Equal(t, slackChannel, channel)
	assert.Equal(t, slackChannel, channel1)
}

func TestRendersSlackOnFailurePlanWithArtifactOnFailure(t *testing.T) {
	slackChannel := "#ee-re"

	man := manifest.Manifest{Repo: manifest.Repo{URI: "git@github.com:foo/reponame"}}
	man.SlackChannel = slackChannel
	man.Tasks = []manifest.Task{
		manifest.DeployCF{},
		manifest.Run{
			SaveArtifactsOnFailure: []string{"test-reports"},
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

	t.Run("top level task", func(t *testing.T) {
		man := manifest.Manifest{
			SlackChannel: slackChannel,
			Tasks: manifest.TaskList{
				manifest.Run{Name: taskName1, NotifyOnSuccess: true},
				manifest.DockerCompose{Name: taskName2},
				manifest.DockerPush{Name: taskName3, NotifyOnSuccess: true},
			},
		}

		pipeline := testPipeline().Render(man)

		firstTask, _ := pipeline.Jobs.Lookup(taskName1)
		assert.Equal(t, (firstTask.Success.InParallel.Steps)[0], slackOnSuccessPlan(slackChannel))

		secondTask, _ := pipeline.Jobs.Lookup(taskName2)
		assert.Nil(t, secondTask.Success)

		thirdTask, _ := pipeline.Jobs.Lookup(taskName3)
		assert.Equal(t, (thirdTask.Success.InParallel.Steps)[0], slackOnSuccessPlan(slackChannel))
	})

	t.Run("pre promote task", func(t *testing.T) {
		man := manifest.Manifest{
			SlackChannel: slackChannel,
			Tasks: manifest.TaskList{
				manifest.Run{Name: taskName1, NotifyOnSuccess: true},
				manifest.DockerCompose{Name: taskName2},
				manifest.DeployCF{
					Name: taskName3,
					PrePromote: manifest.TaskList{
						manifest.Run{NotifyOnSuccess: true},
					},
				},
			},
		}

		pipeline := testPipeline().Render(man)

		firstTask, _ := pipeline.Jobs.Lookup(taskName1)
		assert.Equal(t, (firstTask.Success.InParallel.Steps)[0], slackOnSuccessPlan(slackChannel))

		secondTask, _ := pipeline.Jobs.Lookup(taskName2)
		assert.Nil(t, secondTask.Success)

		thirdTask, _ := pipeline.Jobs.Lookup(taskName3)
		assert.Equal(t, (thirdTask.Success.InParallel.Steps)[0], slackOnSuccessPlan(slackChannel))
	})

}
