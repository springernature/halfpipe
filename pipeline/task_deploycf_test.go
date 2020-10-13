package pipeline

import (
	"github.com/springernature/halfpipe/config"
	"testing"

	"path"

	"fmt"

	cfManifest "code.cloudfoundry.org/cli/util/manifest"
	"github.com/cloudfoundry/bosh-cli/director/template"
	"github.com/concourse/concourse/atc"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestRendersCfDeploy(t *testing.T) {
	liveAPI := "api.live.com"
	liveTestDomain := "test.live.com"

	devAPI := "api.dev.com"
	devTestDomain := "test.dev.com"

	taskDeployDev := manifest.DeployCF{
		Name:       "Deploy to dev",
		API:        devAPI,
		Space:      "dev",
		Org:        "springer",
		Username:   "rob",
		Password:   "supersecret",
		PreStart:   []string{"cf events my-app", "cf blah"},
		TestDomain: devTestDomain,
		Manifest:   "manifest-dev.yml",
		Vars: manifest.Vars{
			"VAR1": "value1",
			"VAR2": "value2",
		},
		CliVersion: "cf6",
	}

	taskDeployQa := manifest.DeployCF{
		Name:       "Deploy to qa",
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
		CliVersion: "cf6",
	}

	taskDeployStaging := manifest.DeployCF{
		Name:       "Deploy to staging",
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
		CliVersion: "cf6",
	}

	timeout := "5m"
	taskDeployLive := manifest.DeployCF{
		Name:       "Deploy to live",
		API:        liveAPI,
		Space:      "prod",
		Org:        "springer",
		TestDomain: liveTestDomain,
		Username:   "rob",
		Password:   "supersecret",
		Manifest:   "manifest-live.yml",
		Timeout:    timeout,
		CliVersion: "cf6",
	}

	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.GitTrigger{
				URI: "git@github.com:foo/reponame",
			},
		},
	}
	man.Tasks = []manifest.Task{
		taskDeployDev,
		manifest.Parallel{
			Tasks: manifest.TaskList{
				manifest.Sequence{
					Tasks: manifest.TaskList{
						taskDeployQa,
					},
				},
				manifest.Sequence{
					Tasks: manifest.TaskList{
						manifest.Run{},
						taskDeployStaging,
					},
				},
			},
		},
		taskDeployLive}

	expectedResourceType := atc.ResourceType{
		Name: "cf-resource",
		Type: "registry-image",
		Source: atc.Source{
			"repository": config.DockerRegistry + "cf-resource-v2",
			"tag":        "stable",
			"password":   "((halfpipe-gcr.private_key))",
			"username":   "_json_key",
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
		CheckEvery: longResourceCheckInterval,
	}

	expectedQAResource := atc.ResourceConfig{
		Name: deployCFResourceName(taskDeployQa),
		Type: "cf-resource",
		Source: atc.Source{
			"api":      devAPI,
			"space":    "dev",
			"org":      "springer",
			"password": "supersecret",
			"username": "rob",
		},
		CheckEvery: longResourceCheckInterval,
	}

	expectedStagingResource := atc.ResourceConfig{
		Name: deployCFResourceName(taskDeployStaging),
		Type: "cf-resource",
		Source: atc.Source{
			"api":      devAPI,
			"space":    "dev",
			"org":      "springer",
			"password": "supersecret",
			"username": "rob",
		},
		CheckEvery: longResourceCheckInterval,
	}

	manifestPath := path.Join(gitDir, "manifest-dev.yml")

	envVars := map[string]interface{}{
		"VAR1": "value1",
		"VAR2": "value2",
	}
	expectedDevJob := atc.JobConfig{
		Name: taskDeployDev.Name,
		BuildLogRetention: &(atc.BuildLogRetention{
			Builds:                 0,
			MinimumSucceededBuilds: 1,
		}),
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{InParallel: &atc.InParallelConfig{FailFast: true, Steps: atc.PlanSequence{atc.PlanConfig{Get: gitDir, Trigger: true, Attempts: gitGetAttempts}}}},
			atc.PlanConfig{
				Put:      "halfpipe-push",
				Attempts: 2,
				Resource: expectedDevResource.Name,
				Params: atc.Params{
					"command":         "halfpipe-push",
					"testDomain":      devTestDomain,
					"manifestPath":    manifestPath,
					"vars":            envVars,
					"appPath":         gitDir,
					"gitRefPath":      path.Join(gitDir, ".git", "ref"),
					"preStartCommand": "cf events my-app; cf blah",
					"cliVersion":      "cf6",
				},
			},
			atc.PlanConfig{
				Put:      "halfpipe-check",
				Attempts: 2,
				Resource: expectedDevResource.Name,
				Params: atc.Params{
					"command":      "halfpipe-check",
					"manifestPath": manifestPath,
					"cliVersion":   "cf6",
				},
			},
			atc.PlanConfig{
				Put:      "halfpipe-promote",
				Attempts: 2,
				Resource: expectedDevResource.Name,
				Params: atc.Params{
					"command":      "halfpipe-promote",
					"testDomain":   devTestDomain,
					"manifestPath": manifestPath,
					"cliVersion":   "cf6",
				},
			},
		},
		Ensure: &atc.PlanConfig{
			Put:      "halfpipe-cleanup",
			Attempts: 2,
			Resource: expectedDevResource.Name,
			Params: atc.Params{
				"command":      "halfpipe-cleanup",
				"manifestPath": manifestPath,
				"cliVersion":   "cf6",
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
		CheckEvery: longResourceCheckInterval,
	}

	liveManifest := path.Join(gitDir, "manifest-live.yml")
	expectedLiveJob := atc.JobConfig{
		Name:   taskDeployLive.Name,
		Serial: true,
		BuildLogRetention: &(atc.BuildLogRetention{
			Builds:                 0,
			MinimumSucceededBuilds: 1,
		}),
		Plan: atc.PlanSequence{
			atc.PlanConfig{
				InParallel: &atc.InParallelConfig{
					FailFast: true,
					Steps:    atc.PlanSequence{atc.PlanConfig{Get: gitDir, Trigger: true, Attempts: gitGetAttempts, Passed: []string{taskDeployQa.Name, taskDeployStaging.Name}}},
				},
				Timeout: timeout,
			},
			atc.PlanConfig{
				Put:      "halfpipe-push",
				Attempts: 2,
				Resource: expectedLiveResource.Name,
				Params: atc.Params{
					"command":      "halfpipe-push",
					"testDomain":   liveTestDomain,
					"manifestPath": liveManifest,
					"appPath":      gitDir,
					"gitRefPath":   path.Join(gitDir, ".git", "ref"),
					"timeout":      "5m",
					"cliVersion":   "cf6",
				},
				Timeout: timeout,
			},
			atc.PlanConfig{
				Put:      "halfpipe-check",
				Attempts: 2,
				Resource: expectedLiveResource.Name,
				Params: atc.Params{
					"command":      "halfpipe-check",
					"manifestPath": liveManifest,
					"timeout":      "5m",
					"cliVersion":   "cf6",
				},
				Timeout: timeout,
			},
			atc.PlanConfig{
				Put:      "halfpipe-promote",
				Attempts: 2,
				Resource: expectedLiveResource.Name,
				Params: atc.Params{
					"command":      "halfpipe-promote",
					"testDomain":   liveTestDomain,
					"manifestPath": liveManifest,
					"timeout":      "5m",
					"cliVersion":   "cf6",
				},
				Timeout: timeout,
			},
		},
		Ensure: &atc.PlanConfig{
			Put:      "halfpipe-cleanup",
			Attempts: 2,
			Resource: expectedLiveResource.Name,
			Params: atc.Params{
				"command":      "halfpipe-cleanup",
				"manifestPath": liveManifest,
				"timeout":      "5m",
				"cliVersion":   "cf6",
			},
			Timeout: timeout,
		},
	}

	cfg := testPipeline().Render(man)

	assert.Len(t, cfg.ResourceTypes, 1)
	assert.Equal(t, expectedResourceType, cfg.ResourceTypes[0])

	// dev
	foundDevResource, found := cfg.Resources.Lookup(expectedDevResource.Name)
	assert.True(t, found)
	assert.Equal(t, expectedDevResource, foundDevResource)

	foundDevJob, found := cfg.Jobs.Lookup(expectedDevJob.Name)
	assert.True(t, found)
	assert.Equal(t, expectedDevJob, foundDevJob)

	// qa
	foundQaResource, found := cfg.Resources.Lookup(expectedQAResource.Name)
	assert.True(t, found)
	assert.Equal(t, expectedQAResource, foundQaResource)

	// staging
	foundStagingResource, found := cfg.Resources.Lookup(expectedStagingResource.Name)
	assert.True(t, found)
	assert.Equal(t, expectedStagingResource, foundStagingResource)

	// live
	foundLiveResource, found := cfg.Resources.Lookup(expectedLiveResource.Name)
	assert.True(t, found)
	assert.Equal(t, expectedLiveResource, foundLiveResource)

	foundLiveJob, found := cfg.Jobs.Lookup(expectedLiveJob.Name)
	assert.True(t, found)
	assert.Equal(t, expectedLiveJob, foundLiveJob)
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

	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.GitTrigger{
				URI: "git@github:org/repo-name",
			},
		},
	}
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
	assert.Equal(t, manifest.GitTrigger{}.GetTriggerName(), (plan[0].InParallel.Steps)[0].Get)
	assert.Equal(t, cfg.Jobs[0].Name, (plan[0].InParallel.Steps)[0].Passed[0])
	assert.Equal(t, restoreArtifactTask(man), plan[1])
	assert.Equal(t, "halfpipe-push", plan[2].Params["command"])

	//halfpipe-check
	assert.Equal(t, "halfpipe-check", plan[3].Params["command"])

	//pre promote 1
	pp1 := (*(plan[4].InParallel.Steps)[0].Do)[0]
	assert.Equal(t, "pp1", pp1.Task)
	assert.Equal(t, "app-cf-space-CANDIDATE.test.domain.com", pp1.TaskConfig.Params["TEST_ROUTE"])
	assert.Equal(t, "pp1", pp1.TaskConfig.Params["PP1"])

	//pre promote 2
	pp2 := (*(plan[4].InParallel.Steps)[1].Do)[0]
	assert.Equal(t, "pp2", pp2.Task)
	assert.Equal(t, "app-cf-space-CANDIDATE.test.domain.com", pp2.TaskConfig.Params["TEST_ROUTE"])

	//halfpipe-promote
	assert.Equal(t, "halfpipe-promote", plan[5].Params["command"])

	//halfpipe-cleanup (ensure)
	assert.Equal(t, "halfpipe-cleanup", deployJob.Ensure.Params["command"])

	//docker-compose after
	dockerComposeAfter := cfg.Jobs[2]
	assert.Equal(t, dockerComposeTaskAfter.Name, dockerComposeAfter.Name)
	assert.Equal(t, []string{deployJob.Name}, (dockerComposeAfter.Plan[0].InParallel.Steps)[0].Passed)
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

	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.GitTrigger{},
		},
	}
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
	assert.Equal(t, manifest.GitTrigger{}.GetTriggerName(), (plan[0].InParallel.Steps)[0].Get)
	assert.Equal(t, cfg.Jobs[0].Name, (plan[0].InParallel.Steps)[0].Passed[0])
	assert.Equal(t, restoreArtifactTask(man), plan[1])
	assert.Equal(t, "halfpipe-push", plan[2].Params["command"])

	//halfpipe-check
	assert.Equal(t, "halfpipe-check", plan[3].Params["command"])

	//pre promote 1
	pp1 := (*plan[4].Do)[0]
	assert.Equal(t, "pp1", pp1.Task)
	assert.Equal(t, "app-cf-space-CANDIDATE.test.domain.com", pp1.TaskConfig.Params["TEST_ROUTE"])
	assert.Equal(t, "pp1", pp1.TaskConfig.Params["PP1"])

	//pre promote 2
	pp2 := (*plan[5].Do)[0]
	assert.Equal(t, "pp2", pp2.Task)
	assert.Equal(t, "app-cf-space-CANDIDATE.test.domain.com", pp2.TaskConfig.Params["TEST_ROUTE"])
	assert.Equal(t, "pp2", pp2.TaskConfig.Params["PP2"])

	//halfpipe-promote
	assert.Equal(t, "halfpipe-promote", plan[6].Params["command"])

	//halfpipe-cleanup (ensure)
	assert.Equal(t, "halfpipe-cleanup", deployJob.Ensure.Params["command"])

	//docker-compose after
	dockerComposeAfter := cfg.Jobs[2]
	assert.Equal(t, dockerComposeTaskAfter.Name, dockerComposeAfter.Name)
	assert.Equal(t, []string{deployJob.Name}, (dockerComposeAfter.Plan[0].InParallel.Steps)[0].Passed)
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
		Triggers: manifest.TriggerList{
			manifest.GitTrigger{},
		},
		Tasks: []manifest.Task{taskDeploy},
	}

	plan := testPipeline().Render(man).Jobs[0].Plan

	getSteps := plan[0].InParallel.Steps
	assert.Equal(t, manifest.GitTrigger{}.GetTriggerName(), getSteps[0].Get)
	assert.Equal(t, restoreArtifactTask(man), plan[1])

	assert.Equal(t, path.Join(artifactsInDir, "manifest.yml"), plan[2].Params["manifestPath"])

}

