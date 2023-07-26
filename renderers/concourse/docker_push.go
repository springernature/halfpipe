package concourse

import (
	"fmt"
	"github.com/concourse/concourse/atc"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/renderers/shared"
	"path"
	"strings"
)

const tagList_Dir = "tagList"

var tagListFile = path.Join(tagList_Dir, "tagList")

func (c Concourse) dockerPushJob(task manifest.DockerPush, basePath string, man manifest.Manifest) atc.JobConfig {
	var steps []atc.Step
	resourceName := manifest.DockerTrigger{Image: task.Image}.GetTriggerName()

	fullBasePath := path.Join(gitDir, basePath)
	if task.RestoreArtifacts {
		fullBasePath = path.Join(dockerBuildTmpDir, basePath)
	}

	steps = append(steps, restoreArtifacts(task)...)
	steps = append(steps, createTagList(task, man.FeatureToggles.UpdatePipeline())...)
	steps = append(steps, buildAndPush(task, resourceName, fullBasePath)...)

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
					Type: "registry-image",
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

func createTagList(task manifest.DockerPush, updatePipeline bool) []atc.Step {
	gitRefFile := path.Join(gitDir, ".git", "ref")
	versionFile := path.Join(versionName, "version")

	createTagList := &atc.TaskStep{
		Name: "create-tag-list",
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
					fmt.Sprintf("GIT_REF=`[ -f %s ] && cat %s || true`", gitRefFile, gitRefFile),
					fmt.Sprintf("VERSION=`[ -f %s ] && cat %s || true`", versionFile, versionFile),
					fmt.Sprintf("%s > %s", `printf "%s %s latest" "$GIT_REF" "$VERSION"`, tagListFile),
					fmt.Sprintf("%s $(cat %s)", `printf "Image will be tagged with: %s\n"`, tagListFile),
				}, "\n")},
			},
			Inputs: []atc.TaskInputConfig{
				{Name: gitDir},
			},
			Outputs: []atc.TaskOutputConfig{
				{Name: tagList_Dir},
			},
		},
	}
	if updatePipeline {
		createTagList.Config.Inputs = append(createTagList.Config.Inputs, atc.TaskInputConfig{Name: versionName})
	}
	return append([]atc.Step{}, stepWithAttemptsAndTimeout(createTagList, task.GetAttempts(), task.Timeout))
}

func trivyTask(task manifest.DockerPush, fullBasePath string) atc.StepConfig {
	var imageFile string
	var gitRef string
	if !multiPlatform(task) {
		imageFile = fmt.Sprintf("--input %s", path.Join(relativePathToRepoRoot(gitDir, fullBasePath), "image/image.tar"))
	} else {
		imageFile = shared.CachePath(task, "")
		gitRef = fmt.Sprintf(":$(cat %s)", path.Join(gitDir, ".git", "ref"))
		fullBasePath = ""
	}

	//exitCode := 1
	//if task.IgnoreVulnerabilities {
	//	exitCode = 0
	//}

	// temporary: always exit 0 until we have communicated the ignoreVulnerabilites opt-in
	exitCode := 0

	step := &atc.TaskStep{
		Name: "trivy",
		Config: &atc.TaskConfig{
			Platform: "linux",
			ImageResource: &atc.ImageResource{
				Type: "docker-image",
				Source: atc.Source{
					"repository": "aquasec/trivy",
				},
			},
			Run: atc.TaskRunConfig{
				Path: "/bin/sh",
				Args: []string{"-c", strings.Join([]string{
					`[ -f .trivyignore ] && echo "Ignoring the following CVE's due to .trivyignore" || true`,
					`[ -f .trivyignore ] && cat .trivyignore; echo || true`,
					fmt.Sprintf(`trivy image --timeout %dm --ignore-unfixed --severity CRITICAL --scanners vuln --exit-code %d %s%s || true`, task.ScanTimeout, exitCode, imageFile, gitRef),
				}, "\n")},
				Dir: fullBasePath,
			},
			Inputs: []atc.TaskInputConfig{
				{Name: gitDir},
				{Name: "image"},
			},
		},
	}

	if task.ReadsFromArtifacts() {
		step.Config.Inputs = append(step.Config.Inputs, atc.TaskInputConfig{Name: dockerBuildTmpDir})
	}

	return step
}

