package concourse

import (
	"fmt"
	"github.com/concourse/concourse/atc"
	"github.com/springernature/halfpipe/manifest"
	"path"
	"strings"
)

func (p pipeline) dockerPushJob(task manifest.DockerPush, basePath string) atc.JobConfig {
	resourceName := manifest.DockerTrigger{Image: task.Image}.GetTriggerName()
	if task.RestoreArtifacts {
		return dockerPushJobWithRestoreArtifacts(task, resourceName, basePath)
	}
	return dockerPushJobWithoutRestoreArtifacts(task, resourceName, basePath)
}

func dockerPushJobWithoutRestoreArtifacts(task manifest.DockerPush, resourceName string, basePath string) atc.JobConfig {
	timeoutStep := atc.TimeoutStep{
		Step: &atc.PutStep{
			Name: resourceName,
			Params: atc.Params{
				"build":         path.Join(gitDir, basePath, task.BuildPath),
				"dockerfile":    path.Join(gitDir, basePath, task.DockerfilePath),
				"tag_as_latest": true,
				"tag_file":      task.GetTagPath(),
				"build_args":    convertVars(task.Vars),
			},
		},
		Duration: task.GetTimeout(),
	}

	step := atc.Step{}
	if task.GetAttempts() > 1 {
		step.Config = &atc.RetryStep{
			Step:     &timeoutStep,
			Attempts: task.GetAttempts(),
		}
	} else {
		step.Config = &timeoutStep
	}

	return atc.JobConfig{
		Name:         task.GetName(),
		Serial:       true,
		PlanSequence: []atc.Step{step},
	}
}

func dockerPushJobWithRestoreArtifacts(task manifest.DockerPush, resourceName string, basePath string) atc.JobConfig {
	copyArtifact := atc.Step{
		Config: &atc.TimeoutStep{
			Step: &atc.TaskStep{
				Name: "copying-git-repo-and-artifacts-to-a-temporary-build-dir",
				Config: &atc.TaskConfig{
					Platform: "linux",
					ImageResource: &atc.ImageResource{
						Type: "docker-image",
						Source: atc.Source{
							"repository": "alpine",
						},
					},
					Run: atc.TaskRunConfig{
						Path: "/bin/sh",
						Args: []string{"-c", strings.Join([]string{
							fmt.Sprintf("cp -r %s/. %s", gitDir, dockerBuildTmpDir),
							fmt.Sprintf("cp -r %s/. %s", artifactsInDir, dockerBuildTmpDir),
						}, "\n")},
					},
					Inputs: []atc.TaskInputConfig{
						{Name: gitDir},
						{Name: artifactsName},
					},
					Outputs: []atc.TaskOutputConfig{
						{Name: dockerBuildTmpDir},
					},
				},
			},
			Duration: task.GetTimeout(),
		},
	}

	putStep := atc.PutStep{
		Name: resourceName,
		Params: atc.Params{
			"build":         path.Join(dockerBuildTmpDir, basePath, task.BuildPath),
			"dockerfile":    path.Join(dockerBuildTmpDir, basePath, task.DockerfilePath),
			"tag_as_latest": true,
			"tag_file":      task.GetTagPath(),
			"build_args":    convertVars(task.Vars),
		},
	}
	timeoutStep := atc.TimeoutStep{
		Step:     &putStep,
		Duration: task.GetTimeout(),
	}

	step := atc.Step{}
	if task.GetAttempts() > 1 {
		retryStep := atc.RetryStep{
			Step:     &timeoutStep,
			Attempts: task.GetAttempts(),
		}
		step.Config = &retryStep
	} else {
		step.Config = &timeoutStep
	}

	return atc.JobConfig{
		Name:         task.GetName(),
		Serial:       true,
		PlanSequence: []atc.Step{copyArtifact, step},
	}
}
