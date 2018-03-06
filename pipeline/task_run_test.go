package pipeline

import (
	"fmt"
	"testing"

	"github.com/concourse/atc"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestRenderRunTask(t *testing.T) {
	man := manifest.Manifest{}
	man.Repo.Uri = "git@github.com:/springernature/foo.git"
	man.Tasks = []manifest.Task{
		manifest.Run{
			Script: "./yolo.sh",
			Docker: manifest.Docker{
				Image:    "imagename:TAG",
				Username: "",
				Password: "",
			},
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
			atc.PlanConfig{Get: man.Repo.GetName(), Trigger: true},
			atc.PlanConfig{Task: "./yolo.sh", Privileged: true, TaskConfig: &atc.TaskConfig{
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
					Dir:  man.Repo.GetName(),
					Args: []string{"-ec", fmt.Sprintf("./yolo.sh")},
				},
				Inputs: []atc.TaskInputConfig{
					{Name: man.Repo.GetName()},
				},
			}},
		}}

	assert.Equal(t, expected, testPipeline().Render(man).Jobs[0])
}
func TestRenderRunTaskWithPrivateRepo(t *testing.T) {
	man := manifest.Manifest{}
	man.Repo.Uri = "git@github.com:/springernature/foo.git"
	man.Tasks = []manifest.Task{
		manifest.Run{
			Script: "./yolo.sh",
			Docker: manifest.Docker{
				Image:    "imagename:TAG",
				Username: "user",
				Password: "pass",
			},
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
			atc.PlanConfig{Get: man.Repo.GetName(), Trigger: true},
			atc.PlanConfig{Task: "./yolo.sh", Privileged: true, TaskConfig: &atc.TaskConfig{
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
						"username":   "user",
						"password":   "pass",
					},
				},
				Run: atc.TaskRunConfig{
					Path: "/bin/sh",
					Dir:  man.Repo.GetName(),
					Args: []string{"-ec", fmt.Sprintf("./yolo.sh")},
				},
				Inputs: []atc.TaskInputConfig{
					{Name: man.Repo.GetName()},
				},
			}},
		}}

	assert.Equal(t, expected, testPipeline().Render(man).Jobs[0])
}

func TestRenderRunTaskFromHalfpipeNotInRoot(t *testing.T) {
	man := manifest.Manifest{}
	basePath := "subapp"
	man.Repo.Uri = "git@github.com:/springernature/foo.git"
	man.Repo.BasePath = basePath

	man.Tasks = []manifest.Task{
		manifest.Run{
			Script: "./yolo.sh",
			Docker: manifest.Docker{
				Image: "imagename:TAG",
			},
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
			atc.PlanConfig{Get: man.Repo.GetName(), Trigger: true},
			atc.PlanConfig{Task: "./yolo.sh", Privileged: true, TaskConfig: &atc.TaskConfig{
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
					Dir:  man.Repo.GetName() + "/" + basePath,
					Args: []string{"-ec", fmt.Sprintf("./yolo.sh")},
				},
				Inputs: []atc.TaskInputConfig{
					{Name: man.Repo.GetName()},
				},
			}},
		}}

	assert.Equal(t, expected, testPipeline().Render(man).Jobs[0])
}