func TestSetsBuildVersionPathParamForVersionedPipelines(t *testing.T) {
	task := manifest.DeployCF{
		API:        "api",
		Space:      "space",
		Org:        "org",
		Username:   "user",
		Password:   "password",
		TestDomain: "test.domain",
		Manifest:   fmt.Sprintf("../%s/manifest.yml", artifactsInDir),
	}

	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.Update{},
			task,
		},
		FeatureToggles: manifest.FeatureToggles{manifest.FeatureUpdatePipeline},
	}

	//unversioned
	rendered := testPipeline().Render(man)
	man.FeatureToggles = append(man.FeatureToggles, "update-pipeline")
	buildVersionPath := rendered.Jobs[1].Plan[2].Params["buildVersionPath"]
	assert.Equal(t, path.Join("version", "version"), buildVersionPath)

}

func TestIncludesResourcesForDeployCF(t *testing.T) {
	deployTask := manifest.DeployCF{
		API:        "api",
		Space:      "space",
		Org:        "org",
		Username:   "user",
		Password:   "password",
		TestDomain: "test.domain",
		Manifest:   fmt.Sprintf("../%s/manifest.yml", artifactsInDir),
	}

	rollingDeployTask := manifest.DeployCF{
		API:        "api",
		Space:      "space",
		Org:        "org",
		Username:   "user",
		Password:   "password",
		TestDomain: "test.domain",
		Manifest:   fmt.Sprintf("../%s/manifest.yml", artifactsInDir),
		Rolling:    true,
	}

	man := manifest.Manifest{
		Tasks: []manifest.Task{
			deployTask,
			rollingDeployTask,
		},
	}

	pipeline := testPipeline().Render(man)

	r, found := pipeline.ResourceTypes.Lookup(deployCfResourceTypeName)
	assert.True(t, found)

	assert.Contains(t, r.Source["repository"].(string), "cf-resource-v2")
}

func TestTestDomain(t *testing.T) {
	t.Run("Without underscore", func(t *testing.T) {
		assert.Equal(t, "appName-spaceName-CANDIDATE.domain.com", buildTestRoute("appName", "spaceName", "domain.com"))
	})

	t.Run("With underscore", func(t *testing.T) {
		assert.Equal(t, "app-Name-space-Name-CANDIDATE.domain.com", buildTestRoute("app_Name", "space_Name", "domain.com"))
	})
}
