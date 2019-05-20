package pipeline

import (
	"testing"

	"github.com/concourse/concourse/atc"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestRenderConsumerIntegrationTestTaskInPrePromoteStage(t *testing.T) {
	p := testPipeline()

	dockerComposeService := "blah"
	envVars := manifest.Vars{"blah": "value", "blah1": "value1"}
	cit := manifest.ConsumerIntegrationTest{
		Name:                 "c-name",
		Consumer:             "c-consumer/c-path",
		ConsumerHost:         "c-host",
		Script:               "c-script",
		DockerComposeService: dockerComposeService,
		Vars:                 envVars,
	}
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
					cit,
				},
			},
		},
	}

	expectedVars := map[string]string{
		"CONSUMER_GIT_URI":       "git@github.com:springernature/c-consumer",
		"CONSUMER_PATH":          "c-path",
		"CONSUMER_SCRIPT":        "c-script",
		"CONSUMER_GIT_KEY":       "((halfpipe-github.private_key))",
		"CONSUMER_HOST":          "c-host",
		"PROVIDER_NAME":          "p-name",
		"PROVIDER_HOST_KEY":      "P_NAME_DEPLOYED_HOST",
		"PROVIDER_HOST":          buildTestRoute("test-name", "cf-space", ""),
		"DOCKER_COMPOSE_SERVICE": dockerComposeService,
		"GCR_PRIVATE_KEY":        "((halfpipe-gcr.private_key))",
		"GIT_CLONE_OPTIONS":      "",
		"blah":                   "value",
		"blah1":                  "value1",
	}

	expectedPlan := atc.PlanConfig{
		Attempts:   1,
		Task:       "c-name",
		Privileged: true,
		TaskConfig: &atc.TaskConfig{
			Platform:      "linux",
			Params:        expectedVars,
			ImageResource: &dockerComposeImageResource,
			Run: atc.TaskRunConfig{
				Path: "docker.sh",
				Dir:  gitDir + "/base.path",
				Args: runScriptArgs(consumerIntegrationTestToRunTask(cit, man), man, false),
			},
			Inputs: []atc.TaskInputConfig{
				{Name: gitDir},
			},
			Caches: config.CacheDirs,
		}}

	job := p.Render(man).Jobs[0]
	assert.Equal(t, expectedPlan, (*(*job.Plan[2].Aggregate)[0].Do)[0])
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
	assert.Equal(t, "p-host", (*(*job.Plan[2].Aggregate)[0].Do)[0].TaskConfig.Params["PROVIDER_HOST"])
}

func TestRenderConsumerIntegrationTestTaskOutsidePrePromote(t *testing.T) {
	p := testPipeline()

	cit := manifest.ConsumerIntegrationTest{
		Retries:      2,
		Name:         "c-name",
		Consumer:     "c-consumer/c-path",
		ConsumerHost: "c-host",
		ProviderHost: "p-host",
		Script:       "c-script",
	}
	man := manifest.Manifest{
		Pipeline: "p-name",
		Repo: manifest.Repo{
			URI:      "git@git:user/repo",
			BasePath: "base.path",
		},
		Tasks: []manifest.Task{
			cit,
		},
	}

	expectedVars := map[string]string{
		"CONSUMER_GIT_URI":       "git@github.com:springernature/c-consumer",
		"CONSUMER_PATH":          "c-path",
		"CONSUMER_SCRIPT":        "c-script",
		"CONSUMER_GIT_KEY":       "((halfpipe-github.private_key))",
		"CONSUMER_HOST":          "c-host",
		"PROVIDER_NAME":          "p-name",
		"PROVIDER_HOST_KEY":      "P_NAME_DEPLOYED_HOST",
		"PROVIDER_HOST":          "p-host",
		"DOCKER_COMPOSE_SERVICE": "",
		"GCR_PRIVATE_KEY":        "((halfpipe-gcr.private_key))",
		"GIT_CLONE_OPTIONS":      "",
	}

	expectedJob := atc.JobConfig{
		Name:   "c-name",
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{Aggregate: &atc.PlanSequence{
				atc.PlanConfig{
					Get:     gitName,
					Trigger: true,
				}}},
			atc.PlanConfig{
				Attempts:   3,
				Task:       "c-name",
				Privileged: true,
				TaskConfig: &atc.TaskConfig{
					Platform:      "linux",
					Params:        expectedVars,
					ImageResource: &dockerComposeImageResource,
					Run: atc.TaskRunConfig{
						Path: "docker.sh",
						Dir:  gitDir + "/base.path",
						Args: runScriptArgs(consumerIntegrationTestToRunTask(cit, man), man, false),
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

func TestBuildsTheRightDockerComposeRunScriptWithEnvVars(t *testing.T) {
	envVars := manifest.Vars{"blah": "value", "blah1": "value1"}

	actual := consumerIntegrationTestScript(envVars)

	expected := `docker-compose run --no-deps \
  --entrypoint "${CONSUMER_SCRIPT}" \
  -e DEPENDENCY_NAME=${PROVIDER_NAME} \
  -e ${PROVIDER_HOST_KEY}=${PROVIDER_HOST} -e blah -e blah1 \
  ${DOCKER_COMPOSE_SERVICE:-code}`

	assert.Contains(t, actual, expected)
}
