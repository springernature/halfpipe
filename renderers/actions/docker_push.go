package actions

import (
	"fmt"
	"github.com/springernature/halfpipe/config"
	"path"
	"strings"

	"github.com/springernature/halfpipe/manifest"
)

func (a *Actions) dockerPushSteps(task manifest.DockerPush) (steps Steps) {
	steps = dockerLogin(task.Image, task.Username, task.Password)
	buildArgs := Env{}
	for k, v := range globalEnv {
		buildArgs[k] = v
	}
	for k, v := range task.Vars {
		buildArgs[k] = v
	}

	steps = append(steps, buildImage(a, task, buildArgs))
	steps = append(steps, scanImage(a, task))
	steps = append(steps, pushImage(task))
	steps = append(steps, repositoryDispatch(task.Image))
	steps = append(steps, jobSummary(task.Image, tags(task)))
	return steps
}

func tags(task manifest.DockerPush) []string {
	tag1 := fmt.Sprintf("%s:latest", task.Image)
	tag2 := fmt.Sprintf("%s:${{ env.BUILD_VERSION }}", task.Image)
	tag3 := fmt.Sprintf("%s:${{ env.GIT_REVISION }}", task.Image)

	return []string{tag1, tag2, tag3}
}

func tagWithCachePath(task manifest.DockerPush) string {
	tag := ":${{ env.GIT_REVISION }}"
	if strings.HasPrefix(task.Image, config.DockerRegistry) {
		r := strings.Replace(task.Image, config.DockerRegistry, fmt.Sprintf("%scache/", config.DockerRegistry), 1)
		return r + tag
	} else {
		return config.DockerRegistry + "cache/" + task.Image + tag
	}
}

func repositoryDispatch(eventName string) Step {
	return Step{
		Name: "Repository dispatch",
		Uses: "peter-evans/repository-dispatch@v2",
		With: With{
			{"token", githubSecrets.RepositoryDispatchToken},
			{"event-type", "docker-push:" + eventName},
		},
	}
}

func buildImage(a *Actions, task manifest.DockerPush, buildArgs Env) Step {
	step := Step{
		Name: "Build Image",
		Uses: "docker/build-push-action@v3",
		With: With{
			{"context", path.Join(a.workingDir, task.BuildPath)},
			{"file", path.Join(a.workingDir, task.DockerfilePath)},
			{"push", true},
			{"tags", tagWithCachePath(task)},
			{"build-args", buildArgs.ToString()},
			{"platforms", strings.Join(task.Platforms, ",")},
		},
	}
	return step
}

func scanImage(a *Actions, task manifest.DockerPush) Step {
	exitCode := 1
	if task.IgnoreVulnerabilities {
		exitCode = 0
	}
	prefix := ""
	if a.workingDir != "" {
		prefix = fmt.Sprintf("cd %s; ", a.workingDir)
	}

	step := Step{
		Name: "Run Trivy vulnerability scanner",
		Uses: "docker://aquasec/trivy",
		With: With{
			{"entrypoint", "/bin/sh"},
			{"args", fmt.Sprintf(`-c "%s [ -f .trivyignore ] && echo \"Ignoring the following CVE's due to .trivyignore\" || true; [ -f .trivyignore ] && cat .trivyignore; echo || true; trivy image --timeout 30m --ignore-unfixed --severity CRITICAL --exit-code %s %s"`, prefix, fmt.Sprint(exitCode), tagWithCachePath(task))},
		},
	}
	return step
}

func pushImage(task manifest.DockerPush) Step {
	sRun := []string{}
	for _, tag := range tags(task) {
		sRun = append(sRun, fmt.Sprintf("docker buildx imagetools create %s --tag %s", tagWithCachePath(task), tag))
	}

	step := Step{
		Name: "Push Image",
		Run:  strings.Join(sRun, "\n"),
	}

	return step
}

func jobSummary(img string, tags []string) Step {
	sRun := []string{}
	sRun = append(sRun, `echo ":ship: **Image Pushed Successfully**" >> $GITHUB_STEP_SUMMARY`)
	sRun = append(sRun, `echo "" >> $GITHUB_STEP_SUMMARY`)

	// Taken from dockerLogin(task.Image, task.Username, task.Password)
	// set registry if not docker hub by counting slashes
	// docker hub format: repository:tag or user/repository:tag
	// other registries:  another.registry/user/repository:tag
	if strings.Count(img, "/") > 1 {
		registry := fmt.Sprintf("https://%s", img)
		sRun = append(sRun, fmt.Sprintf(`echo "[%s](%s)" >> $GITHUB_STEP_SUMMARY`, img, registry))
	} else {
		registry := fmt.Sprintf("https://hub.docker.com/r/%s", img)
		sRun = append(sRun, fmt.Sprintf(`echo "[%s](%s)" >> $GITHUB_STEP_SUMMARY`, img, registry))
	}

	sRun = append(sRun, `echo "" >> $GITHUB_STEP_SUMMARY`)
	sRun = append(sRun, `echo "Tags:" >> $GITHUB_STEP_SUMMARY`)
	// Split the tag lines, remove the last empty line and add the tags to summary
	for _, tag := range tags {
		sRun = append(sRun, fmt.Sprintf(`echo "- %s" >> $GITHUB_STEP_SUMMARY`, tag))
	}

	return Step{
		Name: "Summary",
		Run:  strings.Join(sRun, "\n"),
	}
}
