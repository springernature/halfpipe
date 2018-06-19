package pipeline

import (
	"path"

	"github.com/concourse/atc"
	"github.com/springernature/halfpipe/manifest"
)

func (p pipeline) addUpdatePipelineJob(cfg *atc.Config, man manifest.Manifest, failurePlan *atc.PlanConfig) {
	if man.AutoUpdate {
		job := updatePipelineJobConfig(man.Repo.BasePath)
		job.Failure = failurePlan
		cfg.Jobs = append(cfg.Jobs, job)
	}
	return
}

func updatePipelineJobConfig(basePath string) atc.JobConfig {
	return atc.JobConfig{
		Name:   "Update Pipeline",
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{Get: gitDir, Trigger: true},
			atc.PlanConfig{
				Task:       "Update Pipeline",
				Privileged: false,
				TaskConfig: &atc.TaskConfig{
					Platform: "linux",
					Params: map[string]string{
						"CONCOURSE_TEAM":     "((concourse.team))",
						"CONCOURSE_PASSWORD": "((concourse.password))",
						"CONCOURSE_USERNAME": "((concourse.username))",
					},
					ImageResource: &atc.ImageResource{
						Type: "docker-image",
						Source: atc.Source{
							"password":   "((gcr.private_key))",
							"repository": "eu.gcr.io/halfpipe-io/halfpipe-hacky-incept",
							"tag":        "latest",
							"username":   "_json_key",
						},
					},
					Run: atc.TaskRunConfig{
						Path: "/bin/incept",
						Dir:  path.Join(gitDir, basePath),
					},
					Inputs: []atc.TaskInputConfig{
						{Name: gitDir},
					},
				},
			},
		},
	}
}
