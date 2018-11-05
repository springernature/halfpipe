package pipeline

import (
	"strings"

	"regexp"

	"path"

	"fmt"
	"github.com/concourse/atc"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
)

func (p pipeline) gitResource(repo manifest.Repo) atc.ResourceConfig {
	sources := atc.Source{
		"uri": repo.URI,
	}

	if repo.PrivateKey != "" {
		sources["private_key"] = repo.PrivateKey
	}

	if len(repo.WatchedPaths) > 0 {
		sources["paths"] = repo.WatchedPaths
	}

	if len(repo.IgnoredPaths) > 0 {
		sources["ignore_paths"] = repo.IgnoredPaths
	}

	if repo.GitCryptKey != "" {
		sources["git_crypt_key"] = repo.GitCryptKey
	}

	if repo.Branch == "" {
		sources["branch"] = "master"
	} else {
		sources["branch"] = repo.Branch
	}

	return atc.ResourceConfig{
		Name:   gitDir,
		Type:   "git",
		Source: sources,
	}
}

const slackResourceName = "slack-notification"

func (p pipeline) slackResourceType() atc.ResourceType {
	return atc.ResourceType{
		Name: slackResourceName,
		Type: "docker-image",
		Source: atc.Source{
			"repository": "cfcommunity/slack-notification-resource",
			"tag":        "v1.4.2",
		},
	}
}

func (p pipeline) slackResource() atc.ResourceConfig {
	return atc.ResourceConfig{
		Name: slackResourceName,
		Type: slackResourceName,
		Source: atc.Source{
			"url": config.SlackWebhook,
		},
	}
}

func (p pipeline) gcpResourceType() atc.ResourceType {
	return atc.ResourceType{
		Name: artifactsResourceName,
		Type: "docker-image",
		Source: atc.Source{
			"repository": "platformengineering/gcp-resource",
			"tag":        "stable",
		},
	}
}

func (p pipeline) gcpResource(team, pipeline string, artifactConfig manifest.ArtifactConfig) atc.ResourceConfig {
	filter := func(str string) string {
		reg := regexp.MustCompile(`[^a-z0-9\-]+`)
		return reg.ReplaceAllString(strings.ToLower(str), "")
	}

	bucket := "halfpipe-io-artifacts"
	json_key := "((gcr.private_key))"

	if artifactConfig.Bucket != "" {
		bucket = artifactConfig.Bucket
	}
	if artifactConfig.JsonKey != "" {
		json_key = artifactConfig.JsonKey
	}

	return atc.ResourceConfig{
		Name: GenerateArtifactsResourceName(team, pipeline),
		Type: artifactsResourceName,
		Source: atc.Source{
			"bucket":   bucket,
			"folder":   path.Join(filter(team), filter(pipeline)),
			"json_key": json_key,
		},
	}
}

func (p pipeline) timerResource(interval string) atc.ResourceConfig {
	return atc.ResourceConfig{
		Name:   timerName,
		Type:   "time",
		Source: atc.Source{"interval": interval},
	}
}

func (p pipeline) cronResource(expression string) atc.ResourceConfig {
	return atc.ResourceConfig{
		Name: cronName,
		Type: "cron-resource",
		Source: atc.Source{
			"expression":       expression,
			"location":         "UTC",
			"fire_immediately": true,
		},
	}
}

func cronResourceType() atc.ResourceType {
	return atc.ResourceType{
		Name: "cron-resource",
		Type: "docker-image",
		Source: atc.Source{
			"repository":       "cftoolsmiths/cron-resource",
			"tag":              "v0.3",
			"fire_immediately": true,
		},
	}
}

func halfpipeCfDeployResourceType(name string) atc.ResourceType {
	return atc.ResourceType{
		Name: name,
		Type: "docker-image",
		Source: atc.Source{
			"repository": "platformengineering/cf-resource",
			"tag":        "stable",
		},
	}
}

func (p pipeline) deployCFResource(deployCF manifest.DeployCF, resourceName string) atc.ResourceConfig {
	sources := atc.Source{
		"api":                  deployCF.API,
		"org":                  deployCF.Org,
		"space":                deployCF.Space,
		"username":             deployCF.Username,
		"password":             deployCF.Password,
		"prometheusGatewayURL": config.PrometheusGatewayURL,
	}

	return atc.ResourceConfig{
		Name:   resourceName,
		Type:   "cf-resource",
		Source: sources,
	}
}

func (p pipeline) dockerPushResource(docker manifest.DockerPush, resourceName string) atc.ResourceConfig {
	return atc.ResourceConfig{
		Name: resourceName,
		Type: "docker-image",
		Source: atc.Source{
			"username":   docker.Username,
			"password":   docker.Password,
			"repository": docker.Image,
		},
	}
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
	}

	if docker.Username != "" && docker.Password != "" {
		source["username"] = docker.Username
		source["password"] = docker.Password
	}

	return &atc.ImageResource{
		Type:   "docker-image",
		Source: source,
	}
}

func (p pipeline) versionResource(manifest manifest.Manifest) atc.ResourceConfig {
	key := fmt.Sprintf("%s-%s", manifest.Team, manifest.Pipeline)
	if manifest.Repo.Branch != "" && manifest.Repo.Branch != "master" {
		key = fmt.Sprintf("%s-%s", key, manifest.Repo.Branch)
	}

	return atc.ResourceConfig{
		Name: versionName,
		Type: "semver",
		Source: atc.Source{
			"driver":   "gcs",
			"key":      key,
			"bucket":   config.VersionBucket,
			"json_key": config.VersionJsonKey,
		},
	}
}

func (p pipeline) versionUpdateJob(manifest manifest.Manifest) atc.JobConfig {
	return atc.JobConfig{
		Name: "update version",
		Plan: atc.PlanSequence{
			atc.PlanConfig{
				Put: versionName,
				Params: atc.Params{
					"bump": "minor",
				},
			},
		},
	}
}
