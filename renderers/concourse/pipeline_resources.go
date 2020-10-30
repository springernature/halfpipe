package concourse

import (
	"strings"

	"regexp"

	"path"

	"fmt"
	"github.com/concourse/concourse/atc"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
)

const longResourceCheckInterval = "24h"
const webHookAssistedResourceCheckInterval = "10m"

func (p pipeline) gitResource(trigger manifest.GitTrigger) atc.ResourceConfig {
	sources := atc.Source{
		"uri": trigger.URI,
	}

	if trigger.PrivateKey != "" {
		sources["private_key"] = trigger.PrivateKey
	}

	if len(trigger.WatchedPaths) > 0 {
		sources["paths"] = trigger.WatchedPaths
	}

	if len(trigger.IgnoredPaths) > 0 {
		sources["ignore_paths"] = trigger.IgnoredPaths
	}

	if trigger.GitCryptKey != "" {
		sources["git_crypt_key"] = trigger.GitCryptKey
	}

	sources["branch"] = trigger.Branch

	cfg := atc.ResourceConfig{
		Name:   trigger.GetTriggerName(),
		Type:   "git",
		Source: sources,
	}

	if strings.HasPrefix(trigger.URI, config.WebHookAssistedGitPrefix) {
		cfg.CheckEvery = webHookAssistedResourceCheckInterval
	}

	return cfg
}

const slackResourceName = "slack"
const slackResourceTypeName = "halfpipe-slack-resource"

func (p pipeline) slackResourceType() atc.ResourceType {
	return atc.ResourceType{
		Name:       slackResourceTypeName,
		Type:       "registry-image",
		CheckEvery: longResourceCheckInterval,
		Source: atc.Source{
			"repository": config.DockerRegistry + "halfpipe-slack-resource",
			"tag":        "latest",
			"password":   "((halfpipe-gcr.private_key))",
			"username":   "_json_key",
		},
	}
}

func (p pipeline) slackResource() atc.ResourceConfig {
	return atc.ResourceConfig{
		Name:       slackResourceName,
		Type:       slackResourceTypeName,
		CheckEvery: longResourceCheckInterval,
		Source: atc.Source{
			"token": config.SlackToken,
		},
	}
}

func (p pipeline) gcpResourceType() atc.ResourceType {
	return atc.ResourceType{
		Name: artifactsResourceName,
		Type: "registry-image",
		Source: atc.Source{
			"repository": config.DockerRegistry + "gcp-resource",
			"tag":        "stable",
			"password":   "((halfpipe-gcr.private_key))",
			"username":   "_json_key",
		},
	}
}

func (p pipeline) artifactResource(man manifest.Manifest) atc.ResourceConfig {
	filter := func(str string) string {
		reg := regexp.MustCompile(`[^a-z0-9\-]+`)
		return reg.ReplaceAllString(strings.ToLower(str), "")
	}

	bucket := config.ArtifactsBucket
	jsonKey := config.ArtifactsJSONKey

	if man.ArtifactConfig.Bucket != "" {
		bucket = man.ArtifactConfig.Bucket
	}
	if man.ArtifactConfig.JSONKey != "" {
		jsonKey = man.ArtifactConfig.JSONKey
	}

	return atc.ResourceConfig{
		Name:       artifactsName,
		Type:       artifactsResourceName,
		CheckEvery: longResourceCheckInterval,
		Source: atc.Source{
			"bucket":   bucket,
			"folder":   path.Join(filter(man.Team), filter(man.PipelineName())),
			"json_key": jsonKey,
		},
	}
}

func (p pipeline) artifactResourceOnFailure(man manifest.Manifest) atc.ResourceConfig {
	config := p.artifactResource(man)
	config.Name = artifactsOnFailureName
	return config
}

func (p pipeline) cronResource(trigger manifest.TimerTrigger) atc.ResourceConfig {
	return atc.ResourceConfig{
		Name:       trigger.GetTriggerName(),
		Type:       cronResourceTypeName,
		CheckEvery: "1m",
		Source: atc.Source{
			"expression":       trigger.Cron,
			"location":         "UTC",
			"fire_immediately": true,
		},
	}
}

const cronResourceTypeName = "halfpipe-cron-resource"

func cronResourceType() atc.ResourceType {

	return atc.ResourceType{
		Name:                 cronResourceTypeName,
		Type:                 "registry-image",
		UniqueVersionHistory: true,
		Source: atc.Source{
			"repository": config.DockerRegistry + cronResourceTypeName,
			"password":   "((halfpipe-gcr.private_key))",
			"username":   "_json_key",
			"tag":        "stable",
		},
	}
}

