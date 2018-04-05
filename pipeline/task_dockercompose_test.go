package pipeline

import (
	"strings"
	"testing"

	"github.com/concourse/atc"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestRenderDockerComposeTask(t *testing.T) {
	p := testPipeline()

	man := manifest.Manifest{
		Repo: manifest.Repo{
			URI:      "git@git:user/repo",
			BasePath: "base.path",
		},
		Tasks: []manifest.Task{
			manifest.DockerCompose{
				Name: "",
				Vars: manifest.Vars{
					"VAR1": "Value1",
					"VAR2": "Value2",
				},
			},
		},
	}

	expectedJob := atc.JobConfig{
		Name:   "docker-compose",
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{Get: gitDir, Trigger: true},
			atc.PlanConfig{
				Task:       "run",
				Privileged: true,
				TaskConfig: &atc.TaskConfig{
					Platform: "linux",
					Params: map[string]string{
						"VAR1": "Value1",
						"VAR2": "Value2",
					},
					ImageResource: &atc.ImageResource{
						Type: "docker-image",
						Source: atc.Source{
							"repository": strings.Split(config.DockerComposeImage, ":")[0],
							"tag":        strings.Split(config.DockerComposeImage, ":")[1],
						},
					},
					Run: atc.TaskRunConfig{
						Path: "/bin/sh",
						Dir:  gitDir + "/base.path",
						Args: runScriptArgs(dockerComposeScript(), "", nil, "../.git/ref"),
					},
					Inputs: []atc.TaskInputConfig{
						{Name: gitDir},
					},
				}},
		}}

	assert.Equal(t, expectedJob, p.Render(man).Jobs[0])
}
