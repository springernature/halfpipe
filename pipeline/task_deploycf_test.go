package pipeline

import (
	"testing"

	"github.com/concourse/atc"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestRendersCfDeployResources(t *testing.T) {
	taskDeployDev := manifest.DeployCF{
		Api:      "dev-api",
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
		Api:      "live-api",
		Space:    "prod",
		Org:      "springer",
		Username: "rob",
		Password: "supersecret",
		Manifest: "manifest-live.yml",
	}

	man := manifest.Manifest{Repo: manifest.Repo{Uri: "git@github.com:foo/reponame"}}
	man.Tasks = []manifest.Task{taskDeployDev, taskDeployLive}

	expectedDevResource := atc.ResourceConfig{
		Name: deployCFResourceName(taskDeployDev),
		Type: "cf",
		Source: atc.Source{
			"api":          "dev-api",
			"space":        "dev",
			"organization": "springer",
			"password":     "supersecret",
			"username":     "rob",
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
					"manifest": "reponame/manifest-dev.yml",
					"environment_variables": map[string]interface{}{
						"VAR1": "value1",
						"VAR2": "value2",
					},
					"path": "reponame",
				},
			},
		},
	}

	expectedLiveResource := atc.ResourceConfig{
		Name: deployCFResourceName(taskDeployLive),
		Type: "cf",
		Source: atc.Source{
			"api":          "live-api",
			"space":        "prod",
			"organization": "springer",
			"password":     "supersecret",
			"username":     "rob",
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
					"manifest":              "reponame/manifest-live.yml",
					"environment_variables": map[string]interface{}{},
					"path":                  "reponame",
				},
			},
		},
	}

	config := testPipeline().Render(man)

	assert.Equal(t, expectedDevResource, config.Resources[1])
	assert.Equal(t, expectedDevJob, config.Jobs[0])

	assert.Equal(t, expectedLiveResource, config.Resources[2])
	assert.Equal(t, expectedLiveJob, config.Jobs[1])
}
