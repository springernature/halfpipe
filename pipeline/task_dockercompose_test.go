package pipeline

import (
	"strings"
	"testing"

	"github.com/concourse/concourse/atc"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

var dockerComposeImageResource = atc.ImageResource{
	Type: "registry-image",
	Source: atc.Source{
		"repository": config.DockerRegistry + strings.Split(config.DockerComposeImage, ":")[0],
		"tag":        strings.Split(config.DockerComposeImage, ":")[1],
		"username":   "_json_key",
		"password":   "((halfpipe-gcr.private_key))",
	},
}

func TestRenderDockerComposeTask(t *testing.T) {
	p := testPipeline()

	service := "asdf"
	dockerComposeTask := manifest.DockerCompose{
		Name:    "docker-compose",
		Service: service,
		Vars: manifest.Vars{
			"VAR1": "Value1",
			"VAR2": "Value2",
		},
	}
	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.GitTrigger{
				URI:      "git@git:user/repo",
				BasePath: "base.path",
			},
		},
		Tasks: []manifest.Task{
			dockerComposeTask,
		},
	}

	expectedVars := map[string]string{
		"VAR1":                "Value1",
		"VAR2":                "Value2",
		"GCR_PRIVATE_KEY":     "((halfpipe-gcr.private_key))",
		"HALFPIPE_CACHE_TEAM": "",
	}

	expectedJob := atc.JobConfig{
		Name:   "docker-compose",
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{InParallel: &atc.InParallelConfig{Steps: atc.PlanSequence{atc.PlanConfig{Get: gitName, Trigger: true, Attempts: gitGetAttempts}}}},
			atc.PlanConfig{
				Attempts:   1,
				Task:       "docker-compose",
				Privileged: true,
				TaskConfig: &atc.TaskConfig{
					Platform:      "linux",
					Params:        expectedVars,
					ImageResource: &dockerComposeImageResource,
					Run: atc.TaskRunConfig{
						Path: "docker.sh",
						Dir:  gitDir + "/base.path",
						Args: runScriptArgs(dockerComposeToRunTask(dockerComposeTask, man), man, false, man.Triggers.GetGitTrigger().BasePath),
					},
					Inputs: []atc.TaskInputConfig{
						{Name: gitName},
					},
					Caches: config.CacheDirs,
				}},
		}}

	assert.Equal(t, expectedJob, p.Render(man).Jobs[0])
}

func TestRenderDockerComposeTaskWithCommand(t *testing.T) {
	p := testPipeline()

	dockerComposeTask := manifest.DockerCompose{
		Name:    "docker-compose",
		Service: "app",
		Command: "/usr/bin/a-command",
		Vars: manifest.Vars{
			"VAR1": "Value 1",
			"VAR2": "Value 2",
		},
	}
	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.GitTrigger{
				URI:      "git@git:user/repo",
				BasePath: "base.path",
			},
		},
		Tasks: []manifest.Task{
			dockerComposeTask,
		},
	}

	expectedVars := map[string]string{
		"VAR1":                "Value 1",
		"VAR2":                "Value 2",
		"GCR_PRIVATE_KEY":     "((halfpipe-gcr.private_key))",
		"HALFPIPE_CACHE_TEAM": "",
	}

	expectedJob := atc.JobConfig{
		Name:   "docker-compose",
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{InParallel: &atc.InParallelConfig{Steps: atc.PlanSequence{atc.PlanConfig{Get: gitName, Trigger: true, Attempts: gitGetAttempts}}}},
			atc.PlanConfig{
				Attempts:   1,
				Task:       "docker-compose",
				Privileged: true,
				TaskConfig: &atc.TaskConfig{
					Platform:      "linux",
					Params:        expectedVars,
					ImageResource: &dockerComposeImageResource,
					Run: atc.TaskRunConfig{
						Path: "docker.sh",
						Dir:  gitDir + "/base.path",
						Args: runScriptArgs(dockerComposeToRunTask(dockerComposeTask, man), man, false, man.Triggers.GetGitTrigger().BasePath),
					},
					Inputs: []atc.TaskInputConfig{
						{Name: gitName},
					},
					Caches: config.CacheDirs,
				}},
		}}

	assert.Equal(t, expectedJob, p.Render(man).Jobs[0])
}

func TestDockerComposeRunJobIsPrivileged(t *testing.T) {
	p := testPipeline()

	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DockerCompose{
				Name:             "docker-compose",
				RestoreArtifacts: true,
			},
		},
	}

	step := p.Render(man).Jobs[0].Plan[2]
	assert.Equal(t, "docker-compose", step.Task)
	assert.True(t, step.Privileged)

}
