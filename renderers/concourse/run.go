package concourse

import (
	"fmt"
	"path"
	"strings"

	"github.com/concourse/concourse/atc"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
)

func (c Concourse) runJob(task manifest.Run, man manifest.Manifest, isDockerCompose bool, basePath string) atc.JobConfig {
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

	taskOutputs := func() []atc.TaskOutputConfig {
		var outputs []atc.TaskOutputConfig
		if len(task.SaveArtifacts) > 0 {
			outputs = append(outputs, atc.TaskOutputConfig{Name: artifactsOutDir})
		}

		if len(task.SaveArtifactsOnFailure) > 0 {
			outputs = append(outputs, atc.TaskOutputConfig{Name: artifactsOutDirOnFailure})
		}
		return outputs
	}

	jobConfig := atc.JobConfig{
		Name:   task.GetName(),
		Serial: true,
	}

	taskPath := "/bin/sh"
	if isDockerCompose {
		taskPath = "docker.sh"
	}

	taskEnv := make(atc.TaskEnv)
	for key, value := range task.Vars {
		taskEnv[key] = value
	}

	var caches []atc.TaskCacheConfig
	for _, dir := range config.CacheDirs {
		caches = append(caches, atc.TaskCacheConfig{Path: dir})
	}

	runStep := &atc.TaskStep{
		Name:       restrictAllowedCharacterSet(task.GetName()),
		Privileged: task.Privileged,
		Config: &atc.TaskConfig{
			Platform:      "linux",
			Params:        taskEnv,
			ImageResource: c.imageResource(task.Docker),
			Run: atc.TaskRunConfig{
				Path: taskPath,
				Dir:  path.Join(gitDir, basePath),
				Args: runScriptArgs(task, man, !isDockerCompose, basePath),
			},
			Inputs:  taskInputs(),
			Outputs: taskOutputs(),
			Caches:  caches,
		},
	}

	step := stepWithAttemptsAndTimeout(runStep, task.GetAttempts(), task.GetTimeout())

	jobConfig.PlanSequence = append(jobConfig.PlanSequence, step)

	if len(task.SaveArtifacts) > 0 {
		artifactPut := &atc.PutStep{
			Name: artifactsName,
			Params: atc.Params{
				"folder":       artifactsOutDir,
				"version_file": path.Join(gitDir, ".git", "ref"),
			},
			NoGet: true,
		}

		jobConfig.PlanSequence = append(jobConfig.PlanSequence, stepWithAttemptsAndTimeout(artifactPut, defaultStepAttempts, defaultStepTimeout))
	}

	return jobConfig
}

var warningMissingBash = `if ! which bash > /dev/null && [ "$SUPPRESS_BASH_WARNING" != "true" ]; then
  echo "WARNING: Bash is not present in the docker image"
  echo "If your script depends on bash you will get a strange error message like:"
  echo "  sh: yourscript.sh: command not found"
  echo "To fix, make sure your docker image contains bash!"
  echo "Or if you are sure you don't need bash you can suppress this warning by setting the environment variable \"SUPPRESS_BASH_WARNING\" to \"true\"."
  echo ""
  echo ""
fi
`

var warningAlpineImage = `if [ -e /etc/alpine-release ]
then
  echo "WARNING: you are running your build in a Alpine image or one that is based on the Alpine"
  echo "There is a known issue where DNS resolving does not work as expected"
  echo "https://github.com/gliderlabs/docker-alpine/issues/255"
  echo "If you see any errors related to resolving hostnames the best course of action is to switch to another image"
  echo "we recommend debian:buster-slim as an alternative"
  echo ""
  echo ""
fi
`

func runScriptArgs(task manifest.Run, man manifest.Manifest, enableWarningMessages bool, basePath string) []string {

	script := task.Script
	if !strings.HasPrefix(script, "./") && !strings.HasPrefix(script, "/") && !strings.HasPrefix(script, `\`) {
		script = "./" + script
	}

	var out []string

	if enableWarningMessages {
		out = append(out, warningMissingBash, warningAlpineImage)
	}

	if len(task.SaveArtifacts) != 0 || len(task.SaveArtifactsOnFailure) != 0 {
		out = append(out, `copyArtifact() {
  ARTIFACT=$1
  ARTIFACT_OUT_PATH=$2

  if [ -e $ARTIFACT ] ; then
    mkdir -p $ARTIFACT_OUT_PATH
    cp -r $ARTIFACT $ARTIFACT_OUT_PATH
  else
    echo "ERROR: Artifact '$ARTIFACT' not found. Try fly hijack to check the filesystem."
    exit 1
  fi
}
`)
	}

	if task.RestoreArtifacts {
		out = append(out, "# Copying in artifacts from previous task")
		out = append(out, fmt.Sprintf("cp -r %s/. %s\n", pathToArtifactsDir(gitDir, basePath, artifactsInDir), relativePathToRepoRoot(gitDir, basePath)))
	}

	out = append(out,
		fmt.Sprintf("export GIT_REVISION=`cat %s`", pathToGitRef(gitDir, basePath)),
	)

	if man.FeatureToggles.UpdatePipeline() {
		out = append(out,
			fmt.Sprintf("export BUILD_VERSION=`cat %s`", pathToVersionFile(gitDir, basePath)),
		)
	}

	scriptCall := fmt.Sprintf(`
%s
EXIT_STATUS=$?
if [ $EXIT_STATUS != 0 ] ; then
%s
fi
`, script, onErrorScript(task.SaveArtifactsOnFailure, basePath))
	out = append(out, scriptCall)

	if len(task.SaveArtifacts) != 0 {
		out = append(out, "# Artifacts to copy from task")
	}
	for _, artifactPath := range task.SaveArtifacts {
		out = append(out, fmt.Sprintf("copyArtifact %s %s", artifactPath, fullPathToArtifactsDir(gitDir, basePath, artifactsOutDir, artifactPath)))
	}
	return []string{"-c", strings.Join(out, "\n")}
}
