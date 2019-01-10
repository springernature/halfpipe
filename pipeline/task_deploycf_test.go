package pipeline

import (
	"testing"

	"path"

	"fmt"

	cfManifest "code.cloudfoundry.org/cli/util/manifest"
	"github.com/cloudfoundry/bosh-cli/director/template"
	"github.com/concourse/atc"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/config"
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
		Timeout:    "5m",
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
			"api":                  devAPI,
			"space":                "dev",
			"org":                  "springer",
			"password":             "supersecret",
			"prometheusGatewayURL": config.PrometheusGatewayURL,
			"username":             "rob",
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
			atc.PlanConfig{Aggregate: &atc.PlanSequence{atc.PlanConfig{Get: gitDir, Trigger: true}}},
			atc.PlanConfig{
				Put:      "cf halfpipe-push",
				Attempts: 2,
				Resource: expectedDevResource.Name,
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
				Put:      "cf halfpipe-promote",
				Attempts: 2,
				Resource: expectedDevResource.Name,
				Params: atc.Params{
					"command":      "halfpipe-promote",
					"testDomain":   devTestDomain,
					"manifestPath": manifestPath,
				},
			},
		},
		Ensure: &atc.PlanConfig{
			Put:      "cf halfpipe-cleanup",
			Attempts: 2,
			Resource: expectedDevResource.Name,
			Params: atc.Params{
				"command":      "halfpipe-cleanup",
				"manifestPath": manifestPath,
			},
		},
	}

	expectedLiveResource := atc.ResourceConfig{
		Name: deployCFResourceName(taskDeployLive),
		Type: "cf-resource",
		Source: atc.Source{
			"api":                  liveAPI,
			"space":                "prod",
			"org":                  "springer",
			"password":             "supersecret",
			"prometheusGatewayURL": config.PrometheusGatewayURL,
			"username":             "rob",
		},
	}

	liveManifest := path.Join(gitDir, "manifest-live.yml")
	expectedLiveJob := atc.JobConfig{
		Name:   "deploy-cf (1)",
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{Aggregate: &atc.PlanSequence{atc.PlanConfig{Get: gitDir, Trigger: true, Passed: []string{"deploy-cf"}}}},
			atc.PlanConfig{
				Put:      "cf halfpipe-push",
				Attempts: 2,
				Resource: expectedLiveResource.Name,
				Params: atc.Params{
					"command":      "halfpipe-push",
					"testDomain":   liveTestDomain,
					"manifestPath": liveManifest,
					"appPath":      gitDir,
					"gitRefPath":   path.Join(gitDir, ".git", "ref"),
					"timeout":      "5m",
				},
			},
			atc.PlanConfig{
				Put:      "cf halfpipe-promote",
				Attempts: 2,
				Resource: expectedLiveResource.Name,
				Params: atc.Params{
					"command":      "halfpipe-promote",
					"testDomain":   liveTestDomain,
					"manifestPath": liveManifest,
					"timeout":      "5m",
				},
			},
		},
		Ensure: &atc.PlanConfig{
			Put:      "cf halfpipe-cleanup",
			Attempts: 2,
			Resource: expectedLiveResource.Name,
			Params: atc.Params{
				"command":      "halfpipe-cleanup",
				"manifestPath": liveManifest,
				"timeout":      "5m",
			},
		},
	}

	cfg := testPipeline().Render(man)

	assert.Len(t, cfg.ResourceTypes, 1)
	assert.Equal(t, expectedResourceType, cfg.ResourceTypes[0])

	assert.Equal(t, expectedDevResource, cfg.Resources[1])
	assert.Equal(t, expectedDevJob, cfg.Jobs[0])

	assert.Equal(t, expectedLiveResource, cfg.Resources[2])
	assert.Equal(t, expectedLiveJob, cfg.Jobs[1])
}

func TestRenderWithPrePromoteTasks(t *testing.T) {
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
				Vars: manifest.Vars{
					"PP1": "pp1",
				},
			},
			manifest.DockerCompose{Name: "pp2"},
		},
	}

	man := manifest.Manifest{Repo: manifest.Repo{URI: "git@github:org/repo-name"}}
	man.Team = "myteam"
	man.Pipeline = "mypipeline"
	man.Tasks = []manifest.Task{dockerComposeTaskBefore, deployCfTask, dockerComposeTaskAfter}

	cfManifestReader := func(pathToManifest string, pathsToVarsFiles []string, vars []template.VarKV) ([]cfManifest.Application, error) {
		return []cfManifest.Application{
			{
				Name:   "app",
				Routes: []string{"route"},
			},
		}, nil
	}

	pipeline := NewPipeline(cfManifestReader, afero.Afero{Fs: afero.NewMemMapFs()})
	cfg := pipeline.Render(man)

	assert.Len(t, cfg.Jobs, 3, "should be 3 jobs")

	//docker-compose before
	assert.Equal(t, dockerComposeTaskBefore.Name, cfg.Jobs[0].Name)

	//deploy-cf
	deployJob := cfg.Jobs[1]
	assert.Equal(t, "deploy to dev", deployJob.Name)

	plan := deployJob.Plan

	//halfpipe-push
	assert.Equal(t, gitName, (*plan[0].Aggregate)[0].Get)
	assert.Equal(t, cfg.Jobs[0].Name, (*plan[0].Aggregate)[0].Passed[0])
	assert.Equal(t, restoreArtifactTask(man), plan[1])
	assert.Equal(t, "halfpipe-push", plan[2].Params["command"])

	//pre promote 1
	pp1 := (*(*plan[3].Aggregate)[0].Do)[0]
	assert.Equal(t, "pp1", pp1.Task)
	assert.Equal(t, "app-cf-space-CANDIDATE.test.domain.com", pp1.TaskConfig.Params["TEST_ROUTE"])
	assert.Equal(t, "pp1", pp1.TaskConfig.Params["PP1"])

	//pre promote 2
	pp2 := (*(*plan[3].Aggregate)[1].Do)[0]
	assert.Equal(t, "pp2", pp2.Task)
	assert.Equal(t, "app-cf-space-CANDIDATE.test.domain.com", pp2.TaskConfig.Params["TEST_ROUTE"])

	//halfpipe-promote
	assert.Equal(t, "halfpipe-promote", plan[4].Params["command"])

	//halfpipe-cleanup (ensure)
	assert.Equal(t, "halfpipe-cleanup", deployJob.Ensure.Params["command"])

	//docker-compose after
	dockerComposeAfter := cfg.Jobs[2]
	assert.Equal(t, dockerComposeTaskAfter.Name, dockerComposeAfter.Name)
	assert.Equal(t, []string{deployJob.Name}, (*dockerComposeAfter.Plan[0].Aggregate)[0].Passed)
}

