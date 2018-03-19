package pipeline

import (
	"testing"

	"github.com/concourse/atc"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestRendersCfDeployResources(t *testing.T) {
	taskDeployDev := manifest.DeployCF{
		API:      "dev-api",
		Space:    "dev",
		Org:      "springer",
		Username: "rob",
		Password: "supersecret",
		Manifest: "manifest-dev.yml",
		Vars: manifest.Vars{
			"VAR1": "value1",
			"VAR2": "value2",
		},
	}
	taskDeployLive := manifest.DeployCF{
		API:      "live-api",
		Space:    "prod",
		Org:      "springer",
		Username: "rob",
		Password: "supersecret",
		Manifest: "manifest-live.yml",
	}

	man := manifest.Manifest{Repo: manifest.Repo{URI: "git@github.com:foo/reponame"}}
	man.Tasks = []manifest.Task{taskDeployDev, taskDeployLive}

	expectedResourceConfig := atc.ResourceType{
		Name: "cf-resource",
		Type: "docker-image",
		Source: atc.Source{
			"repository": "platformengineering/cf-resource",
			"tag":        "stable",
		},
	}

	expectedDevResource := atc.ResourceConfig{
		Name: deployCFResourceName(taskDeployDev),
		Type: "cf-resource",
		Source: atc.Source{
			"api":      "dev-api",
			"space":    "dev",
			"org":      "springer",
			"password": "supersecret",
			"username": "rob",
		},
	}

	expectedDevJob := atc.JobConfig{
		Name:   "deploy-cf",
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{Get: "reponame", Trigger: true},
			atc.PlanConfig{
				Put: expectedDevResource.Name,
				Params: atc.Params{
					"command":      "halfpipe-push",
					"manifestPath": "reponame/manifest-dev.yml",
					"vars": map[string]interface{}{
						"VAR1": "value1",
						"VAR2": "value2",
					},
					"appPath": "reponame",
				},
			},
			atc.PlanConfig{
				Put: expectedDevResource.Name,
				Params: atc.Params{
					"command":      "halfpipe-promote",
					"manifestPath": "reponame/manifest-dev.yml",
					"vars": map[string]interface{}{
						"VAR1": "value1",
						"VAR2": "value2",
					},
					"appPath": "reponame",
				},
			},
			atc.PlanConfig{
				Put: expectedDevResource.Name,
				Params: atc.Params{
					"command":      "halfpipe-delete",
					"manifestPath": "reponame/manifest-dev.yml",
					"vars": map[string]interface{}{
						"VAR1": "value1",
						"VAR2": "value2",
					},
					"appPath": "reponame",
				},
			},
		},
	}

	expectedLiveResource := atc.ResourceConfig{
		Name: deployCFResourceName(taskDeployLive),
		Type: "cf-resource",
		Source: atc.Source{
			"api":      "live-api",
			"space":    "prod",
			"org":      "springer",
			"password": "supersecret",
			"username": "rob",
		},
	}

	expectedLiveJob := atc.JobConfig{
		Name:   "deploy-cf (1)",
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{Get: "reponame", Trigger: true, Passed: []string{"deploy-cf"}},
			atc.PlanConfig{
				Put: expectedLiveResource.Name,
				Params: atc.Params{
					"command":      "halfpipe-push",
					"manifestPath": "reponame/manifest-live.yml",
					"appPath":      "reponame",
				},
			},
			atc.PlanConfig{
				Put: expectedLiveResource.Name,
				Params: atc.Params{
					"command":      "halfpipe-promote",
					"manifestPath": "reponame/manifest-live.yml",
					"appPath":      "reponame",
				},
			},
			atc.PlanConfig{
				Put: expectedLiveResource.Name,
				Params: atc.Params{
					"command":      "halfpipe-delete",
					"manifestPath": "reponame/manifest-live.yml",
					"appPath":      "reponame",
				},
			},
		},
	}

	config := testPipeline().Render(man)

	assert.Len(t, config.ResourceTypes, 1)
	assert.Equal(t, expectedResourceConfig, config.ResourceTypes[0])

	assert.Equal(t, expectedDevResource, config.Resources[1])
	assert.Equal(t, expectedDevJob, config.Jobs[0])

	assert.Equal(t, expectedLiveResource, config.Resources[2])
	assert.Equal(t, expectedLiveJob, config.Jobs[1])
}
