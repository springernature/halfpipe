package pipeline

import (
	"strings"

	"path/filepath"
	"regexp"

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

	return atc.ResourceConfig{
		Name:   gitDir,
		Type:   "git",
		Source: sources,
	}
}

func (p pipeline) slackResourceType() atc.ResourceType {
	return atc.ResourceType{
		Name: "slack-notification",
		Type: "docker-image",
		Source: atc.Source{
			"repository": "cfcommunity/slack-notification-resource",
			"tag":        "latest",
		},
	}
}

func (p pipeline) slackResource() atc.ResourceConfig {
	return atc.ResourceConfig{
		Name: "slack",
		Type: "slack-notification",
		Source: atc.Source{
			"url": config.SlackWebhook,
		},
	}
}

func (p pipeline) gcpResourceType() atc.ResourceType {
	return atc.ResourceType{
		Name: "gcp-resource",
		Type: "docker-image",
		Source: atc.Source{
			"repository": "platformengineering/gcp-resource",
			"tag":        "0.14.0",
		},
	}
}

func (p pipeline) gcpResource(team, pipeline string) atc.ResourceConfig {
	filter := func(str string) string {
		reg := regexp.MustCompile(`[^a-z0-9\-]+`)
		return reg.ReplaceAllString(strings.ToLower(str), "")
	}

	return atc.ResourceConfig{
		Name: GenerateArtifactsFolderName(team, pipeline),
		Type: "gcp-resource",
		Source: atc.Source{
			"json_key": "((gcr.private_key))",
			"folder":   filepath.Join(filter(team), filter(pipeline)),
			"bucket":   "halfpipe-artifacts",
		},
	}
}

func (p pipeline) timerResource(interval string) atc.ResourceConfig {
	return atc.ResourceConfig{
		Name:   "timer " + interval,
		Type:   "time",
		Source: atc.Source{"interval": interval},
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
		"api":      deployCF.API,
		"org":      deployCF.Org,
		"space":    deployCF.Space,
		"username": deployCF.Username,
		"password": deployCF.Password,
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
