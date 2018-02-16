package pipeline

import (
	"fmt"
	"testing"

	"github.com/concourse/atc"
	"github.com/springernature/halfpipe/model"
	"github.com/stretchr/testify/assert"
)

var pipe = Pipeline{}

func TestRendersHttpGitResource(t *testing.T) {
	name := "yolo"
	gitUri := fmt.Sprintf("git@github.com:springernature/%s.git", name)

	manifest := model.Manifest{}
	manifest.Repo.Uri = gitUri

	expected := atc.Config{
		Resources: atc.ResourceConfigs{
			atc.ResourceConfig{
				Name: name,
				Type: "git",
				Source: atc.Source{
					"uri": gitUri,
				},
			},
		},
	}
	assert.Equal(t, expected, pipe.Render(manifest))
}

func TestRendersSshGitResource(t *testing.T) {
	name := "asdf"
	gitUri := fmt.Sprintf("git@github.com:springernature/%s.git/", name)
	privateKey := "blurgh"

	manifest := model.Manifest{}
	manifest.Repo.Uri = gitUri
	manifest.Repo.PrivateKey = privateKey

	expected := atc.Config{
		Resources: atc.ResourceConfigs{
			atc.ResourceConfig{
				Name: name,
				Type: "git",
				Source: atc.Source{
					"uri":         gitUri,
					"private_key": privateKey,
				},
			},
		},
	}
	assert.Equal(t, expected, pipe.Render(manifest))
}

func TestRenderRunTask(t *testing.T) {
	manifest := model.Manifest{}
	manifest.Repo.Uri = "git@github.com:/springernature/foo.git"
	manifest.Tasks = []model.Task{
		model.Run{
			Script: "./yolo.sh",
			Image:  "imagename:TAG",
			Vars: map[string]string{
				"VAR1": "Value1",
				"VAR2": "Value2",
			},
		},
	}

	expected := atc.JobConfig{
		Name:   "1. Run ./yolo.sh",
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{Get: manifest.Repo.GetName(), Trigger: true},
			atc.PlanConfig{Task: "./yolo.sh", TaskConfig: &atc.TaskConfig{
				Platform: "linux",
				Params: map[string]string{
					"VAR1": "Value1",
					"VAR2": "Value2",
				},
				ImageResource: &atc.ImageResource{
					Type: "docker-image",
					Source: atc.Source{
						"repository": "imagename",
						"tag":        "TAG",
					},
				},
				Run: atc.TaskRunConfig{
					Path: "/bin/sh",
					Dir:  manifest.Repo.GetName(),
					Args: []string{"-exc", fmt.Sprintf("././yolo.sh")},
				},
				Inputs: []atc.TaskInputConfig{
					{Name: manifest.Repo.GetName()},
				},
			}},
		}}

	assert.Equal(t, expected, pipe.Render(manifest).Jobs[0])
}

func TestRenderDockerPushTask(t *testing.T) {
	manifest := model.Manifest{}
	manifest.Repo.Uri = "git@github.com:/springernature/foo.git"

	username := "halfpipe"
	password := "secret"
	repo := "halfpipe/halfpipe-cli"
	manifest.Tasks = []model.Task{
		model.DockerPush{
			Username: username,
			Password: password,
			Repo:     repo,
		},
	}

	expectedResource := atc.ResourceConfig{
		Name: "1. Docker Registry",
		Type: "docker-image",
		Source: atc.Source{
			"username":   username,
			"password":   password,
			"repository": repo,
		},
	}

	expectedJobConfig := atc.JobConfig{
		Name:   "1. docker-push",
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{Get: manifest.Repo.GetName(), Trigger: true},
			atc.PlanConfig{Put: "1. Docker Registry", Params: atc.Params{"build": manifest.Repo.GetName()}},
		},
	}

	// First resource will always be the git resource.
	assert.Equal(t, expectedResource, pipe.Render(manifest).Resources[1])
	assert.Equal(t, expectedJobConfig, pipe.Render(manifest).Jobs[0])
}

func TestRenderWithTriggerTrueAndPassedOnPreviousTask(t *testing.T) {

	task1 := model.Run{Script: "asd.sh"}
	task2 := model.DeployCF{}
	task3 := model.DockerPush{}
	manifest := model.Manifest{
		Repo: model.Repo{
			Uri: "git@github.com:/springernature/foo.git",
		},
		Tasks: []model.Task{
			task1,
			task2,
			task3,
		},
	}

	config := pipe.Render(manifest)
	assert.Equal(t, config.Jobs[0].Plan[0], atc.PlanConfig{
		Get:     manifest.Repo.GetName(),
		Trigger: true,
		Passed:  nil,
	})

	assert.Equal(t, config.Jobs[1].Plan[0], atc.PlanConfig{
		Get:     manifest.Repo.GetName(),
		Trigger: true,
		Passed: []string{
			config.Jobs[0].Name,
		}})

	assert.Equal(t, config.Jobs[2].Plan[0], atc.PlanConfig{
		Get:     manifest.Repo.GetName(),
		Trigger: true,
		Passed: []string{
			config.Jobs[1].Name,
		}})

}

func TestToString(t *testing.T) {
	man := model.Manifest{}
	man.Repo.Uri = "repo.git"

	actual, err := ToString(pipe.Render(man))
	expected := "uri: repo.git"

	assert.Nil(t, err)
	assert.Contains(t, actual, expected)
}
