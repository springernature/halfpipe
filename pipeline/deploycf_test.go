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
		},
		model.DeployCF{
			Api:      "live-api",
			Space:    "space-station",
			Org:      "springer",
			Username: "rob",
			Password: "supersecret",
		},
	}

	expectedDev := atc.ResourceConfig{
		Name: "resource-deploy-cf_Task0",
		Type: "cf",
		Source: atc.Source{
			"api":          "dev-api",
			"space":        "space-station",
			"organization": "springer",
			"password":     "supersecret",
			"username":     "rob",
		},
	}

	expectedLive := atc.ResourceConfig{
		Name: "resource-deploy-cf_Task1",
		Type: "cf",
		Source: atc.Source{
			"api":          "live-api",
			"space":        "space-station",
			"organization": "springer",
			"password":     "supersecret",
			"username":     "rob",
		},
	}
	assert.Equal(t, expectedDev, pipe.Render(manifest).Resources[1])
	assert.Equal(t, expectedLive, pipe.Render(manifest).Resources[2])
}
