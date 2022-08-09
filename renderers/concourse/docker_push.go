package concourse

import (
	"fmt"
	"github.com/concourse/concourse/atc"
	"github.com/springernature/halfpipe/manifest"
	"path"
	"strings"
)

func (c Concourse) dockerPushJob(task manifest.DockerPush, basePath string, ociBuild bool) atc.JobConfig {
	var steps []atc.Step
	resourceName := manifest.DockerTrigger{Image: task.Image}.GetTriggerName()

	fullBasePath := path.Join(gitDir, basePath)
	if task.RestoreArtifacts {
		fullBasePath = path.Join(dockerBuildTmpDir, basePath)
	}

	steps = append(steps, restoreArtifacts(task)...)
	steps = append(steps, buildAndPush(task, resourceName, ociBuild, fullBasePath, task.RestoreArtifacts)...)

	return atc.JobConfig{
		Name:         task.GetName(),
		Serial:       true,
		PlanSequence: steps,
	}
}

func restoreArtifacts(task manifest.DockerPush) []atc.Step {
	if task.RestoreArtifacts {
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
		return append([]atc.Step{}, stepWithAttemptsAndTimeout(copyArtifact, task.GetAttempts(), task.Timeout))
	}
	return []atc.Step{}
}

func buildAndPushOci(task manifest.DockerPush, resourceName string, fullBasePath string, restore bool) []atc.Step {
	var steps []atc.Step

	params := atc.TaskEnv{
		"CONTEXT":    path.Join(fullBasePath, task.BuildPath),
		"DOCKERFILE": path.Join(fullBasePath, task.DockerfilePath),
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
	if restore {
		buildStep.Config.Inputs = append(buildStep.Config.Inputs, atc.TaskInputConfig{Name: dockerBuildTmpDir})
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
	return steps
}

func buildAndPush(task manifest.DockerPush, resourceName string, ociBuild bool, fullBasePath string, restore bool) []atc.Step {
	if ociBuild {
		return buildAndPushOci(task, resourceName, fullBasePath, restore)
	}

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
	return append([]atc.Step{}, stepWithAttemptsAndTimeout(step, task.GetAttempts(), task.GetTimeout()))
}