func TestRenderWithPrePromoteTasksWhenSavingAndRestoringArtifacts(t *testing.T) {
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
				Vars: manifest.Vars{
					"PP1": "pp1",
				},
				SaveArtifacts: []string{"build"},
			},
			manifest.Run{
				Name:   "pp2",
				Script: "run-script",
				Docker: manifest.Docker{
					Image: "docker-img",
				},
				Vars: manifest.Vars{
					"PP2": "pp2",
				},
				RestoreArtifacts: true,
			},
		},
	}

	man := manifest.Manifest{Repo: manifest.Repo{URI: "git@github:org/repo-name"}}
	man.Team = "myteam"
	man.Pipeline = "mypipeline"
	man.Tasks = []manifest.Task{dockerComposeTaskBefore, deployCfTask, dockerComposeTaskAfter}

	cfManifestReader := func(pathToManifest string, pathsToVarsFiles []string, vars []template.VarKV) ([]cfManifest.Application, error) {
		return []cfManifest.Application{
			{
				Name:   "app",
				Routes: []string{"route"},
			},
		}, nil
	}

	pipeline := NewPipeline(cfManifestReader, afero.Afero{Fs: afero.NewMemMapFs()})
	cfg := pipeline.Render(man)

	assert.Len(t, cfg.Jobs, 3, "should be 3 jobs")

	//docker-compose before
	assert.Equal(t, dockerComposeTaskBefore.Name, cfg.Jobs[0].Name)

	//deploy-cf
	deployJob := cfg.Jobs[1]
	assert.Equal(t, "deploy to dev", deployJob.Name)

	plan := deployJob.Plan

	//halfpipe-push
	assert.Equal(t, gitName, (*plan[0].Aggregate)[0].Get)
	assert.Equal(t, cfg.Jobs[0].Name, (*plan[0].Aggregate)[0].Passed[0])
	assert.Equal(t, restoreArtifactTask(man), plan[1])
	assert.Equal(t, "halfpipe-push", plan[2].Params["command"])

	//pre promote 1
	pp1 := (*plan[3].Do)[0]
	assert.Equal(t, "pp1", pp1.Task)
	assert.Equal(t, "app-cf-space-CANDIDATE.test.domain.com", pp1.TaskConfig.Params["TEST_ROUTE"])
	assert.Equal(t, "pp1", pp1.TaskConfig.Params["PP1"])

	//pre promote 2
	pp2 := (*plan[4].Do)[0]
	assert.Equal(t, "pp2", pp2.Task)
	assert.Equal(t, "app-cf-space-CANDIDATE.test.domain.com", pp2.TaskConfig.Params["TEST_ROUTE"])
	assert.Equal(t, "pp2", pp2.TaskConfig.Params["PP2"])

	//halfpipe-promote
	assert.Equal(t, "halfpipe-promote", plan[5].Params["command"])

	//halfpipe-cleanup (ensure)
	assert.Equal(t, "halfpipe-cleanup", deployJob.Ensure.Params["command"])

	//docker-compose after
	dockerComposeAfter := cfg.Jobs[2]
	assert.Equal(t, dockerComposeTaskAfter.Name, dockerComposeAfter.Name)
	assert.Equal(t, []string{deployJob.Name}, (*dockerComposeAfter.Plan[0].Aggregate)[0].Passed)
}

func TestRendersCfDeploy_GetsArtifactWhenCfManifestFromArtifacts(t *testing.T) {
	taskDeploy := manifest.DeployCF{
		API:        "api",
		Space:      "space",
		Org:        "org",
		Username:   "user",
		Password:   "password",
		TestDomain: "test.domain",
		Manifest:   fmt.Sprintf("../%s/manifest.yml", artifactsInDir),
	}

	man := manifest.Manifest{
		Repo:  manifest.Repo{URI: "git@github.com:foo/reponame"},
		Tasks: []manifest.Task{taskDeploy},
	}

	plan := testPipeline().Render(man).Jobs[0].Plan

	getSteps := *plan[0].Aggregate
	assert.Equal(t, gitName, getSteps[0].Get)
	assert.Equal(t, restoreArtifactTask(man), plan[1])

	assert.Equal(t, path.Join(artifactsInDir, "manifest.yml"), plan[2].Params["manifestPath"])

}
