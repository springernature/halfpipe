package concourse

import (
	"fmt"
	"github.com/concourse/concourse/atc"
	"github.com/springernature/halfpipe/manifest"
	"path"
	"strings"
)

func (c Concourse) dockerPushJob(task manifest.DockerPush, basePath string, ociBuild bool) atc.JobConfig {
	resourceName := manifest.DockerTrigger{Image: task.Image}.GetTriggerName()
	if task.RestoreArtifacts {
		return dockerPushJobWithRestoreArtifacts(task, resourceName, basePath, ociBuild)
	}
	return dockerPushJobWithoutRestoreArtifacts(task, resourceName, basePath, ociBuild)
}

func dockerPushJobWithoutRestoreArtifacts(task manifest.DockerPush, resourceName string, basePath string, ociBuild bool) atc.JobConfig {
	fullBasePath := path.Join(gitDir, basePath)

	var steps []atc.Step
	if ociBuild {

		params := atc.TaskEnv{
			"CONTEXT": fullBasePath,
		}

		for k, v := range convertVars(task.Vars) {
			params[fmt.Sprintf("BUILD_ARG_%s", k)] = fmt.Sprintf("%s", v)
		}

		buildStep := &atc.TaskStep{
			Name:       "build",
			Privileged: true,
			Config: &atc.TaskConfig{
				Platform: "linux",
				ImageResource: &atc.ImageResource{
					Type: "registry-image",
					Source: atc.Source{
						"repository": "concourse/oci-build-task",
					},
				},
				Params: params,
				Run: atc.TaskRunConfig{
					Path: "build",
				},
				Inputs: []atc.TaskInputConfig{
					{Name: gitDir},
				},
				Outputs: []atc.TaskOutputConfig{
					{Name: "image"},
				},
			},
		}

		putStep := &atc.PutStep{
			Name: resourceName,
			Params: atc.Params{
				"image":   "image/image.tar",
				"version": "6.6.6",
			},
		}
		steps = append(steps, stepWithAttemptsAndTimeout(buildStep, task.GetAttempts(), task.GetTimeout()))
		steps = append(steps, stepWithAttemptsAndTimeout(putStep, task.GetAttempts(), task.GetTimeout()))
	} else {
		step := &atc.PutStep{
			Name: resourceName,
			Params: atc.Params{
				"build":         path.Join(fullBasePath, task.BuildPath),
				"dockerfile":    path.Join(fullBasePath, task.DockerfilePath),
				"tag_as_latest": true,
				"tag_file":      task.GetTagPath(fullBasePath),
				"build_args":    convertVars(task.Vars),
			},
		}
		steps = append(steps, stepWithAttemptsAndTimeout(step, task.GetAttempts(), task.GetTimeout()))
	}

	return atc.JobConfig{
		Name:         task.GetName(),
		Serial:       true,
		PlanSequence: steps,
	}
}

func dockerPushJobWithRestoreArtifacts(task manifest.DockerPush, resourceName string, basePath string, ociBuild bool) atc.JobConfig {
	copyArtifact := &atc.TaskStep{
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
	}

	fullBasePath := path.Join(dockerBuildTmpDir, basePath)
	put := &atc.PutStep{
		Name: resourceName,
		Params: atc.Params{
			"build":         path.Join(dockerBuildTmpDir, basePath, task.BuildPath),
			"dockerfile":    path.Join(dockerBuildTmpDir, basePath, task.DockerfilePath),
			"tag_as_latest": true,
			"tag_file":      task.GetTagPath(fullBasePath),
			"build_args":    convertVars(task.Vars),
		},
	}

	return atc.JobConfig{
		Name:   task.GetName(),
		Serial: true,
		PlanSequence: []atc.Step{
			stepWithAttemptsAndTimeout(copyArtifact, task.GetAttempts(), task.Timeout),
			stepWithAttemptsAndTimeout(put, task.GetAttempts(), task.Timeout),
		},
	}
}
