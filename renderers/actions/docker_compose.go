package actions

import (
	"fmt"
	"sort"
	"strings"

	"github.com/springernature/halfpipe/manifest"
)

func (a *Actions) dockerComposeSteps(task manifest.DockerCompose) Steps {
	runTask := convertDockerComposeToRunTask(task)
	return append(a.runSteps(runTask), dockerCleanup(task))

}

func dockerCleanup(task manifest.DockerCompose) Step {
	return Step{
		Name: "Docker cleanup",
		If:   "always()",
		Run:  fmt.Sprintf("set -x; docker-compose -f %s down; docker ps -a; docker network ls; docker volume ls", task.ComposeFile),
	}
}

func convertDockerComposeToRunTask(task manifest.DockerCompose) manifest.Run {
	return manifest.Run{
		Retries:                task.Retries,
		Name:                   task.GetName(),
		Script:                 dockerComposeScript(task),
		Vars:                   task.Vars,
		SaveArtifacts:          task.SaveArtifacts,
		RestoreArtifacts:       task.RestoreArtifacts,
		SaveArtifactsOnFailure: task.SaveArtifactsOnFailure,
		Timeout:                task.GetTimeout(),
	}
}

func dockerComposeScript(task manifest.DockerCompose) string {
	allEnvVars := manifest.Vars{}
	for k := range task.Vars {
		allEnvVars[k] = ""
	}
	for k := range globalEnv {
		allEnvVars[k] = ""
	}

	envOptions := []string{}
	for key := range allEnvVars {
		envOptions = append(envOptions, fmt.Sprintf("-e %s", key))
	}
	sort.Strings(envOptions)

	command := []string{"docker-compose"}
	command = append(command, "-f "+task.ComposeFile)
	command = append(command, "run")
	command = append(command, envOptions...)
	command = append(command, task.Service)
	if task.Command != "" {
		command = append(command, task.Command)
	}

	login := `docker login -u _json_key -p "$GCR_PRIVATE_KEY" https://eu.gcr.io`
	compose := strings.Join(command, " \\\n  ")
	return fmt.Sprintf("%s\n%s\n", login, compose)
}
