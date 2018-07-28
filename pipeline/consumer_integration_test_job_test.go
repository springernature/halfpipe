package pipeline

import (
	"testing"

	"github.com/concourse/atc"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestRenderConsumerIntegrationTestTaskInPrePromoteStage(t *testing.T) {
	p := testPipeline()

	dockerComposeService := "blah"
	man := manifest.Manifest{
		Pipeline: "p-name",
		Repo: manifest.Repo{
			URI:      "git@git:user/repo",
			BasePath: "base.path",
		},
		Tasks: []manifest.Task{
			manifest.DeployCF{
				Name:  "cf-deploy",
				API:   "cf-api",
				Space: "cf-space",
				Org:   "cf-org",
				PrePromote: []manifest.Task{
					manifest.ConsumerIntegrationTest{
						Name:                 "c-name",
						Consumer:             "c-consumer/c-path",
						ConsumerHost:         "c-host",
						Script:               "c-script",
						DockerComposeService: dockerComposeService,
					},
				},
			},
		},
	}

	expectedVars := map[string]string{
		"CONSUMER_GIT_URI":       "git@github.com:springernature/c-consumer",
		"CONSUMER_PATH":          "c-path",
		"CONSUMER_SCRIPT":        "c-script",
		"CONSUMER_GIT_KEY":       "((github.private_key))",
		"CONSUMER_HOST":          "c-host",
		"PROVIDER_NAME":          "p-name",
		"PROVIDER_HOST_KEY":      "P_NAME_DEPLOYED_HOST",
		"PROVIDER_HOST":          buildTestRoute("test-name", "cf-space", ""),
		"DOCKER_COMPOSE_SERVICE": dockerComposeService,
		"GCR_PRIVATE_KEY":        "((gcr.private_key))",
	}

	expectedPlan := atc.PlanConfig{
		Task:       "c-name",
		Privileged: true,
		TaskConfig: &atc.TaskConfig{
			Platform:      "linux",
			Params:        expectedVars,
			ImageResource: &dockerComposeImageResource,
			Run: atc.TaskRunConfig{
				Path: "docker.sh",
				Dir:  gitDir + "/base.path",
				Args: runScriptArgs(consumerIntegrationTestScript, false, "", false, nil, "../.git/ref"),
			},
			Inputs: []atc.TaskInputConfig{
				{Name: gitDir},
			},
			Caches: config.CacheDirs,
		}}

	job := p.Render(man).Jobs[0]
	assert.Equal(t, expectedPlan, job.Plan[2])
}

func TestRenderConsumerIntegrationTestTaskWithProviderHost(t *testing.T) {
	p := testPipeline()

	man := manifest.Manifest{
		Pipeline: "p-name",
		Repo: manifest.Repo{
			URI:      "git@git:user/repo",
			BasePath: "base.path",
		},
		Tasks: []manifest.Task{
			manifest.DeployCF{
				Name:  "cf-deploy",
				API:   "cf-api",
				Space: "cf-space",
				Org:   "cf-org",
				PrePromote: []manifest.Task{
					manifest.ConsumerIntegrationTest{
						Name:         "c-name",
						Consumer:     "c-consumer/c-path",
						ConsumerHost: "c-host",
						ProviderHost: "p-host",
						Script:       "c-script",
					},
				},
			},
		},
	}

	job := p.Render(man).Jobs[0]
	assert.Equal(t, "p-host", job.Plan[2].TaskConfig.Params["PROVIDER_HOST"])
}

func TestRenderConsumerIntegrationTestTaskOutsidePrePromote(t *testing.T) {
	p := testPipeline()

	man := manifest.Manifest{
		Pipeline: "p-name",
		Repo: manifest.Repo{
			URI:      "git@git:user/repo",
			BasePath: "base.path",
		},
		Tasks: []manifest.Task{
			manifest.ConsumerIntegrationTest{
				Name:         "c-name",
				Consumer:     "c-consumer/c-path",
				ConsumerHost: "c-host",
				ProviderHost: "p-host",
				Script:       "c-script",
			},
		},
	}

	expectedVars := map[string]string{
		"CONSUMER_GIT_URI":       "git@github.com:springernature/c-consumer",
		"CONSUMER_PATH":          "c-path",
		"CONSUMER_SCRIPT":        "c-script",
		"CONSUMER_GIT_KEY":       "((github.private_key))",
		"CONSUMER_HOST":          "c-host",
		"PROVIDER_NAME":          "p-name",
		"PROVIDER_HOST_KEY":      "P_NAME_DEPLOYED_HOST",
		"PROVIDER_HOST":          "p-host",
		"DOCKER_COMPOSE_SERVICE": "",
		"GCR_PRIVATE_KEY":        "((gcr.private_key))",
	}

	expectedJob := atc.JobConfig{
		Name:   "c-name",
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{Aggregate: &atc.PlanSequence{
				atc.PlanConfig{
					Get:     gitDir,
					Trigger: true,
				}}},
			atc.PlanConfig{
				Task:       "c-name",
				Privileged: true,
				TaskConfig: &atc.TaskConfig{
					Platform:      "linux",
					Params:        expectedVars,
					ImageResource: &dockerComposeImageResource,
					Run: atc.TaskRunConfig{
						Path: "docker.sh",
						Dir:  gitDir + "/base.path",
						Args: runScriptArgs(consumerIntegrationTestScript, false, "", false, nil, "../.git/ref"),
					},
					Inputs: []atc.TaskInputConfig{
						{Name: gitDir},
					},
					Caches: config.CacheDirs,
				}},
		},
	}

	jobs := p.Render(man).Jobs
	if assert.Len(t, jobs, 1) {
		assert.Equal(t, expectedJob, jobs[0])
	}
}
