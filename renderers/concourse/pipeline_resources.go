package concourse

import (
	"strings"
	"time"

	"regexp"

	"path"

	"fmt"

	"github.com/concourse/concourse/atc"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
)

var halfpipeDockerImage = manifest.Docker{
	Image:      path.Join(config.DockerRegistry, "engineering-enablement", "halfpipe-docker:latest"),
	Username:   "oauth2accesstoken",
	Password:   "((gcp:platform-gar/token.token))",
	Entrypoint: "docker.sh",
}

var halfpipeDockerComposeImage = manifest.Docker{
	Image:      path.Join(config.DockerRegistry, "halfpipe-docker-compose:stable"),
	Username:   "_json_key",
	Password:   "((halfpipe-gcr.private_key))",
	Entrypoint: "docker.sh",
}

var longResourceCheckInterval = atc.CheckEvery{
	Interval: 24 * time.Hour,
}
var shortResourceCheckInterval = atc.CheckEvery{
	Interval: 10 * time.Minute,
}

func (c Concourse) gitResource(trigger manifest.GitTrigger) atc.ResourceConfig {
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
		cfg.CheckEvery = &shortResourceCheckInterval
	}

	return cfg
}

const slackResourceName = "slack"
const slackResourceTypeName = "halfpipe-slack-resource"

func (c Concourse) slackResourceType() atc.ResourceType {
	return atc.ResourceType{
		Name:       slackResourceTypeName,
		Type:       "registry-image",
		CheckEvery: &longResourceCheckInterval,
		Source: atc.Source{
			"repository": path.Join(config.DockerRegistry, "halfpipe-slack-resource"),
			"tag":        "latest",
			"password":   "((gcp:platform-gar/token.token))",
			"username":   "oauth2accesstoken",
		},
	}
}

func (c Concourse) slackResource() atc.ResourceConfig {
	return atc.ResourceConfig{
		Name:       slackResourceName,
		Type:       slackResourceTypeName,
		CheckEvery: &longResourceCheckInterval,
		Source: atc.Source{
			"token": config.SlackToken,
		},
	}
}

const teamsResourceName = "teams"
const teamsResourceTypeName = "halfpipe-teams-resource"

func (c Concourse) teamsResourceType() atc.ResourceType {
	return atc.ResourceType{
		Name:       teamsResourceTypeName,
		Type:       "registry-image",
		CheckEvery: &longResourceCheckInterval,
		Source: atc.Source{
			"repository": path.Join(config.DockerRegistry, "halfpipe-teams-resource"),
			"tag":        "latest",
			"password":   "((gcp:platform-gar/token.token))",
			"username":   "oauth2accesstoken",
		},
	}
}

func (c Concourse) teamsResource() atc.ResourceConfig {
	return atc.ResourceConfig{
		Name:       teamsResourceName,
		Type:       teamsResourceTypeName,
		CheckEvery: &longResourceCheckInterval,
	}
}

func (c Concourse) gcpResourceType() atc.ResourceType {
	return atc.ResourceType{
		Name:       artifactsResourceName,
		Type:       "registry-image",
		CheckEvery: &longResourceCheckInterval,
		Source: atc.Source{
			"repository": path.Join(config.DockerRegistry, "engineering-enablement", "gcp-resource"),
			"tag":        "latest",
			"password":   "((gcp:platform-gar/token.token))",
			"username":   "oauth2accesstoken",
		},
	}
}

func (c Concourse) artifactResource(man manifest.Manifest) atc.ResourceConfig {
	filter := func(str string) string {
		reg := regexp.MustCompile(`[^a-z0-9\-]+`)
		return reg.ReplaceAllString(strings.ToLower(str), "")
	}

	return atc.ResourceConfig{
		Name:       artifactsName,
		Type:       artifactsResourceName,
		CheckEvery: &longResourceCheckInterval,
		Source: atc.Source{
			"bucket": config.ArtifactsBucket,
			"folder": path.Join(filter(man.Team), filter(man.PipelineName())),
			"token":  "((gcp:platform-artifacts/token.token))",
		},
	}
}

func (c Concourse) artifactResourceOnFailure(man manifest.Manifest) atc.ResourceConfig {
	config := c.artifactResource(man)
	config.Name = artifactsOnFailureName
	return config
}

