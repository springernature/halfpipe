package actions

import (
	"fmt"
	"path"
	"strconv"
	"strings"

	"github.com/springernature/halfpipe/manifest"
)

func (a *Actions) deployKateeSteps(task manifest.DeployKatee) (steps Steps) {

	revision := "2.${{ github.run_number }}.${{ github.run_attempt }}"
	if task.Tag == "gitref" {
		revision = "${{ env.GIT_REVISION }}"
	} else if task.Tag == "version" {
		revision = "${{ env.BUILD_VERSION }}"
	}

	deployKatee := Step{
		Name: "Deploy to Katee",
		Uses: ExternalActions.DeployKatee,
		With: With{
			"credentials":   fmt.Sprintf("((%s-service-account-prod.key))", strings.Replace(task.Namespace, "katee", "katee-v2", 1)),
			"namespace":     task.Namespace,
			"revision":      revision,
			"velaFile":      path.Join(a.workingDir, task.VelaManifest),
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
