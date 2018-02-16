package pipeline

import (
	"testing"

	"github.com/concourse/atc"
	"github.com/springernature/halfpipe/model"
	"github.com/stretchr/testify/assert"
)

func TestRendersCfDeployResources(t *testing.T) {
	manifest := model.Manifest{}
	manifest.Tasks = []model.Task{
		model.DeployCF{
			Api:      "dev-api",
			Space:    "space-station",
			Org:      "springer",
			Username: "rob",
			Password: "supersecret",
			Manifest: "manifest-dev.yml",
			Vars: model.Vars{
				"VAR1": "value1",
				"VAR2": "value2",
			},
		},
		model.DeployCF{
			Api:      "live-api",
			Space:    "space-station",
			Org:      "springer",
			Username: "rob",
			Password: "supersecret",
			Manifest: "manifest-live.yml",
		},
	}

	expectedDevResource := atc.ResourceConfig{
		Name: "1. Cloud Foundry",
		Type: "cf",
		Source: atc.Source{
			"api":          "dev-api",
			"space":        "space-station",
			"organization": "springer",
			"password":     "supersecret",
			"username":     "rob",
		},
	}

	expectedLiveResource := atc.ResourceConfig{
		Name: "2. Cloud Foundry",
		Type: "cf",
		Source: atc.Source{
			"api":          "live-api",
			"space":        "space-station",
			"organization": "springer",
			"password":     "supersecret",
			"username":     "rob",
		},
	}

	expectedDevJob := atc.JobConfig{
		Name:   "1. deploy-cf",
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{Get: manifest.Repo.GetName(), Trigger: true},
			atc.PlanConfig{
				Put: "1. deploy-cf",
				Params: atc.Params{
					"manifest": "manifest-dev.yml",
					"environment_variables": map[string]interface{}{
						"VAR1": "value1",
						"VAR2": "value2",
					},
				},
			},
		},
	}

	config := pipe.Render(manifest)

	assert.Equal(t, expectedDevResource, config.Resources[1])
	assert.Equal(t, expectedLiveResource, config.Resources[2])

	assert.Equal(t, expectedDevJob, config.Jobs[0])
	assert.Equal(t, expectedDevJob, config.Jobs[0])
}
