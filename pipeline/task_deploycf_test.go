package pipeline

import (
	"testing"

	"path/filepath"

	"path"

	cfManifest "code.cloudfoundry.org/cli/util/manifest"
	"github.com/concourse/atc"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestRendersCfDeploy(t *testing.T) {
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

	expectedResourceType := atc.ResourceType{
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

	manifestPath := filepath.Join(gitDir, "manifest-dev.yml")
	testDomain := "dev.cf.private.springer.com"
	envVars := map[string]interface{}{
		"VAR1": "value1",
		"VAR2": "value2",
	}
	expectedDevJob := atc.JobConfig{
		Name:   "deploy-cf",
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{Get: gitDir, Trigger: true},
			atc.PlanConfig{
				Put: expectedDevResource.Name,
				Params: atc.Params{
					"command":      "halfpipe-push",
					"testDomain":   testDomain,
					"manifestPath": manifestPath,
					"vars":         envVars,
					"appPath":      gitDir,
					"gitRefPath":   path.Join(gitDir, ".git", "ref"),
				},
			},
			atc.PlanConfig{
				Put: expectedDevResource.Name,
				Params: atc.Params{
					"command":      "halfpipe-promote",
					"testDomain":   testDomain,
					"manifestPath": manifestPath,
					"vars":         envVars,
					"appPath":      gitDir,
					"gitRefPath":   path.Join(gitDir, ".git", "ref"),
				},
			},
		},
		Ensure: &atc.PlanConfig{
			Put: expectedDevResource.Name,
			Params: atc.Params{
				"command":      "halfpipe-cleanup",
				"testDomain":   testDomain,
				"manifestPath": manifestPath,
				"vars":         envVars,
				"appPath":      gitDir,
				"gitRefPath":   path.Join(gitDir, ".git", "ref"),
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
	liveManifest := filepath.Join(gitDir, "manifest-live.yml")
	expectedLiveJob := atc.JobConfig{
		Name:   "deploy-cf (1)",
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{Get: gitDir, Trigger: true, Passed: []string{"deploy-cf"}},
			atc.PlanConfig{
				Put: expectedLiveResource.Name,
				Params: atc.Params{
					"command":      "halfpipe-push",
					"testDomain":   liveTestDomain,
					"manifestPath": liveManifest,
					"appPath":      gitDir,
					"gitRefPath":   path.Join(gitDir, ".git", "ref"),
				},
			},
			atc.PlanConfig{
				Put: expectedLiveResource.Name,
				Params: atc.Params{
					"command":      "halfpipe-promote",
					"testDomain":   liveTestDomain,
					"manifestPath": liveManifest,
					"appPath":      gitDir,
					"gitRefPath":   path.Join(gitDir, ".git", "ref"),
				},
			},
		},
		Ensure: &atc.PlanConfig{
			Put: expectedLiveResource.Name,
			Params: atc.Params{
				"command":      "halfpipe-cleanup",
				"testDomain":   liveTestDomain,
				"manifestPath": liveManifest,
				"appPath":      gitDir,
				"gitRefPath":   path.Join(gitDir, ".git", "ref"),
			},
		},
	}

	config := testPipeline().Render(man)

	assert.Len(t, config.ResourceTypes, 1)
	assert.Equal(t, expectedResourceType, config.ResourceTypes[0])

	assert.Equal(t, expectedDevResource, config.Resources[1])
	assert.Equal(t, expectedDevJob, config.Jobs[0])

	assert.Equal(t, expectedLiveResource, config.Resources[2])
	assert.Equal(t, expectedLiveJob, config.Jobs[1])
}

func TestRenderPrePromoteTask(t *testing.T) {
	prePromoteRun := manifest.Run{
		Script: "run-script",
		Docker: manifest.Docker{
			Image: "docker-img",
		},
	}

	prePromoteDockerCompose := manifest.DockerCompose{Name: "dock-comp"}

	deployCfTask := manifest.DeployCF{
		API:      "api.dev.cf.springer-sbm.com",
		Space:    "cf-space",
		Org:      "cf-org",
		Manifest: "manifest",
		Vars: manifest.Vars{
			"A": "a",
		},
		DeployArtifact: "artifact.jar",
		PrePromote:     []manifest.Task{prePromoteRun, prePromoteDockerCompose},
	}

	man := manifest.Manifest{Repo: manifest.Repo{URI: "git@github:org/repo-name"}}
	man.Pipeline = "mypipeline"
	man.Tasks = []manifest.Task{deployCfTask}

	cfManifestReader := func(name string) ([]cfManifest.Application, error) {
		return []cfManifest.Application{{
			Name:   name,
			Routes: []string{"route"},
		}}, nil
	}

	pipeline := NewPipeline(cfManifestReader)

	config := pipeline.Render(man)
	plan := config.Jobs[0].Plan

	if assert.Len(t, plan, 6) {
		assert.Equal(t, gitDir, plan[0].Get)
		assert.Equal(t, "artifacts-"+man.Pipeline, plan[1].Get)

		assert.Equal(t, "halfpipe-push", plan[2].Params["command"])

		assert.Equal(t, "run", plan[3].Task)
		assert.Equal(t, "manifest-cf-space-CANDIDATE.dev.cf.private.springer.com", plan[3].TaskConfig.Params["TEST_ROUTE"])
		assert.NotNil(t, plan[3].TaskConfig)

		assert.Equal(t, "run", plan[4].Task)
		assert.Equal(t, "manifest-cf-space-CANDIDATE.dev.cf.private.springer.com", plan[4].TaskConfig.Params["TEST_ROUTE"])
		assert.NotNil(t, plan[4].TaskConfig)

		assert.Equal(t, "halfpipe-promote", plan[5].Params["command"])
	}
	assert.Equal(t, "halfpipe-cleanup", config.Jobs[0].Ensure.Params["command"])

}

func IGNORETestRenderAsSeparateJobsWhenThereIsAPrePromoteTask(t *testing.T) {
	prePromoteTasks := []manifest.Task{
		manifest.Run{
			Script: "run-script",
			Docker: manifest.Docker{
				Image: "docker-img",
			},
		},
		manifest.DockerCompose{Name: "dock-comp"},
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
		PrePromote:     prePromoteTasks,
	}

	man := manifest.Manifest{Repo: manifest.Repo{URI: "git@github:org/repo-name"}}
	man.Pipeline = "mypipeline"
	man.Tasks = []manifest.Task{deployCfTask}

	cfManifestReader := func(name string) ([]cfManifest.Application, error) {
		return []cfManifest.Application{
			{
				Name:   name,
				Routes: []string{"route"},
			},
		}, nil
	}

	pipeline := NewPipeline(cfManifestReader)
	config := pipeline.Render(man)

	//assert.Len(t, config.Jobs, 3, "should be split into 3 jobs")

	//push
	planPush := config.Jobs[0].Plan
	assert.Equal(t, gitDir, planPush[0].Get)
	assert.Equal(t, "artifacts-"+man.Pipeline, planPush[1].Get)
	assert.Equal(t, "halfpipe-push", planPush[2].Params["command"])

	//pre promote
	planPrePromote := config.Jobs[1].Plan
	assert.Equal(t, "run", planPrePromote[0].Task)
	assert.Equal(t, "manifest-cf-space-CANDIDATE.dev.cf.private.springer.com", planPrePromote[0].TaskConfig.Params)
	assert.NotNil(t, planPrePromote[0].TaskConfig)

	assert.Equal(t, "run", planPrePromote[1].Task)
	assert.Equal(t, "manifest-cf-space-CANDIDATE.dev.cf.private.springer.com", planPrePromote[1].TaskConfig.Params)
	assert.NotNil(t, planPrePromote[1].TaskConfig)

	//promote
	planPromote := config.Jobs[2].Plan
	assert.Equal(t, "halfpipe-promote", planPromote[0].Params["command"])
	assert.Equal(t, "halfpipe-cleanup", planPromote[1].Params["command"])

}
