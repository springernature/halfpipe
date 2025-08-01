package concourse

import (
	"fmt"
	"github.com/concourse/concourse/atc"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
	"maps"
	"path"
	"slices"
	"strings"
)

func (c Concourse) PackJob(task manifest.Buildpack, basePath string, man manifest.Manifest) atc.JobConfig {
	taskInputs := func() []atc.TaskInputConfig {
		inputs := []atc.TaskInputConfig{{Name: manifest.GitTrigger{}.GetTriggerName()}}
		if task.RestoreArtifacts {
			inputs = append(inputs, atc.TaskInputConfig{Name: artifactsName})
		}

		if man.FeatureToggles.UpdatePipeline() {
			inputs = append(inputs, atc.TaskInputConfig{Name: versionName})
		}
		return inputs
	}

	jobConfig := atc.JobConfig{
		Name:   task.GetName(),
		Serial: true,
	}

	taskEnv := make(atc.TaskEnv)
	for key, value := range task.Vars {
		taskEnv[key] = value
	}

	taskEnv["DOCKER_CONFIG_JSON"] = "((halfpipe-gcr.docker_config))"

	var caches []atc.TaskCacheConfig
	for _, dir := range config.CacheDirs {
		caches = append(caches, atc.TaskCacheConfig{Path: dir})
	}

	packStep := &atc.TaskStep{
		Name:       restrictAllowedCharacterSet(task.GetName()),
		Privileged: true,
		Config: &atc.TaskConfig{
			Platform: "linux",
			Params:   taskEnv,
			ImageResource: c.imageResource(manifest.Docker{
				Image:    config.DockerRegistry + "engineering-enablement/halfpipe-buildx-pack",
				Username: "_json_key",
				Password: "((halfpipe-gcr.private_key))",
			}),
			Run: atc.TaskRunConfig{
				Path: "docker.sh",
				Dir:  path.Join(gitDir, basePath),
				Args: packScriptArgs(task, man, basePath),
			},
			Inputs: taskInputs(),
			Caches: caches,
		},
	}

	step := stepWithAttemptsAndTimeout(packStep, task.GetAttempts(), task.GetTimeout())

	jobConfig.PlanSequence = append(jobConfig.PlanSequence, step)

	return jobConfig
}

func packScriptArgs(task manifest.Buildpack, man manifest.Manifest, basePath string) []string {
	var out []string
	var versionTag string

	if task.RestoreArtifacts {
		out = append(out, `# Copying in artifacts from previous task`)
		out = append(out, fmt.Sprintf("cp -r %s/. %s\n", pathToArtifactsDir(gitDir, basePath, artifactsInDir), relativePathToRepoRoot(gitDir, basePath)))
	}

	out = append(out,
		fmt.Sprintf("export GIT_REVISION=`cat %s`", pathToGitRef(gitDir, basePath)),
	)

	if man.FeatureToggles.UpdatePipeline() {
		out = append(out,
			fmt.Sprintf("export BUILD_VERSION=`cat %s`", pathToVersionFile(gitDir, basePath)),
		)

		versionTag = fmt.Sprintf("--tag %s:${BUILD_VERSION} ", task.Image)
	}

	appPath := basePath
	if len(task.Path) > 0 {
		appPath = path.Join(appPath, task.Path)
	}

	out = append(out, `echo $DOCKER_CONFIG_JSON > ~/.docker/config.json`)

	envVars := ""
	for _, key := range slices.Sorted(maps.Keys(task.Vars)) {
		envVars += fmt.Sprintf(`--env "%s=%s" \
`, key, task.Vars[key])
	}

	command := fmt.Sprintf(`pack build %s \
--path %s \
--builder paketobuildpacks/builder-jammy-buildpackless-full \
--buildpack %s \
--tag %s:${GIT_REVISION} %s \
%s--publish \
--trust-builder
`, task.Image, appPath, task.Buildpacks, task.Image, versionTag, envVars)

	out = append(out, "echo "+command, command)

	return []string{"-c", strings.Join(out, "\n")}
}
