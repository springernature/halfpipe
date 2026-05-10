package actions

import (
	"maps"
	"path"
	"strconv"

	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
)

func (a *Actions) deployKateeSteps(task manifest.DeployKatee, man manifest.Manifest) (steps Steps) {

	revision := "2.${{ github.run_number }}.${{ github.run_attempt }}"
	if task.Tag == "gitref" {
		revision = "${{ env.GIT_REVISION }}"
	} else if task.Tag == "version" {
		revision = "${{ env.BUILD_VERSION }}"
	}

	deployKatee := Step{
		Name: "Deploy to Katee",
		Uses: ExternalActions.DeployKatee.Ref,
		With: With{
			"credentials":   config.VaultSecrets.KateeKey(task.Namespace),
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

	maps.Copy(deployKatee.Env, task.Vars)
	if man.OpsLevel.System != "" {
		deployKatee.With["eaid"] = man.OpsLevel.System
	}

	return append(steps, deployKatee)
}
