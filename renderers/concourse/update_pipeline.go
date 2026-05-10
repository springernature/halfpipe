package concourse

import (
	"path"

	"github.com/concourse/concourse/atc"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
)

func (c Concourse) updateJobConfig(task manifest.Update, pipelineName string, basePath string) atc.JobConfig {
	update := &atc.TaskStep{
		Name: "update",
		Config: &atc.TaskConfig{
			Platform: "linux",
			Params: map[string]string{
				"CONCOURSE_URL":      config.VaultSecrets.ConcourseURL,
				"CONCOURSE_PASSWORD": config.VaultSecrets.ConcoursePassword,
				"CONCOURSE_TEAM":     config.VaultSecrets.ConcourseTeam,
				"CONCOURSE_USERNAME": config.VaultSecrets.ConcourseUsername,
				"PIPELINE_NAME":      pipelineName,
				"HALFPIPE_DOMAIN":    config.Domain,
				"HALFPIPE_PROJECT":   config.Project,
				"HALFPIPE_FILE_PATH": c.halfpipeFilePath,
			},
			ImageResource: imageResource(manifest.Docker{
				Image:    path.Join(config.DockerRegistry, "halfpipe-auto-update"),
				Username: "oauth2accesstoken",
				Password: config.VaultSecrets.GARToken,
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
		NoGet:  true,
	}

	steps := []atc.Step{
		stepWithDefaultAttemptsAndTimeout(update),
		stepWithDefaultAttemptsAndTimeout(bumpVersion),
	}

	if task.TagRepo {
		bumpVersion.NoGet = false

		tagRepo := &atc.PutStep{
			Name:     "tag-git-repository",
			Resource: manifest.GitTrigger{}.GetTriggerName(),
			Params: atc.Params{
				"only_tag":   true,
				"repository": manifest.GitTrigger{}.GetTriggerName(),
				"tag":        "version/version",
				"tag_prefix": pipelineName + "/v",
			},
			NoGet: true,
		}
		steps = append(steps, stepWithDefaultAttemptsAndTimeout(tagRepo))
	}

	return atc.JobConfig{
		Name:         task.GetName(),
		Serial:       true,
		PlanSequence: steps,
	}
}