func buildAndPush(task manifest.DockerPush, resourceName string, fullBasePath string) []atc.Step {
	var steps []atc.Step
	image, tag := shared.SplitTag(task.Image)
	dockerImageWithCachePath := shared.CachePath(task, "")

	params := atc.TaskEnv{
		"CONTEXT":            path.Join(fullBasePath, task.BuildPath),
		"DOCKERFILE":         path.Join(fullBasePath, task.DockerfilePath),
		"DOCKER_CONFIG_JSON": "((halfpipe-gcr.docker_config))",
	}

	for k, v := range convertVars(task.Vars) {
		params[fmt.Sprintf("BUILD_ARG_%s", k)] = fmt.Sprintf("%s", v)
	}

	var buildStep *atc.TaskStep

	if !multiPlatform(task) {
		buildStep = &atc.TaskStep{
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
					Path: "/bin/sh",
					Args: []string{
						"-c",
						fmt.Sprintf("%s\n%s\n%s", "mkdir ~/.docker", "echo $DOCKER_CONFIG_JSON > ~/.docker/config.json", "build"),
					},
				},
				Inputs: []atc.TaskInputConfig{
					{Name: gitDir},
				},
				Outputs: []atc.TaskOutputConfig{
					{Name: "image"},
				},
			},
		}
	} else {
		gitRef := fmt.Sprintf("$(cat %s)", path.Join(gitDir, ".git", "ref"))

		buildStep = &atc.TaskStep{
			Name:       "build",
			Privileged: true,
			Config: &atc.TaskConfig{
				Platform: "linux",
				ImageResource: &atc.ImageResource{
					Type: "registry-image",
					Source: atc.Source{
						"repository": config.DockerRegistry + "halfpipe-buildx",
						"tag":        "latest",
						"password":   "((halfpipe-gcr.private_key))",
						"username":   "_json_key",
					},
				},
				Params: params,
				Run: atc.TaskRunConfig{
					Path: "/bin/sh",
					Args: []string{"-c", strings.Join([]string{
						`echo $DOCKER_CONFIG_JSON > ~/.docker/config.json`,
						fmt.Sprintf(`docker buildx build -f $DOCKERFILE --platform linux/amd64,linux/arm64 -t %s:%s --push --provenance=false $CONTEXT`, dockerImageWithCachePath, gitRef)}, "\n"),
					},
				},
				Inputs: []atc.TaskInputConfig{
					{Name: gitDir},
					{Name: tagList_Dir},
				},
				Outputs: []atc.TaskOutputConfig{
					{Name: "image"},
				},
			},
		}
	}

	if task.ReadsFromArtifacts() {
		buildStep.Config.Inputs = append(buildStep.Config.Inputs, atc.TaskInputConfig{Name: dockerBuildTmpDir})
	}

	steps = append(steps, stepWithAttemptsAndTimeout(buildStep, task.GetAttempts(), task.GetTimeout()))
	steps = append(steps, stepWithAttemptsAndTimeout(trivyTask(task, fullBasePath), task.GetAttempts(), task.GetTimeout()))

	if !multiPlatform(task) {
		putStep := &atc.PutStep{
			Name: resourceName,
			Params: atc.Params{
				"image":           "image/image.tar",
				"additional_tags": tagListFile,
			},
			NoGet: true,
		}
		steps = append(steps, stepWithAttemptsAndTimeout(putStep, task.GetAttempts(), task.GetTimeout()))
	} else {
		gitRef := fmt.Sprintf("$(cat %s)", path.Join(gitDir, ".git", "ref"))
		publishCommand := fmt.Sprintf(`for tag in $(cat %s) %s; do docker buildx imagetools create %s:%s --tag %s:$tag; done`, tagListFile, tag, dockerImageWithCachePath, gitRef, image)

		pushStep := &atc.TaskStep{
			Name:       "publish-final-image",
			Privileged: true,
			Config: &atc.TaskConfig{
				Platform: "linux",
				ImageResource: &atc.ImageResource{
					Type: "registry-image",
					Source: atc.Source{
						"repository": config.DockerRegistry + "halfpipe-buildx",
						"tag":        "latest",
						"password":   "((halfpipe-gcr.private_key))",
						"username":   "_json_key",
					},
				},
				Params: params,
				Run: atc.TaskRunConfig{
					Path: "/bin/sh",
					Args: []string{"-c", strings.Join([]string{
						`echo $DOCKER_CONFIG_JSON > ~/.docker/config.json`,
						publishCommand,
					}, "\n"),
					},
				},
				Inputs: []atc.TaskInputConfig{
					{Name: gitDir},
					{Name: tagList_Dir},
				},
				Outputs: []atc.TaskOutputConfig{
					{Name: "image"},
				},
			},
		}
		steps = append(steps, stepWithAttemptsAndTimeout(pushStep, task.GetAttempts(), task.GetTimeout()))
	}
	return steps
}

func multiPlatform(task manifest.DockerPush) bool {
	return !(len(task.Platforms) == 1 && task.Platforms[0] == "linux/amd64")
}
