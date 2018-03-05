package pipeline

import (
	"testing"

	"github.com/concourse/atc"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestRenderDockerPushTask(t *testing.T) {
	man := manifest.Manifest{}
	man.Repo.Uri = "git@github.com:/springernature/foo.git"

	username := "halfpipe"
	password := "secret"
	repo := "halfpipe/halfpipe-cli"
	man.Tasks = []manifest.Task{
		manifest.DockerPush{
			Username: username,
			Password: password,
			Image:    repo,
		},
	}

	expectedResource := atc.ResourceConfig{
		Name: "Docker Registry",
		Type: "docker-image",
		Source: atc.Source{
			"username":   username,
			"password":   password,
			"repository": repo,
		},
	}

	expectedJobConfig := atc.JobConfig{
		Name:   "docker-push",
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{Get: man.Repo.GetName(), Trigger: true},
			atc.PlanConfig{Put: "Docker Registry", Params: atc.Params{"build": man.Repo.GetName()}},
		},
	}

	// First resource will always be the git resource.
	assert.Equal(t, expectedResource, testPipeline().Render(man).Resources[1])
	assert.Equal(t, expectedJobConfig, testPipeline().Render(man).Jobs[0])
}

func TestRenderDockerPushTaskNotInRoot(t *testing.T) {
	man := manifest.Manifest{}
	man.Repo.Uri = "git@github.com:/springernature/foo.git"
	basePath := "subapp/sub2"
	man.Repo.BasePath = basePath

	username := "halfpipe"
	password := "secret"
	repo := "halfpipe/halfpipe-cli"
	man.Tasks = []manifest.Task{
		manifest.DockerPush{
			Username: username,
			Password: password,
			Image:    repo,
		},
	}

	expectedResource := atc.ResourceConfig{
		Name: "Docker Registry",
		Type: "docker-image",
		Source: atc.Source{
			"username":   username,
			"password":   password,
			"repository": repo,
		},
	}

	expectedJobConfig := atc.JobConfig{
		Name:   "docker-push",
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{Get: man.Repo.GetName(), Trigger: true},
			atc.PlanConfig{Put: "Docker Registry", Params: atc.Params{"build": man.Repo.GetName() + "/" + basePath}},
		},
	}

	// First resource will always be the git resource.
	assert.Equal(t, expectedResource, testPipeline().Render(man).Resources[1])
	assert.Equal(t, expectedJobConfig, testPipeline().Render(man).Jobs[0])
}
