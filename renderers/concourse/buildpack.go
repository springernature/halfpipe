package concourse

import (
	"fmt"
	"maps"
	"path"
	"slices"
	"strings"

	"github.com/concourse/concourse/atc"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
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
	maps.Copy(taskEnv, task.Vars)

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

	appPath := "."
	if len(task.Path) > 0 {
		appPath = task.Path
	}

	out = append(out, `echo $DOCKER_CONFIG_JSON > ~/.docker/config.json`)

	var envVars strings.Builder
	for _, key := range slices.Sorted(maps.Keys(task.Vars)) {
		envVars.WriteString(fmt.Sprintf(`--env "%s=%s" \
`, key, task.Vars[key]))
	}

	buildpacksFlags := ""
	for _, bp := range task.Buildpacks {
		buildpacksFlags += fmt.Sprintf("--buildpack %s \\\n", bp)
	}
	if buildpacksFlags == "" { // ensure at least one flag placeholder to avoid malformed command
		buildpacksFlags = ""
	}
	command := fmt.Sprintf(`pack build %s \
--path %s \
--builder %s \
%s--tag %s:${GIT_REVISION} %s \
%s--publish \
--trust-builder
`, task.Image, appPath, task.Builder, buildpacksFlags, task.Image, versionTag, envVars.String())

	out = append(out, "echo "+command, command)

	return []string{"-c", strings.Join(out, "\n")}
}
