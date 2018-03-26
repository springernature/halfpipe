package pipeline

import (
	"testing"

	cfManifest "code.cloudfoundry.org/cli/util/manifest"
	"github.com/concourse/atc"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestRendersCfDeployResources(t *testing.T) {
	taskDeployDev := manifest.DeployCF{
		API:      "http://api.dev.cf.springer-sbm.com",
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
		API:      "http://api.live.cf.springer-sbm.com",
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
			"api":      "http://api.dev.cf.springer-sbm.com",
			"space":    "dev",
			"org":      "springer",
			"password": "supersecret",
			"username": "rob",
		},
	}

	manifestPath := "reponame/manifest-dev.yml"
	testDomain := "dev.cf.private.springer.com"
	envVars := map[string]interface{}{
		"VAR1": "value1",
		"VAR2": "value2",
	}
	repoName := "reponame"
	expectedDevJob := atc.JobConfig{
		Name:   "deploy-cf",
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{Get: repoName, Trigger: true},
			atc.PlanConfig{
				Put: expectedDevResource.Name,
				Params: atc.Params{
					"command":      "halfpipe-push",
					"testDomain":   testDomain,
					"manifestPath": manifestPath,
					"vars":         envVars,
					"appPath":      repoName,
				},
			},
			atc.PlanConfig{
				Put: expectedDevResource.Name,
				Params: atc.Params{
					"command":      "halfpipe-promote",
					"testDomain":   testDomain,
					"manifestPath": manifestPath,
					"vars":         envVars,
					"appPath":      repoName,
				},
			},
			atc.PlanConfig{
				Put: expectedDevResource.Name,
				Params: atc.Params{
					"command":      "halfpipe-delete",
					"testDomain":   testDomain,
					"manifestPath": manifestPath,
					"vars":         envVars,
					"appPath":      repoName,
				},
			},
		},
	}

	expectedLiveResource := atc.ResourceConfig{
		Name: deployCFResourceName(taskDeployLive),
		Type: "cf-resource",
		Source: atc.Source{
			"api":      "http://api.live.cf.springer-sbm.com",
			"space":    "prod",
			"org":      "springer",
			"password": "supersecret",
			"username": "rob",
		},
	}

	liveTestDomain := "live.cf.private.springer.com"
	liveManifest := "reponame/manifest-live.yml"
	expectedLiveJob := atc.JobConfig{
		Name:   "deploy-cf (1)",
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{Get: repoName, Trigger: true, Passed: []string{"deploy-cf"}},
			atc.PlanConfig{
				Put: expectedLiveResource.Name,
				Params: atc.Params{
					"command":      "halfpipe-push",
					"testDomain":   liveTestDomain,
					"manifestPath": liveManifest,
					"appPath":      repoName,
				},
			},
			atc.PlanConfig{
				Put: expectedLiveResource.Name,
				Params: atc.Params{
					"command":      "halfpipe-promote",
					"testDomain":   liveTestDomain,
					"manifestPath": liveManifest,
					"appPath":      repoName,
				},
			},
			atc.PlanConfig{
				Put: expectedLiveResource.Name,
				Params: atc.Params{
					"command":      "halfpipe-delete",
					"testDomain":   liveTestDomain,
					"manifestPath": liveManifest,
					"appPath":      repoName,
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

func TestRenderPrePromoteTask(t *testing.T) {
	prePromoteTask := manifest.Run{
		Script: "run-script",
		Docker: manifest.Docker{
			Image: "docker-img",
		},
	}

	deployCfTask := manifest.DeployCF{
		API:      "api.dev.cf.springer-sbm.com",
		Space:    "cf-space",
		Org:      "cf-org",
		Manifest: "manifest",
		Vars: manifest.Vars{
			"A": "a",
		},
		DeployArtifact: "artifact.jar",
		PrePromote:     []manifest.Task{prePromoteTask},
	}

	man := manifest.Manifest{Repo: manifest.Repo{URI: "git@github:org/repo-name"}}
	man.Pipeline = "mypipeline"
	man.Tasks = []manifest.Task{deployCfTask}

	readerGivesOneApp := func(name string) ([]cfManifest.Application, error) {
		return []cfManifest.Application{
			{
				Name:   name,
				Routes: []string{"route"},
			},
		}, nil
	}

	pipeline := NewPipeline(readerGivesOneApp)

	config := pipeline.Render(man)
	plan := config.Jobs[0].Plan

	if assert.Len(t, plan, 6) {
		assert.Equal(t, "repo-name", plan[0].Get)
		assert.Equal(t, "artifacts-"+man.Pipeline, plan[1].Get)

		assert.Equal(t, "halfpipe-push", plan[2].Params["command"])

		assert.Equal(t, "run", plan[3].Task)
		expectedVars := map[string]string{
			"TEST_ROUTE": "manifest-CANDIDATE.dev.cf.private.springer.com",
		}
		assert.Equal(t, expectedVars, plan[3].TaskConfig.Params)
		assert.NotNil(t, plan[3].TaskConfig)

		assert.Equal(t, "halfpipe-promote", plan[4].Params["command"])
		assert.Equal(t, "halfpipe-delete", plan[5].Params["command"])
	}
}
