package concourse

import (
	"fmt"
	"github.com/concourse/concourse/atc"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/renderers/shared"
	"golang.org/x/exp/slices"
	"path"
	"strings"
)

const tagList_Dir = "tagList"

var tagListFile = path.Join(tagList_Dir, "tagList")

func (c Concourse) dockerPushJob(task manifest.DockerPush, basePath string, man manifest.Manifest) atc.JobConfig {
	var steps []atc.Step

	steps = append(steps, restoreArtifacts(task)...)
	steps = append(steps, createTagList(task, man.FeatureToggles.UpdatePipeline())...)
	steps = append(steps, buildAndPush(task, basePath)...)

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

func trivyTask(task manifest.DockerPush, fullBasePath string, basePath string) atc.StepConfig {
	imageFile := shared.CachePath(task, "")
	gitRef := fmt.Sprintf(":$(cat %s)", pathToGitRef(gitDir, basePath))

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
			Params: atc.TaskEnv{
				"DOCKER_CONFIG_JSON": "((halfpipe-gcr.docker_config))",
			},
			Inputs: []atc.TaskInputConfig{
				{Name: gitDir},
			},
		},
	}

	if task.ReadsFromArtifacts() {
		step.Config.Inputs = append(step.Config.Inputs, atc.TaskInputConfig{Name: dockerBuildTmpDir})
	}

	return step
}

func buildAndPush(task manifest.DockerPush, basePath string) []atc.Step {
	var steps []atc.Step
	image, tag := shared.SplitTag(task.Image)
	dockerImageWithCachePath := shared.CachePath(task, "")

	fullBasePath := path.Join(gitDir, basePath)
	if task.RestoreArtifacts {
		fullBasePath = path.Join(dockerBuildTmpDir, basePath)
	}

	params := atc.TaskEnv{
		"DOCKER_CONFIG_JSON": "((halfpipe-gcr.docker_config))",
	}

	var buildStep *atc.TaskStep

	buildCommand := []string{
		"docker buildx build",
		fmt.Sprintf("-f %s", path.Join(fullBasePath, task.DockerfilePath)),
		"--push",
		"--provenance false",
		fmt.Sprintf("--platform %s", strings.Join(task.Platforms, ",")),
		fmt.Sprintf("--tag %s", shared.CachePath(task, "$(cat git/.git/ref)")),
	}

	buildArgs := []string{}
	for k, v := range convertVars(task.Vars) {
		params[k] = v.(string)
		buildArgs = append(buildArgs, fmt.Sprintf("--build-arg %s", k))
	}
	slices.Sort(buildArgs)
	buildCommand = append(buildCommand, buildArgs...)

	secrets := []string{}
	for k, v := range convertVars(task.Secrets) {
		params[k] = v.(string)
		secrets = append(secrets, fmt.Sprintf("--secret id=%s", k))

	}
	slices.Sort(secrets)
	buildCommand = append(buildCommand, secrets...)

	if task.UseCache {
		buildCommand = append(buildCommand, fmt.Sprintf("--tag %s", shared.CachePath(task, "buildcache")))
		buildCommand = append(buildCommand, fmt.Sprintf("--cache-from type=registry,ref=%s", shared.CachePath(task, "buildcache")))
		buildCommand = append(buildCommand, "--cache-to type=inline")
	}

	buildCommand = append(buildCommand, path.Join(fullBasePath, task.BuildPath))

	buildCommandStr := strings.Join(buildCommand, ` \
  `)
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
					fmt.Sprintf(`echo $ %s`, buildCommandStr),
					buildCommandStr}, "\n"),
				},
			},
			Inputs: []atc.TaskInputConfig{
				{Name: gitDir},
				{Name: tagList_Dir},
			},
		},
	}

	if task.ReadsFromArtifacts() {
		buildStep.Config.Inputs = append(buildStep.Config.Inputs, atc.TaskInputConfig{Name: dockerBuildTmpDir})
	}

	steps = append(steps, stepWithAttemptsAndTimeout(buildStep, task.GetAttempts(), task.GetTimeout()))
	steps = append(steps, stepWithAttemptsAndTimeout(trivyTask(task, fullBasePath, basePath), task.GetAttempts(), task.GetTimeout()))

	publishCommand := fmt.Sprintf(`for tag in $(cat %s) %s; do docker buildx imagetools create %s:$(cat git/.git/ref) --tag %s:$tag; done`, tagListFile, tag, dockerImageWithCachePath, image)

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
			Params: atc.TaskEnv{
				"DOCKER_CONFIG_JSON": "((halfpipe-gcr.docker_config))",
			},
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
		},
	}
	steps = append(steps, stepWithAttemptsAndTimeout(pushStep, task.GetAttempts(), task.GetTimeout()))

	return steps
}