func (c Concourse) cronResource(trigger manifest.TimerTrigger) atc.ResourceConfig {
	return atc.ResourceConfig{
		Name:       trigger.GetTriggerName(),
		Type:       cronResourceTypeName,
		CheckEvery: &shortResourceCheckInterval,
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
		Name:       cronResourceTypeName,
		Type:       "registry-image",
		CheckEvery: &longResourceCheckInterval,
		Source: atc.Source{
			"repository": "platformengineering/halfpipe-cron-resource",
			"tag":        "sha256:cf4e7dd0acaa914e1202b846548d6478d2bb1a178f71484766ca9dfc027468a7",
		},
	}
}

func halfpipePipelineTriggerResourceType() atc.ResourceType {
	return atc.ResourceType{
		Name:       "halfpipe-pipeline-trigger",
		Type:       "registry-image",
		CheckEvery: &longResourceCheckInterval,
		Source: atc.Source{
			"repository": path.Join(config.DockerRegistry, "halfpipe-pipeline-trigger-resource"),
			"password":   "((gcp:platform-gar/token.token))",
			"username":   "oauth2accesstoken",
		},
	}
}

const deployCfResourceTypeName = "cf-resource"

func (c Concourse) halfpipeCfDeployResourceType() atc.ResourceType {
	image := strings.Join([]string{deployCfResourceTypeName, "v2"}, "-")
	fullPath := path.Join(config.DockerRegistry, image)
	return atc.ResourceType{
		Name:       deployCfResourceTypeName,
		Type:       "registry-image",
		CheckEvery: &longResourceCheckInterval,
		Source: atc.Source{
			"repository": fullPath,
			"password":   "((gcp:platform-gar/token.token))",
			"username":   "oauth2accesstoken",
		},
	}
}

func (c Concourse) pipelineTriggerResource(pipelineTrigger manifest.PipelineTrigger) atc.ResourceConfig {
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

func (c Concourse) deployCFResource(deployCF manifest.DeployCF, resourceName string) atc.ResourceConfig {
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
		CheckEvery: &longResourceCheckInterval,
	}
}

func (c Concourse) dockerTriggerResource(trigger manifest.DockerTrigger) atc.ResourceConfig {
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

func (c Concourse) imageResource(docker manifest.Docker) *atc.ImageResource {
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

func (c Concourse) versionResource(manifest manifest.Manifest) atc.ResourceConfig {
	key := fmt.Sprintf("%s-%s", manifest.Team, manifest.Pipeline)
	gitTrigger := manifest.Triggers.GetGitTrigger()
	if gitTrigger.Branch != "master" && gitTrigger.Branch != "main" {
		key = fmt.Sprintf("%s-%s", key, gitTrigger.Branch)
	}

	return atc.ResourceConfig{
		Name:       versionName,
		Type:       "semver",
		CheckEvery: &longResourceCheckInterval,
		Source: atc.Source{
			"driver":   "gcs",
			"key":      key,
			"bucket":   config.VersionBucket,
			"json_key": config.VersionJSONKey,
		},
	}
}

const githubStatusesResourceName = "github-statuses"
const githubStatusesResourceTypeName = "halfpipe-github-statuses-resource"

func (c Concourse) githubStatusesResourceType() atc.ResourceType {
	return atc.ResourceType{
		Name:       githubStatusesResourceTypeName,
		Type:       "registry-image",
		CheckEvery: &longResourceCheckInterval,
		Source: atc.Source{
			"repository": path.Join(config.DockerRegistry, "engineering-enablement/github-status-resource"),
			"password":   "((gcp:platform-gar/token.token))",
			"username":   "oauth2accesstoken",
		},
	}
}

func (c Concourse) githubStatusesResource(manifest manifest.Manifest) atc.ResourceConfig {
	return atc.ResourceConfig{
		Name:       githubStatusesResourceName,
		Type:       githubStatusesResourceTypeName,
		CheckEvery: &longResourceCheckInterval,
		Source: atc.Source{
			"repo":         fmt.Sprintf("%s/%s", config.GithubOrg, manifest.Triggers.GetGitTrigger().GetRepoName()),
			"access_token": config.GithubToken,
		},
	}
}
