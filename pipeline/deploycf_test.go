package pipeline

import (
	"testing"

	"github.com/concourse/atc"
	"github.com/springernature/halfpipe/model"
	"github.com/stretchr/testify/assert"
)

func TestRendersCfDeployResources(t *testing.T) {
	manifest := model.Manifest{Repo: model.Repo{Uri: "git@github.com:foo/reponame"}}
	manifest.Tasks = []model.Task{
		model.DeployCF{
			Api:      "dev-api",
			Space:    "dev",
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
			ApiAlias: "live",
			Space:    "prod",
			Org:      "springer",
			Username: "rob",
			Password: "supersecret",
			Manifest: "manifest-live.yml",
		},
	}

	expectedDevResource := atc.ResourceConfig{
		Name: "CF springer-dev",
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
		Name: "CF live-springer-prod",
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

	config := testPipeline().Render(manifest)

	assert.Equal(t, expectedDevResource, config.Resources[1])
	assert.Equal(t, expectedDevJob, config.Jobs[0])

	assert.Equal(t, expectedLiveResource, config.Resources[2])
	assert.Equal(t, expectedLiveJob, config.Jobs[1])
}
