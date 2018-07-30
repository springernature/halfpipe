package pipeline

import (
	"testing"

	"github.com/concourse/atc"
	"github.com/springernature/halfpipe/dockercompose"
	"github.com/stretchr/testify/assert"
)

func TestResourcesFromDockerCompose_Empty(t *testing.T) {
	assert.Empty(t, resourcesFromDockerCompose(dockercompose.DockerCompose{}))
}

func TestResourcesFromDockerCompose_MultipleServices(t *testing.T) {
	resources := resourcesFromDockerCompose(dockercompose.DockerCompose{
		Services: []dockercompose.Service{
			{Name: "app", Image: "eu.gcr.io/halfpipe-io/golang:latest"},
			{Name: "db", Image: "mydb"},
			{Name: "no-image", Image: ""},
		},
	})

	expected := []atc.ResourceConfig{
		{
			Name: "dockercompose-app",
			Type: "docker-image",
			Source: atc.Source{
				"repository": "eu.gcr.io/halfpipe-io/golang:latest",
				"username":   "_json_key",
				"password":   "((gcr.private_key))",
			},
		}, {
			Name: "dockercompose-db",
			Type: "docker-image",
			Source: atc.Source{
				"repository": "mydb",
			},
		},
	}

	assert.Equal(t, expected, resources)
}
