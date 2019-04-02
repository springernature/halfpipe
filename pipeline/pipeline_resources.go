package pipeline

import (
	"strings"

	"regexp"

	"path"

	"fmt"
	"github.com/concourse/concourse/atc"
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
		Name:   gitName,
		Type:   "git",
		Source: sources,
	}
}

const slackResourceName = "slack-notification"

func (p pipeline) slackResourceType() atc.ResourceType {
	return atc.ResourceType{
		Name: slackResourceName,
		Type: "registry-image",
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
		Type: "registry-image",
		Source: atc.Source{
			"repository": "platformengineering/gcp-resource",
			"tag":        "stable",
		},
	}
}

func (p pipeline) artifactResource(team, pipeline string, artifactConfig manifest.ArtifactConfig) atc.ResourceConfig {
	filter := func(str string) string {
		reg := regexp.MustCompile(`[^a-z0-9\-]+`)
		return reg.ReplaceAllString(strings.ToLower(str), "")
	}

	bucket := config.ArtifactsBucket
	jsonKey := config.ArtifactsJSONKey

	if artifactConfig.Bucket != "" {
		bucket = artifactConfig.Bucket
	}
	if artifactConfig.JSONKey != "" {
		jsonKey = artifactConfig.JSONKey
	}

	return atc.ResourceConfig{
		Name: GenerateArtifactsResourceName(team, pipeline),
		Type: artifactsResourceName,
		Source: atc.Source{
			"bucket":   bucket,
			"folder":   path.Join(filter(team), filter(pipeline)),
			"json_key": jsonKey,
		},
	}
}

func (p pipeline) artifactResourceOnFailure(team, pipeline string, artifactConfig manifest.ArtifactConfig) atc.ResourceConfig {
	config := p.artifactResource(team, pipeline, artifactConfig)
	config.Name = GenerateArtifactsOnFailureResourceName(team, pipeline)
	return config
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
		Name:                 "cron-resource",
		Type:                 "registry-image",
		UniqueVersionHistory: true,
		Source: atc.Source{
			"repository": "cftoolsmiths/cron-resource",
			"tag":        "v0.3",
		},
	}
}

func halfpipeCfDeployResourceType(name string) atc.ResourceType {
	return atc.ResourceType{
		Name: name,
		Type: "registry-image",
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
		Type:   "registry-image",
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
			"json_key": config.VersionJSONKey,
		},
	}
}

func (p pipeline) updateJob(manifest manifest.Manifest) atc.JobConfig {
	job := atc.JobConfig{
		Name: updateJobName,
		Plan: []atc.PlanConfig{{
			Put: versionName,
			Params: atc.Params{
				"bump": "minor",
			},
			Attempts: updateTaskAttempts,
		}},
	}

	if manifest.FeatureToggles.UpdatePipeline() {
		job.Plan = append(job.Plan, p.updatePipelineTask(manifest))
	}

	return job
}

func (p pipeline) updatePipelineTask(man manifest.Manifest) atc.PlanConfig {
	return atc.PlanConfig{
		Task:     updatePipelineName,
		Attempts: updateTaskAttempts,
		TaskConfig: &atc.TaskConfig{
			Platform: "linux",
			Params: map[string]string{
				"CONCOURSE_PASSWORD": "((concourse.password))",
				"CONCOURSE_TEAM":     "((concourse.team))",
				"CONCOURSE_USERNAME": "((concourse.username))",
				"PIPELINE_NAME":      man.PipelineName(),
			},
			ImageResource: p.imageResource(manifest.Docker{
				Image:    "eu.gcr.io/halfpipe-io/halfpipe-auto-update",
				Username: "_json_key",
				Password: "((gcr.private_key))",
			}),
			Run: atc.TaskRunConfig{
				Path: "/bin/update-pipeline",
				Dir:  path.Join(gitDir, man.Repo.BasePath),
			},
			Inputs: []atc.TaskInputConfig{
				{Name: gitName},
				{Name: versionName},
			},
		}}
}
