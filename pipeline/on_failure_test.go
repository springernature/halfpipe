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
	assert.Len(t, pipeline.Resources, 2)

	assert.Equal(t, "slack", pipeline.Resources[1].Name)
	assert.Equal(t, config.SlackWebhook, pipeline.Resources[1].Source["url"])
	assert.Equal(t, "slack-notification", pipeline.Resources[1].Type)
	assert.Equal(t, "docker-image", pipeline.ResourceTypes[0].Type)
	assert.Equal(t, "cfcommunity/slack-notification-resource", pipeline.ResourceTypes[0].Source["repository"])
	assert.Equal(t, "slack-notification", pipeline.ResourceTypes[0].Name)

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

	channel := pipeline.Jobs[0].Failure.Params["channel"]
	channel1 := pipeline.Jobs[1].Failure.Params["channel"]

	assert.Equal(t, slackChannel, channel)
	assert.Equal(t, slackChannel, channel1)
}

func TestDoesntRenderWhenNotSet(t *testing.T) {
	slackChannel := ""

	man := manifest.Manifest{}
	man.SlackChannel = slackChannel

	assert.Len(t, testPipeline().Render(man).Resources, 1)
}
