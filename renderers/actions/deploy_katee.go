package actions

import (
	"fmt"
	"strconv"

	"github.com/springernature/halfpipe/manifest"
)

func (a *Actions) deployKateeSteps(task manifest.DeployKatee) (steps Steps) {
	deployKatee := Step{
		Name: "Deploy to Katee",
		Uses: "docker://eu.gcr.io/halfpipe-io/ee-katee-vela-cli:latest",
		With: With{
			"entrypoint": "/bin/sh",
			"args":       fmt.Sprintf(`-c "cd %s; halfpipe-deploy`, a.workingDir)},
		Env: Env{
			"KATEE_ENVIRONMENT":      task.Environment,
			"KATEE_NAMESPACE":        task.Namespace,
			"KATEE_PLATFORM_VERSION": task.PlatformVersion,
			"KATEE_APPFILE":          task.VelaManifest,
			"BUILD_VERSION":          "${{ env.BUILD_VERSION }}",
			"GIT_REVISION":           "${{ env.GIT_REVISION }}",
			"KATEE_GKE_CREDENTIALS":  fmt.Sprintf("((%s-service-account-prod.key))", task.Namespace),
		},
	}

	if task.Tag == "gitref" {
		deployKatee.Env["TAG"] = "${{ env.GIT_REVISION }}"
	} else if task.Tag == "version" {
		deployKatee.Env["TAG"] = "${{ env.BUILD_VERSION }}"
	}

	if task.DeploymentCheckTimeout != 0 {
		deployKatee.Env["MAX_CHECKS"] = strconv.Itoa(task.DeploymentCheckTimeout)
	}

	for k, v := range task.Vars {
		deployKatee.Env[k] = v
	}
	return append(steps, deployKatee)
}
