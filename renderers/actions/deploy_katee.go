package actions

import (
	"fmt"
	"strconv"
	"strings"

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
			"CHECK_INTERVAL":         strconv.Itoa(task.CheckInterval),
			"KATEE_ENVIRONMENT":      task.Environment,
			"KATEE_NAMESPACE":        task.Namespace,
			"KATEE_PLATFORM_VERSION": task.PlatformVersion,
			"KATEE_APPFILE":          task.VelaManifest,
			"MAX_CHECKS":             strconv.Itoa(task.MaxChecks),
			"BUILD_VERSION":          "${{ env.BUILD_VERSION }}",
			"GIT_REVISION":           "${{ env.GIT_REVISION }}",
			"KATEE_GKE_CREDENTIALS":  fmt.Sprintf("((%s-service-account-prod.key))", task.Namespace),
			"KATEE_V2_GKE_CREDS":     fmt.Sprintf("((%s-service-account-prod.key))", strings.Replace(task.Namespace, "katee", "katee-v2", 1)),
		},
	}

	if task.Tag == "gitref" {
		deployKatee.Env["TAG"] = "${{ env.GIT_REVISION }}"
	} else if task.Tag == "version" {
		deployKatee.Env["TAG"] = "${{ env.BUILD_VERSION }}"
	}

	for k, v := range task.Vars {
		deployKatee.Env[k] = v
	}
	return append(steps, deployKatee)
}