func halfpipePipelineTriggerResourceType() atc.ResourceType {
	return atc.ResourceType{
		Name: "halfpipe-pipeline-trigger",
		Type: "registry-image",
		Source: atc.Source{
			"repository": config.DockerRegistry + "halfpipe-pipeline-trigger-resource",
			"password":   "((halfpipe-gcr.private_key))",
			"username":   "_json_key",
		},
	}
}

const deployCfResourceTypeName = "cf-resource"

func (p pipeline) halfpipeCfDeployResourceType() atc.ResourceType {
	image := strings.Join([]string{deployCfResourceTypeName, "v2"}, "-")
	fullPath := path.Join(config.DockerRegistry + image)
	return atc.ResourceType{
		Name: deployCfResourceTypeName,
		Type: "registry-image",
		Source: atc.Source{
			"repository": fullPath,
			"tag":        "stable",
			"password":   "((halfpipe-gcr.private_key))",
			"username":   "_json_key",
		},
	}
}

func (p pipeline) pipelineTriggerResource(pipelineTrigger manifest.PipelineTrigger) atc.ResourceConfig {
	sources := atc.Source{
		"concourse_url": pipelineTrigger.ConcourseURL,
		"username":      pipelineTrigger.Username,
		"password":      pipelineTrigger.Password,
		"team":          pipelineTrigger.Team,
		"pipeline":      pipelineTrigger.Pipeline,
		"job":           pipelineTrigger.Job,
		"status":        pipelineTrigger.Status,
	}

	return atc.ResourceConfig{
		Name:   pipelineTrigger.GetTriggerName(),
		Type:   "halfpipe-pipeline-trigger",
		Source: sources,
	}
}

func (p pipeline) deployCFResource(deployCF manifest.DeployCF, resourceName string) atc.ResourceConfig {
	resourceType := deployCfResourceTypeName

	sources := atc.Source{
		"api":      deployCF.API,
		"org":      deployCF.Org,
		"space":    deployCF.Space,
		"username": deployCF.Username,
		"password": deployCF.Password,
	}

	return atc.ResourceConfig{
		Name:       resourceName,
		Type:       resourceType,
		Source:     sources,
		CheckEvery: longResourceCheckInterval,
	}
}

func (p pipeline) dockerPushResource(docker manifest.DockerPush) atc.ResourceConfig {
	return atc.ResourceConfig{
		Name: manifest.DockerTrigger{Image: docker.Image}.GetTriggerName(),
		Type: "docker-image",
		Source: atc.Source{
			"username":   docker.Username,
			"password":   docker.Password,
			"repository": docker.Image,
		},
		CheckEvery: longResourceCheckInterval,
	}
}

func (p pipeline) dockerTriggerResource(trigger manifest.DockerTrigger) atc.ResourceConfig {
	config := atc.ResourceConfig{
		Name: trigger.GetTriggerName(),
		Type: "docker-image",
		Source: atc.Source{
			"repository": trigger.Image,
		},
	}

	if trigger.Username != "" && trigger.Password != "" {
		config.Source["username"] = trigger.Username
		config.Source["password"] = trigger.Password
	}

	return config
}

func (p pipeline) imageResource(docker manifest.Docker) *atc.ImageResource {
	repo, tag := docker.Image, "latest"
	if strings.Contains(docker.Image, ":") {
		split := strings.Split(docker.Image, ":")
		repo = split[0]
		tag = split[1]
	}

	source := atc.Source{
		"repository": repo,
		"tag":        tag,
		"registry_mirror": map[string]string{
			"host": "eu-mirror.gcr.io",
		},
	}

	if docker.Username != "" && docker.Password != "" {
		source["username"] = docker.Username
		source["password"] = docker.Password
	}

	return &atc.ImageResource{
		Type:   "registry-image",
		Source: source,
	}
}

func (p pipeline) versionResource(manifest manifest.Manifest) atc.ResourceConfig {
	key := fmt.Sprintf("%s-%s", manifest.Team, manifest.Pipeline)
	gitTrigger := manifest.Triggers.GetGitTrigger()
	if gitTrigger.Branch != "master" && gitTrigger.Branch != "main" {
		key = fmt.Sprintf("%s-%s", key, gitTrigger.Branch)
	}

	return atc.ResourceConfig{
		Name: versionName,
		Type: "semver",
		Source: atc.Source{
			"driver":   "gcs",
			"key":      key,
			"bucket":   config.VersionBucket,
			"json_key": config.VersionJSONKey,
		},
	}
}
