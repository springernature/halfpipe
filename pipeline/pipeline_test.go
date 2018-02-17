package pipeline

import (
	"fmt"
	"testing"

	"github.com/concourse/atc"
	"github.com/springernature/halfpipe/model"
	"github.com/stretchr/testify/assert"
)

func testPipeline() Pipeline {
	return Pipeline{}
}

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
	assert.Equal(t, expected, testPipeline().Render(manifest))
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
	assert.Equal(t, expected, testPipeline().Render(manifest))
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
		Name:   "run yolo.sh",
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
					Args: []string{"-exc", fmt.Sprintf("./yolo.sh")},
				},
				Inputs: []atc.TaskInputConfig{
					{Name: manifest.Repo.GetName()},
				},
			}},
		}}

	assert.Equal(t, expected, testPipeline().Render(manifest).Jobs[0])
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
			atc.PlanConfig{Get: manifest.Repo.GetName(), Trigger: true},
			atc.PlanConfig{Put: "Docker Registry", Params: atc.Params{"build": manifest.Repo.GetName()}},
		},
	}

	// First resource will always be the git resource.
	assert.Equal(t, expectedResource, testPipeline().Render(manifest).Resources[1])
	assert.Equal(t, expectedJobConfig, testPipeline().Render(manifest).Jobs[0])
}

func TestRenderWithTriggerTrueAndPassedOnPreviousTask(t *testing.T) {
	manifest := model.Manifest{
		Tasks: []model.Task{
			model.Run{Script: "asd.sh"},
			model.DeployCF{},
			model.DockerPush{},
		},
	}
	config := testPipeline().Render(manifest)

	assert.Nil(t, config.Jobs[0].Plan[0].Passed)
	assert.Equal(t, config.Jobs[1].Plan[0].Passed[0], config.Jobs[0].Name)
	assert.Equal(t, config.Jobs[2].Plan[0].Passed[0], config.Jobs[1].Name)
}

func TestToString(t *testing.T) {
	man := model.Manifest{}
	man.Repo.Uri = "repo.git"

	actual, err := ToString(testPipeline().Render(man))
	expected := "uri: repo.git"

	assert.Nil(t, err)
	assert.Contains(t, actual, expected)
}

func TestGeneratesUniqueNamesForJobsAndResources(t *testing.T) {
	manifest := model.Manifest{
		Repo: model.Repo{Uri: "https://github.com/springernature/halfpipe.git"},
		Tasks: []model.Task{
			model.Run{Script: "asd.sh"},
			model.Run{Script: "asd.sh"},
			model.Run{Script: "asd.sh"},
			model.Run{Script: "fgh.sh"},
			model.DeployCF{},
			model.DeployCF{},
			model.DeployCF{},
			model.DockerPush{},
			model.DockerPush{},
			model.DockerPush{},
		},
	}
	config := testPipeline().Render(manifest)

	expectedJobNames := []string{
		"run asd.sh",
		"run asd.sh (1)",
		"run asd.sh (2)",
		"run fgh.sh",
		"deploy-cf",
		"deploy-cf (1)",
		"deploy-cf (2)",
		"docker-push",
		"docker-push (1)",
		"docker-push (2)",
	}

	expectedResourceNames := []string{
		"halfpipe",
		"Cloud Foundry",
		"Cloud Foundry (1)",
		"Cloud Foundry (2)",
		"Docker Registry",
		"Docker Registry (1)",
		"Docker Registry (2)",
	}

	assert.Len(t, config.Jobs, len(expectedJobNames))
	assert.Len(t, config.Resources, len(expectedResourceNames))

	for i, name := range expectedJobNames {
		assert.Equal(t, name, config.Jobs[i].Name)
	}

	for i, name := range expectedResourceNames {
		assert.Equal(t, name, config.Resources[i].Name)
	}

}
