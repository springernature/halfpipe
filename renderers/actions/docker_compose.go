package actions

import (
	"fmt"
	"sort"
	"strings"

	"github.com/springernature/halfpipe/manifest"
)

func (a *Actions) dockerComposeSteps(task manifest.DockerCompose, team string) Steps {
	runTask := convertDockerComposeToRunTask(task, team)
	return append(a.runSteps(runTask), dockerCleanup(task))

}

func dockerCleanup(task manifest.DockerCompose) Step {
	return Step{
		Name: "Docker cleanup",
		If:   "always()",
		Run:  fmt.Sprintf("docker-compose -f %s down", task.ComposeFile),
	}
}

func convertDockerComposeToRunTask(task manifest.DockerCompose, team string) manifest.Run {
	return manifest.Run{
		Retries:                task.Retries,
		Name:                   task.GetName(),
		Script:                 dockerComposeScript(task, team),
		Vars:                   task.Vars,
		SaveArtifacts:          task.SaveArtifacts,
		RestoreArtifacts:       task.RestoreArtifacts,
		SaveArtifactsOnFailure: task.SaveArtifactsOnFailure,
		Timeout:                task.GetTimeout(),
	}
}

func dockerComposeScript(task manifest.DockerCompose, team string) string {
	allEnvVars := manifest.Vars{}
	for k := range task.Vars {
		allEnvVars[k] = ""
	}
	for k := range globalEnv {
		allEnvVars[k] = ""
	}

	options := []string{
		"-v /var/run/docker.sock:/var/run/docker.sock",
		fmt.Sprintf("-v /mnt/halfpipe-cache/%s:/var/halfpipe/shared-cache", team),
	}
	for key := range allEnvVars {
		options = append(options, fmt.Sprintf("-e %s", key))
	}
	sort.Strings(options)

	dcRun := []string{fmt.Sprintf("docker-compose -f %s run", task.ComposeFile)}
	dcRun = append(dcRun, "--use-aliases")
	dcRun = append(dcRun, options...)
	dcRun = append(dcRun, task.Service)
	if task.Command != "" {
		dcRun = append(dcRun, task.Command)
	}
	return strings.Join(dcRun, " \\\n  ")
}
