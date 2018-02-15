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
		Name:   "./yolo.sh",
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

func TestToString(t *testing.T) {
	man := model.Manifest{}
	man.Repo.Uri = "repo.git"

	actual, err := ToString(pipe.Render(man))
	expected := "uri: repo.git"

	assert.Nil(t, err)
	assert.Contains(t, actual, expected)
}
