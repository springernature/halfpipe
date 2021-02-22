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

	envOptions := []string{}
	for key := range allEnvVars {
		envOptions = append(envOptions, fmt.Sprintf("-e %s", key))
	}
	sort.Strings(envOptions)

	mounts := []string{
		fmt.Sprintf("-v /mnt/halfpipe-cache/%s:/var/halfpipe/shared-cache", team),
	}

	dcPrefix := fmt.Sprintf("docker-compose -f %s ", task.ComposeFile)
	dcRun := []string{dcPrefix + "run"}
	dcRun = append(dcRun, envOptions...)
	dcRun = append(dcRun, mounts...)
	dcRun = append(dcRun, task.Service)
	if task.Command != "" {
		dcRun = append(dcRun, task.Command)
	}
	run := strings.Join(dcRun, " \\\n  ")
	pull := dcPrefix + "pull"
	return fmt.Sprintf("%s\n%s", pull, run)
}
