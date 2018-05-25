package pipeline

import (
	"testing"

	"path"

	cfManifest "code.cloudfoundry.org/cli/util/manifest"
	"github.com/concourse/atc"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestRendersCfDeploy(t *testing.T) {
	liveAPI := "api.live.com"
	liveTestDomain := "test.live.com"

	devAPI := "api.dev.com"
	devTestDomain := "test.dev.com"

	taskDeployDev := manifest.DeployCF{
		API:        devAPI,
		Space:      "dev",
		Org:        "springer",
		Username:   "rob",
		Password:   "supersecret",
		TestDomain: devTestDomain,
		Manifest:   "manifest-dev.yml",
		Vars: manifest.Vars{
			"VAR1": "value1",
			"VAR2": "value2",
		},
	}
	taskDeployLive := manifest.DeployCF{
		API:        liveAPI,
		Space:      "prod",
		Org:        "springer",
		TestDomain: liveTestDomain,
		Username:   "rob",
		Password:   "supersecret",
		Manifest:   "manifest-live.yml",
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
			"api":      devAPI,
			"space":    "dev",
			"org":      "springer",
			"password": "supersecret",
			"username": "rob",
		},
	}

	manifestPath := path.Join(gitDir, "manifest-dev.yml")

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
					"testDomain":   devTestDomain,
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
					"testDomain":   devTestDomain,
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
				"testDomain":   devTestDomain,
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
			"api":      liveAPI,
			"space":    "prod",
			"org":      "springer",
			"password": "supersecret",
			"username": "rob",
		},
	}

	liveManifest := path.Join(gitDir, "manifest-live.yml")
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

func TestRenderAsSeparateJobsWhenThereIsAPrePromoteTask(t *testing.T) {
	dockerComposeTaskBefore := manifest.DockerCompose{Name: "dc-before"}
	dockerComposeTaskAfter := manifest.DockerCompose{Name: "dc-after"}
	testDomain := "test.domain.com"

	deployCfTask := manifest.DeployCF{
		Name:       "deploy to dev",
		API:        "api.dev.cf.springer-sbm.com",
		Space:      "cf-space",
		Org:        "cf-org",
		TestDomain: testDomain,
		Manifest:   "manifest",
		Vars: manifest.Vars{
			"A": "a",
		},
		DeployArtifact: "artifact.jar",
		PrePromote: []manifest.Task{
			manifest.Run{
				Name:   "pp1",
				Script: "run-script",
				Docker: manifest.Docker{
					Image: "docker-img",
				},
			},
			manifest.DockerCompose{Name: "pp2"},
		},
	}

	man := manifest.Manifest{Repo: manifest.Repo{URI: "git@github:org/repo-name"}}
	man.Pipeline = "mypipeline"
	man.Tasks = []manifest.Task{dockerComposeTaskBefore, deployCfTask, dockerComposeTaskAfter}

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

	assert.Len(t, config.Jobs, 6, "should be 6 jobs")

	//docker-compose before
	assert.Equal(t, dockerComposeTaskBefore.Name, config.Jobs[0].Name)

	//push
	push := config.Jobs[1]
	assert.Equal(t, "deploy to dev - candidate", push.Name)
	assert.Equal(t, gitDir, push.Plan[0].Get)
	assert.Equal(t, config.Jobs[0].Name, push.Plan[0].Passed[0])
	assert.Equal(t, "artifacts-"+man.Pipeline, push.Plan[1].Get)
	assert.Equal(t, "halfpipe-push", push.Plan[2].Params["command"])

	//pre promote 1
	pp1 := config.Jobs[2]
	assert.Equal(t, "pp1", pp1.Name)
	assert.Equal(t, gitDir, pp1.Plan[0].Get)
	assert.Equal(t, push.Name, pp1.Plan[0].Passed[0])
	assert.Equal(t, "run", pp1.Plan[1].Task)
	assert.Equal(t, "manifest-cf-space-CANDIDATE.test.domain.com", pp1.Plan[1].TaskConfig.Params["TEST_ROUTE"])
	assert.NotNil(t, pp1.Plan[1].TaskConfig)

	//pre promote 2
	pp2 := config.Jobs[3]
	assert.Equal(t, "pp2", pp2.Name)
	assert.Equal(t, gitDir, pp2.Plan[0].Get)
	assert.Equal(t, push.Name, pp2.Plan[0].Passed[0])
	assert.Equal(t, "run", pp2.Plan[1].Task)
	assert.Equal(t, "manifest-cf-space-CANDIDATE.test.domain.com", pp2.Plan[1].TaskConfig.Params["TEST_ROUTE"])
	assert.NotNil(t, pp2.Plan[1].TaskConfig)

	//promote
	promote := config.Jobs[4]
	assert.Equal(t, gitDir, promote.Plan[0].Get)
	assert.Equal(t, []string{pp1.Name, pp2.Name}, promote.Plan[0].Passed)
	assert.Equal(t, "deploy to dev - promote", promote.Name)
	assert.Equal(t, "halfpipe-promote", promote.Plan[1].Params["command"])

	assert.Equal(t, "halfpipe-cleanup", promote.Ensure.Params["command"])

	//docker-compose after
	dockerComposeAfter := config.Jobs[5]
	assert.Equal(t, dockerComposeTaskAfter.Name, dockerComposeAfter.Name)
	assert.Equal(t, []string{promote.Name}, dockerComposeAfter.Plan[0].Passed)
}
