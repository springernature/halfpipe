package pipeline

import (
	"github.com/concourse/concourse/atc"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDoesntHaveResourceIfToggleIsDisabled(t *testing.T) {
	man := manifest.Manifest{}

	cfg := testPipeline().Render(man)

	_, found := cfg.Resources.Lookup(versionName)
	assert.False(t, found)
}

func TestHasCorrectResourceIfFeatureToggleIsEnabled(t *testing.T) {
	team := "team"
	pipeline := "pipeline"
	man := manifest.Manifest{
		Team:     team,
		Pipeline: pipeline,
		FeatureToggles: manifest.FeatureToggles{
			manifest.FeatureUpdatePipeline,
		},
	}

	cfg := testPipeline().Render(man)

	resource, found := cfg.Resources.Lookup(versionName)
	assert.True(t, found)
	assert.Equal(t, "semver", resource.Type)

	source := resource.Source
	assert.Equal(t, "gcs", source["driver"])
	assert.Equal(t, config.VersionBucket, source["bucket"])
	assert.Equal(t, config.VersionJSONKey, source["json_key"])
	assert.Equal(t, team+"-"+pipeline, source["key"])

	branch := "branch"
	manWithBranch := manifest.Manifest{
		Team:     team,
		Pipeline: pipeline,
		Triggers: manifest.TriggerList{
			manifest.GitTrigger{
				Branch: branch,
			},
		},
		FeatureToggles: manifest.FeatureToggles{
			manifest.FeatureUpdatePipeline,
		},
	}

	cfgWithBranch := testPipeline().Render(manWithBranch)
	resourceWithBranch, found := cfgWithBranch.Resources.Lookup(versionName)
	assert.True(t, found)
	assert.Equal(t, "semver", resource.Type)

	sourceWithBranch := resourceWithBranch.Source
	assert.Equal(t, "gcs", sourceWithBranch["driver"])
	assert.Equal(t, config.VersionBucket, sourceWithBranch["bucket"])
	assert.Equal(t, config.VersionJSONKey, sourceWithBranch["json_key"])
	assert.Equal(t, team+"-"+pipeline+"-"+branch, sourceWithBranch["key"])
}

func TestShouldNotAddAVersionJobAIfFeatureToggleIsNotEnabled(t *testing.T) {
	man := manifest.Manifest{}
	cfg := testPipeline().Render(man)

	_, found := cfg.Jobs.Lookup(versionName)
	assert.False(t, found)
}

func TestShouldAddAVersionJobAsFirstJobIfFeatureToggleIsEnabled(t *testing.T) {
	man := manifest.Manifest{
		FeatureToggles: manifest.FeatureToggles{
			manifest.FeatureUpdatePipeline,
		},
		Tasks: manifest.TaskList{
			manifest.Update{},
		},
	}

	cfg := testPipeline().Render(man)

	_, found := cfg.Jobs.Lookup(updateJobName)
	assert.True(t, found)
	assert.Equal(t, updateJobName, cfg.Jobs[0].Name)
}

func TestGetShouldNotContainGetOnVersionIfFeatureToggleIsNotEnabled(t *testing.T) {
	jobName := "run"
	man := manifest.Manifest{
		Tasks: manifest.TaskList{
			manifest.Run{
				Name: jobName,
			},
		},
	}

	cfg := testPipeline().Render(man)
	config, found := cfg.Jobs.Lookup(jobName)
	assert.True(t, found)

	inParallel := config.Plan[0].InParallel
	for _, get := range inParallel.Steps {
		assert.NotEqual(t, versionName, get.Get)
	}
}

func TestGetShouldContainGetOnVersionIfFeatureToggleIsEnabled(t *testing.T) {
	jobName := "run"
	man := manifest.Manifest{
		FeatureToggles: manifest.FeatureToggles{
			manifest.FeatureUpdatePipeline,
		},

		Tasks: manifest.TaskList{
			manifest.Run{
				Name: jobName,
			},
		},
	}

	cfg := testPipeline().Render(man)
	config, found := cfg.Jobs.Lookup(jobName)
	assert.True(t, found)

	var versionPlan atc.PlanConfig
	for _, get := range config.Plan[0].InParallel.Steps {
		if get.Get == versionName {
			versionPlan = get
		}
	}

	assert.Equal(t, versionName, versionPlan.Get)
}

func TestVersionGetShouldBeTheOnlyOneWithTriggerTrue(t *testing.T) {
	firstJob := "run"
	secondJob := "run2"
	man := manifest.Manifest{
		FeatureToggles: manifest.FeatureToggles{
			manifest.FeatureUpdatePipeline,
		},

		Tasks: manifest.TaskList{
			manifest.Update{},
			manifest.Run{
				Name: firstJob,
			},
			manifest.Run{
				Name: secondJob,
			},
		},
	}

	cfg := testPipeline().Render(man)

	updateConfig, found := cfg.Jobs.Lookup(updateJobName)
	assert.True(t, found)
	for _, get := range updateConfig.Plan[0].InParallel.Steps {
		assert.True(t, get.Trigger)
	}

	firstTask, found := cfg.Jobs.Lookup(firstJob)
	assert.True(t, found)
	for _, get := range firstTask.Plan[0].InParallel.Steps {
		if get.Get == versionName {
			assert.True(t, get.Trigger)
		} else {
			assert.False(t, get.Trigger)
		}
		assert.Equal(t, []string{updateJobName}, get.Passed)
	}

	secondTask, found := cfg.Jobs.Lookup(secondJob)
	assert.True(t, found)
	for _, get := range secondTask.Plan[0].InParallel.Steps {
		if get.Get == versionName {
			assert.True(t, get.Trigger)
		} else {
			assert.False(t, get.Trigger)
		}
		assert.Equal(t, []string{firstJob}, get.Passed)
	}
}

func TestUpdateVersionShouldBeTheOnlyJobThatHasTheCronTrigger(t *testing.T) {
	firstJob := "run"
	secondJob := "run2"
	man := manifest.Manifest{
		FeatureToggles: manifest.FeatureToggles{
			manifest.FeatureUpdatePipeline,
		},
		Triggers: manifest.TriggerList{
			manifest.GitTrigger{},
			manifest.TimerTrigger{Cron: "* * * * *"},
		},
		Tasks: manifest.TaskList{
			manifest.Update{},
			manifest.Run{
				Name: firstJob,
			},
			manifest.Run{
				Name: secondJob,
			},
		},
	}

	cfg := testPipeline().Render(man)

	var cronFound bool
	var versionFound bool
	updateVersionConfig, found := cfg.Jobs.Lookup(updateJobName)
	assert.True(t, found)
	for _, get := range updateVersionConfig.Plan[0].InParallel.Steps {
		if get.Get == cronName {
			cronFound = true
		}
		if get.Get == versionName {
			versionFound = true
		}
		assert.True(t, get.Trigger)
	}

	assert.True(t, cronFound)
	assert.False(t, versionFound)

	firstTask, found := cfg.Jobs.Lookup(firstJob)
	assert.True(t, found)
	for _, get := range firstTask.Plan[0].InParallel.Steps {
		if get.Get == versionName {
			assert.True(t, get.Trigger)
		} else {
			assert.False(t, get.Trigger)
		}
		assert.NotContains(t, cronName, get.Get)
		assert.Equal(t, []string{updateJobName}, get.Passed)
	}

	secondTask, found := cfg.Jobs.Lookup(secondJob)
	assert.True(t, found)
	for _, get := range secondTask.Plan[0].InParallel.Steps {
		if get.Get == versionName {
			assert.True(t, get.Trigger)
		} else {
			assert.False(t, get.Trigger)
		}
		assert.NotContains(t, cronName, get.Get)
		assert.Equal(t, []string{firstJob}, get.Passed)
	}
}

func TestUpdateVersionShouldAddTheVersionAsAInputToAllJobsAndEnvVar(t *testing.T) {
	// We don't need to care about docker-push and deploy-cf.
	// As they are inputs in the inParallel the put containers will have them mapped..
	man := manifest.Manifest{
		FeatureToggles: manifest.FeatureToggles{
			manifest.FeatureUpdatePipeline,
		},

		Tasks: manifest.TaskList{
			manifest.Run{Name: "run"},
			manifest.DockerCompose{Name: "dockerCompose"},
			manifest.ConsumerIntegrationTest{Name: "cIT"},
			manifest.DeployMLZip{Name: "deployMLZip"},
			manifest.DeployMLModules{Name: "deployMLModules"},
			manifest.DeployCF{
				Name: "deploy",
				PrePromote: manifest.TaskList{
					manifest.Run{Name: "deployRun"},
					manifest.DockerCompose{Name: "deployDockerCompose"},
					manifest.ConsumerIntegrationTest{Name: "deployCIT"},
				},
			},
		},
	}

	expectedInput := atc.TaskInputConfig{
		Name: versionName,
	}

	config := testPipeline().Render(man)

	run, _ := config.Jobs.Lookup("run")
	assert.Contains(t, run.Plan[1].TaskConfig.Inputs, expectedInput)
	assert.Contains(t, run.Plan[1].TaskConfig.Run.Args[1], "export BUILD_VERSION=`cat ../version/version`")

	dockerCompose, _ := config.Jobs.Lookup("dockerCompose")
	assert.Contains(t, dockerCompose.Plan[1].TaskConfig.Inputs, expectedInput)
	assert.Contains(t, dockerCompose.Plan[1].TaskConfig.Run.Args[1], "export BUILD_VERSION=`cat ../version/version`")
	assert.Contains(t, dockerCompose.Plan[1].TaskConfig.Run.Args[1], "-e BUILD_VERSION")

	cIT, _ := config.Jobs.Lookup("cIT")
	assert.Contains(t, cIT.Plan[1].TaskConfig.Inputs, expectedInput)
	assert.Contains(t, cIT.Plan[1].TaskConfig.Run.Args[1], "export BUILD_VERSION=`cat ../version/version`")

	deployMLZip, _ := config.Jobs.Lookup("deployMLZip")
	assert.Contains(t, deployMLZip.Plan[2].TaskConfig.Inputs, expectedInput)
	assert.Contains(t, deployMLZip.Plan[2].TaskConfig.Run.Args[1], "export BUILD_VERSION=`cat ../version/version`")

	deployMLModules, _ := config.Jobs.Lookup("deployMLModules")
	assert.Contains(t, deployMLModules.Plan[1].TaskConfig.Inputs, expectedInput)
	assert.Contains(t, deployMLModules.Plan[1].TaskConfig.Run.Args[1], "export BUILD_VERSION=`cat ../version/version`")

	var foundPrePromoteTasks int
	deploy, _ := config.Jobs.Lookup("deploy")
	for _, plan := range deploy.Plan {
		if plan.InParallel != nil {
			inParallel := *plan.InParallel
			for _, a := range inParallel.Steps {
				if a.Do != nil {
					for _, prePromoteTask := range *a.Do {
						foundPrePromoteTasks++
						assert.Contains(t, prePromoteTask.TaskConfig.Inputs, expectedInput)
						assert.Contains(t, prePromoteTask.TaskConfig.Run.Args[1], "export BUILD_VERSION=`cat ../version/version`")

					}
				}
			}
		}
	}

	assert.Equal(t, 3, foundPrePromoteTasks)
}

func TestUpdateVersionShouldAddTheVersionAsAInputToAllJobsAndEnvVarWhenInMonoRepo(t *testing.T) {
	// We don't need to care about docker-push and deploy-cf.
	// As they are inputs in the inParallel the put containers will have them mapped..

	basePath := "apps/app1"
	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.GitTrigger{
				BasePath: basePath,
			},
		},

		FeatureToggles: manifest.FeatureToggles{
			manifest.FeatureUpdatePipeline,
		},

		Tasks: manifest.TaskList{
			manifest.Run{Name: "run"},
			manifest.DockerCompose{Name: "dockerCompose"},
			manifest.ConsumerIntegrationTest{Name: "cIT"},
			manifest.DeployMLZip{Name: "deployMLZip"},
			manifest.DeployMLModules{Name: "deployMLModules"},
			manifest.DeployCF{
				Name: "deploy",
				PrePromote: manifest.TaskList{
					manifest.Run{Name: "deployRun"},
					manifest.DockerCompose{Name: "deployDockerCompose"},
					manifest.ConsumerIntegrationTest{Name: "deployCIT"},
				},
			},
		},
	}

	expectedInput := atc.TaskInputConfig{
		Name: versionName,
	}

	config := testPipeline().Render(man)

	run, _ := config.Jobs.Lookup("run")
	assert.Contains(t, run.Plan[1].TaskConfig.Inputs, expectedInput)
	assert.Contains(t, run.Plan[1].TaskConfig.Run.Args[1], "export BUILD_VERSION=`cat ../../../version/version`")

	dockerCompose, _ := config.Jobs.Lookup("dockerCompose")
	assert.Contains(t, dockerCompose.Plan[1].TaskConfig.Inputs, expectedInput)
	assert.Contains(t, dockerCompose.Plan[1].TaskConfig.Run.Args[1], "export BUILD_VERSION=`cat ../../../version/version`")
	assert.Contains(t, dockerCompose.Plan[1].TaskConfig.Run.Args[1], "-e BUILD_VERSION")

	cIT, _ := config.Jobs.Lookup("cIT")
	assert.Contains(t, cIT.Plan[1].TaskConfig.Inputs, expectedInput)
	assert.Contains(t, cIT.Plan[1].TaskConfig.Run.Args[1], "export BUILD_VERSION=`cat ../../../version/version`")

	deployMLZip, _ := config.Jobs.Lookup("deployMLZip")
	assert.Contains(t, deployMLZip.Plan[2].TaskConfig.Inputs, expectedInput)
	assert.Contains(t, deployMLZip.Plan[2].TaskConfig.Run.Args[1], "export BUILD_VERSION=`cat ../../../version/version`")

	deployMLModules, _ := config.Jobs.Lookup("deployMLModules")
	assert.Contains(t, deployMLModules.Plan[1].TaskConfig.Inputs, expectedInput)
	assert.Contains(t, deployMLModules.Plan[1].TaskConfig.Run.Args[1], "export BUILD_VERSION=`cat ../../../version/version`")

	var foundPrePromoteTasks int
	deploy, _ := config.Jobs.Lookup("deploy")
	for _, plan := range deploy.Plan {
		if plan.InParallel != nil {
			for _, a := range plan.InParallel.Steps {
				if a.Do != nil {
					for _, prePromoteTask := range *a.Do {
						foundPrePromoteTasks++
						assert.Contains(t, prePromoteTask.TaskConfig.Inputs, expectedInput)
						assert.Contains(t, prePromoteTask.TaskConfig.Run.Args[1], "export BUILD_VERSION=`cat ../../../version/version`")
					}
				}
			}
		}
	}

	assert.Equal(t, 3, foundPrePromoteTasks)
}
