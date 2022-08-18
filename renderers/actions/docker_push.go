package actions

import (
	"fmt"
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
	if task.ShouldScanDockerImage() {
		steps = append(steps, scanImage(task))
	}
	steps = append(steps, pushImage(a, task, buildArgs))
	steps = append(steps, repositoryDispatch(task.Image))
	steps = append(steps, imageSummary(task.Image, tags(task)))
	return steps
}

func tags(task manifest.DockerPush) string {
	tag1 := fmt.Sprintf("%s:latest", task.Image)
	tag2 := fmt.Sprintf("%s:${{ env.BUILD_VERSION }}", task.Image)
	tag3 := fmt.Sprintf("%s:${{ env.GIT_REVISION }}", task.Image)

	return fmt.Sprintf("%s\n%s\n%s\n", tag1, tag2, tag3)
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
			{"push", false},
			{"tags", tags(task)},
			{"build-args", buildArgs.ToString()},
		},
	}
	return step
}

func scanImage(task manifest.DockerPush) Step {
	step := Step{
		Name: "Run Trivy vulnerability scanner",
		Uses: "aquasecurity/trivy-action@0.7.0",
		With: With{
			{"image-ref", task.Image + ":${{ env.GIT_REVISION }}"},
			{"format", "table"},
			{"exit-code", 1},
			{"ignore-unfixed", true},
			{"vuln-type", "os,library"},
			{"severity", task.SeverityList(",")},
		},
	}
	return step
}

func pushImage(a *Actions, task manifest.DockerPush, buildArgs Env) Step {
	step := Step{
		Name: "Push Image",
		Uses: "docker/build-push-action@v3",
		With: With{
			{"context", path.Join(a.workingDir, task.BuildPath)},
			{"file", path.Join(a.workingDir, task.DockerfilePath)},
			{"push", true},
			{"tags", tags(task)},
			{"build-args", buildArgs.ToString()},
		},
	}
	return step
}

func imageSummary(img string, tags string) Step {
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
	imgTags := strings.Split(tags, "\n")
	for _, tag := range imgTags[0 : len(imgTags)-1] {
		sRun = append(sRun, fmt.Sprintf(`echo "- %s" >> $GITHUB_STEP_SUMMARY`, tag))
	}

	return Step{
		Name: "Summary",
		Run:  strings.Join(sRun, "\n"),
	}
}
