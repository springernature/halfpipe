package actions

import (
	"fmt"
	"github.com/springernature/halfpipe/renderers/shared"
	"path"
	"strings"

	"github.com/springernature/halfpipe/manifest"
)

func (a *Actions) dockerPushSteps(task manifest.DockerPush) (steps Steps) {
	steps = dockerLogin(task.Image, task.Username, task.Password)
	steps = append(steps, buildImage(a, task))
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

func repositoryDispatch(eventName string) Step {
	return Step{
		Name: "Repository dispatch",
		Uses: "peter-evans/repository-dispatch@v2",
		With: With{
			"token":      githubSecrets.RepositoryDispatchToken,
			"event-type": "docker-push:" + eventName,
		},
	}
}

func buildImage(a *Actions, task manifest.DockerPush) Step {
	buildArgs := map[string]string{
		"ARTIFACTORY_PASSWORD": "",
		"ARTIFACTORY_URL":      "",
		"ARTIFACTORY_USERNAME": "",
		"BUILD_VERSION":        "",
		"GIT_REVISION":         "",
		"RUNNING_IN_CI":        "",
	}
	for k, v := range task.Vars {
		buildArgs[k] = v
	}

	step := Step{
		Name: "Build Image",
		Uses: "docker/build-push-action@v4",
		With: With{
			"context":    path.Join(a.workingDir, task.BuildPath),
			"file":       path.Join(a.workingDir, task.DockerfilePath),
			"push":       true,
			"tags":       shared.CachePath(task, "${{ env.GIT_REVISION }}"),
			"build-args": MultiLine{buildArgs},
			"platforms":  strings.Join(task.Platforms, ","),
			"provenance": false,
			"secrets":    MultiLine{task.Secrets},
		},
	}

	if task.UseCache {
		step.With["tags"] = fmt.Sprintf("%s\n%s", step.With["tags"], shared.CachePath(task, "buildcache"))
		step.With["cache-from"] = fmt.Sprintf("type=registry,ref=%s", shared.CachePath(task, "buildcache"))
		step.With["cache-to"] = "type=inline"
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
			"entrypoint": "/bin/sh",
			"args":       fmt.Sprintf(`-c "%s [ -f .trivyignore ] && echo \"Ignoring the following CVE's due to .trivyignore\" || true; [ -f .trivyignore ] && cat .trivyignore; echo || true; trivy image --timeout 30m --ignore-unfixed --severity CRITICAL --scanners vuln --exit-code %s %s"`, prefix, fmt.Sprint(exitCode), shared.CachePath(task, ":${{ env.GIT_REVISION }}")),
		},
	}
	return step
}

func pushImage(task manifest.DockerPush) Step {
	var sRun []string
	for _, tag := range tags(task) {
		sRun = append(sRun, fmt.Sprintf("docker buildx imagetools create %s --tag %s", shared.CachePath(task, ":${{ env.GIT_REVISION }}"), tag))
	}

	step := Step{
		Name: "Push Image",
		Run:  strings.Join(sRun, "\n"),
	}

	return step
}

func jobSummary(img string, tags []string) Step {
	var sRun []string
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
