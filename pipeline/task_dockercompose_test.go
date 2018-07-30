package pipeline

import (
	"strings"
	"testing"

	"github.com/concourse/atc"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/dockercompose"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

var dockerComposeImageResource = atc.ImageResource{
	Type: "docker-image",
	Source: atc.Source{
		"repository": strings.Split(config.DockerComposeImage, ":")[0],
		"tag":        strings.Split(config.DockerComposeImage, ":")[1],
		"username":   "_json_key",
		"password":   "((gcr.private_key))",
	},
}

func TestRenderDockerComposeTask(t *testing.T) {
	p := testPipeline()

	dockerCompose := dockercompose.DockerCompose{
		Services: []dockercompose.Service{
			{Name: "app", Image: "eu.gcr.io/halfpipe-io/golang:latest"},
			{Name: "db", Image: "mydb"},
			{Name: "no-image", Image: ""},
		},
	}

	p.readDockerCompose = func() (dockercompose.DockerCompose, error) {
		return dockerCompose, nil
	}

	service := "asdf"
	man := manifest.Manifest{
		Repo: manifest.Repo{
			URI:      "git@git:user/repo",
			BasePath: "base.path",
		},
		Tasks: []manifest.Task{
			manifest.DockerCompose{
				Name:    "",
				Service: service,
				Vars: manifest.Vars{
					"VAR1": "Value1",
					"VAR2": "Value2",
				},
			},
		},
	}

	expectedVars := map[string]string{
		"VAR1":            "Value1",
		"VAR2":            "Value2",
		"GCR_PRIVATE_KEY": "((gcr.private_key))",
	}

	expectedJob := atc.JobConfig{
		Name:   "docker-compose",
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{Aggregate: &atc.PlanSequence{
				atc.PlanConfig{Get: gitDir, Trigger: true},
				atc.PlanConfig{Get: dockerCompose.Services[0].ResourceName(), Params: atc.Params{"save": true}},
				atc.PlanConfig{Get: dockerCompose.Services[1].ResourceName(), Params: atc.Params{"save": true}},
			}},
			atc.PlanConfig{
				Task:       "docker-compose",
				Privileged: true,
				TaskConfig: &atc.TaskConfig{
					Platform:      "linux",
					Params:        expectedVars,
					ImageResource: &dockerComposeImageResource,
					Run: atc.TaskRunConfig{
						Path: "docker.sh",
						Dir:  gitDir + "/base.path",
						Args: runScriptArgs(dockerComposeScript(service, expectedVars, ""), false, "", false, nil, "../.git/ref"),
					},
					Inputs: []atc.TaskInputConfig{
						{Name: gitDir},
						{Name: dockerCompose.Services[0].ResourceName(), Path: "docker-images/" + dockerCompose.Services[0].ResourceName()},
						{Name: dockerCompose.Services[1].ResourceName(), Path: "docker-images/" + dockerCompose.Services[1].ResourceName()},
					},
					Caches: config.CacheDirs,
				}},
		}}

	actualPipeline := p.Render(man)
	assert.Equal(t, expectedJob, actualPipeline.Jobs[0])
}

func TestRenderDockerComposeTaskWithCommand(t *testing.T) {
	p := testPipeline()

	man := manifest.Manifest{
		Repo: manifest.Repo{
			URI:      "git@git:user/repo",
			BasePath: "base.path",
		},
		Tasks: []manifest.Task{
			manifest.DockerCompose{
				Name:    "",
				Service: "app",
				Command: "/usr/bin/a-command",
				Vars: manifest.Vars{
					"VAR1": "Value 1",
					"VAR2": "Value 2",
				},
			},
		},
	}

	expectedVars := map[string]string{
		"VAR1":            "Value 1",
		"VAR2":            "Value 2",
		"GCR_PRIVATE_KEY": "((gcr.private_key))",
	}

	expectedArgs := runScriptArgs(dockerComposeScript("app", expectedVars, "/usr/bin/a-command"), false, "", false, nil, "../.git/ref")
	assert.Equal(t, expectedArgs, p.Render(man).Jobs[0].Plan[1].TaskConfig.Run.Args)
}

func TestDockerComposeRunJobIsPrivileged(t *testing.T) {
	p := testPipeline()

	man := manifest.Manifest{
		Repo: manifest.Repo{
			URI:      "git@git:user/repo",
			BasePath: "base.path",
		},
		Tasks: []manifest.Task{
			manifest.DockerCompose{
				Name:             "",
				RestoreArtifacts: true,
			},
		},
	}

	step := p.Render(man).Jobs[0].Plan[1]
	assert.Equal(t, "docker-compose", step.Task)
	assert.True(t, step.Privileged)

}

func TestAddResourcesForDockerComposeImages(t *testing.T) {
	p := testPipeline()

	dockerCompose := dockercompose.DockerCompose{
		Services: []dockercompose.Service{
			{Name: "app", Image: "eu.gcr.io/halfpipe-io/golang:latest"},
			{Name: "db", Image: "mydb"},
			{Name: "no-image", Image: ""},
		},
	}

	p.readDockerCompose = func() (dockercompose.DockerCompose, error) {
		return dockerCompose, nil
	}

	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DockerCompose{},
		},
	}

	actualPipeline := p.Render(man)
	actualResource0, _ := actualPipeline.Resources.Lookup(dockerCompose.Services[0].ResourceName())
	actualResource1, _ := actualPipeline.Resources.Lookup(dockerCompose.Services[1].ResourceName())

	expectedResources := resourcesFromDockerCompose(dockerCompose)

	assert.Equal(t, expectedResources[0], actualResource0)
	assert.Equal(t, expectedResources[1], actualResource1)
}

func TestDoesNotAddResourcesForDockerComposeImagesWhenThereIsNoDockerComposeTask(t *testing.T) {
	p := testPipeline()

	dockerCompose := dockercompose.DockerCompose{
		Services: []dockercompose.Service{
			{Name: "app", Image: "eu.gcr.io/halfpipe-io/golang:latest"},
			{Name: "db", Image: "mydb"},
			{Name: "no-image", Image: ""},
		},
	}

	p.readDockerCompose = func() (dockercompose.DockerCompose, error) {
		return dockerCompose, nil
	}

	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.Run{},
		},
	}

	actualPipeline := p.Render(man)
	_, resource0exists := actualPipeline.Resources.Lookup(dockerCompose.Services[0].ResourceName())
	_, resource1exists := actualPipeline.Resources.Lookup(dockerCompose.Services[1].ResourceName())
	assert.False(t, resource0exists)
	assert.False(t, resource1exists)
}
