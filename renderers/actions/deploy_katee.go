package actions

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/springernature/halfpipe/manifest"
)

func (a *Actions) deployKateeSteps(task manifest.DeployKatee) (steps Steps) {

	revision := "${{ env.BUILD_VERSION }}"
	if task.Tag == "gitref" {
		revision = "${{ env.GIT_REVISION }}"
	}

	deployKatee := Step{
		Name: "Deploy to Katee",
		Uses: "springernature/ee-action-deploy-katee@v1",
		With: With{
			"credentials":   fmt.Sprintf("((%s-service-account-prod.key))", strings.Replace(task.Namespace, "katee", "katee-v2", 1)),
			"namespace":     task.Namespace,
			"revision":      revision,
			"velaFile":      task.VelaManifest,
			"maxChecks":     strconv.Itoa(task.MaxChecks),
			"checkInterval": strconv.Itoa(task.CheckInterval),
		},
		Env: Env{
			"BUILD_VERSION": "${{ env.BUILD_VERSION }}",
			"GIT_REVISION":  "${{ env.GIT_REVISION }}",
		},
	}

	for k, v := range task.Vars {
		deployKatee.Env[k] = v
	}
	return append(steps, deployKatee)
}
