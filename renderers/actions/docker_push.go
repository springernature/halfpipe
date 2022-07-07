package actions

import (
	"fmt"
	"path"
	"strings"

	"github.com/springernature/halfpipe/manifest"
)

func (a *Actions) dockerPushSteps(task manifest.DockerPush) (steps Steps) {
	steps = dockerLogin(task.Image, task.Username, task.Password)
	imgTags := tags(task)

	steps = append(steps, Step{
		Name: "Build and push",
		Uses: "docker/build-push-action@v3",
		With: With{
			{"context", path.Join(a.workingDir, task.BuildPath)},
			{"file", path.Join(a.workingDir, task.DockerfilePath)},
			{"push", true},
			{"tags", imgTags},
		},
		Env: Env(task.Vars),
	})

	steps = append(steps, repositoryDispatch(task.Image))
	steps = append(steps, imageSummary(task.Image, imgTags))
	return steps
}

func tags(task manifest.DockerPush) string {
	tagVar := "${{ env.BUILD_VERSION }}"
	if task.Tag == "gitref" {
		tagVar = "${{ env.GIT_REVISION }}"
	}
	return fmt.Sprintf("%s:latest\n%s:%s\n", task.Image, task.Image, tagVar)
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
