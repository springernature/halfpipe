package concourse

import (
	"github.com/concourse/concourse/atc"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
	"path"
)

func (c Concourse) updateJobConfig(task manifest.Update, pipelineName string, basePath string) atc.JobConfig {

	const updateTaskAttempts = 2
	const updateTaskTimeout = "15m"

	update := &atc.TaskStep{
		Name: "update",
		Config: &atc.TaskConfig{
			Platform: "linux",
			Params: map[string]string{
				"CONCOURSE_URL":      "((concourse.url))",
				"CONCOURSE_PASSWORD": "((concourse.password))",
				"CONCOURSE_TEAM":     "((concourse.team))",
				"CONCOURSE_USERNAME": "((concourse.username))",
				"PIPELINE_NAME":      pipelineName,
				"HALFPIPE_DOMAIN":    config.Domain,
				"HALFPIPE_PROJECT":   config.Project,
			},
			ImageResource: c.imageResource(manifest.Docker{
				Image:    config.DockerRegistry + "halfpipe-auto-update",
				Username: "_json_key",
				Password: "((halfpipe-gcr.private_key))",
			}),
			Run: atc.TaskRunConfig{
				Path: "update-pipeline",
				Dir:  path.Join(gitDir, basePath),
			},
			Inputs: []atc.TaskInputConfig{
				{Name: manifest.GitTrigger{}.GetTriggerName()},
			},
		},
	}

	bumpVersion := &atc.PutStep{
		Name:   versionName,
		Params: atc.Params{"bump": "minor"},
	}

	return atc.JobConfig{
		Name:   task.GetName(),
		Serial: true,
		PlanSequence: []atc.Step{
			stepWithAttemptsAndTimeout(update, updateTaskAttempts, updateTaskTimeout),
			stepWithAttemptsAndTimeout(bumpVersion, updateTaskAttempts, updateTaskTimeout),
		},
	}
}
