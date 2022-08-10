package actions

import (
	"fmt"
	"github.com/springernature/halfpipe/manifest"
)

func (a *Actions) deployKateeSteps(task manifest.DeployKatee, man manifest.Manifest) (steps Steps) {
	deployKatee := a.createKateeDeployStep(task, man)
	deploymentStatus := a.createDeploymentStatus(task, man)
	return append(steps, deployKatee, deploymentStatus)
}

func (a *Actions) createKateeDeployStep(task manifest.DeployKatee, man manifest.Manifest) Step {
	deployKatee := Step{
		Name: "Deploy to Katee",
		Uses: "docker://eu.gcr.io/halfpipe-io/ee-katee-vela-cli:latest",
		With: With{
			{"entrypoint", "/bin/sh"},
			{"args", fmt.Sprintf(`-c "cd %s; /exe vela up -f $KATEE_APPFILE --publish-version $DOCKER_TAG`, a.workingDir)}},
		Env: Env{
			"KATEE_TEAM":             man.Team,
			"KATEE_APPFILE":          task.VelaManifest,
			"KATEE_APPLICATION_NAME": task.ApplicationName,
			"BUILD_VERSION":          "${{ env.BUILD_VERSION }}",
			"GIT_REVISION":           "${{ env.GIT_REVISION }}",
			"KATEE_GKE_CREDENTIALS":  fmt.Sprintf("((katee-%s-service-account-prod.key))", man.Team),
		},
	}

	if task.Tag == "gitref" {
		deployKatee.Env["DOCKER_TAG"] = "${{ env.GIT_REVISION }}"
		deployKatee.Env["KATEE_APPLICATION_IMAGE"] = fmt.Sprintf("%s:%s", task.Image, "${{ env.GIT_REVISION }}")
	} else if task.Tag == "version" {
		deployKatee.Env["DOCKER_TAG"] = "${{ env.BUILD_VERSION }}"
		deployKatee.Env["KATEE_APPLICATION_IMAGE"] = fmt.Sprintf("%s:%s", task.Image, "${{ env.BUILD_VERSION }}")
	}

	for k, v := range task.Vars {
		deployKatee.Env[k] = v
	}

	return deployKatee
}

func (a Actions) createDeploymentStatus(task manifest.DeployKatee, man manifest.Manifest) Step {
	createDeploymentStatus := Step{
		Name: "Check Deployment Status",
		Uses: "docker://eu.gcr.io/halfpipe-io/ee-katee-vela-cli:latest",
		With: With{
			{"entrypoint", "/bin/sh"},
			{"args", fmt.Sprintf(`-c "cd %s; /exe deployment-status katee-%s %s $PUBLISHED_VERSION`, a.workingDir, man.Team, task.ApplicationName)}},
		Env: Env{
			"KATEE_GKE_CREDENTIALS": fmt.Sprintf("((katee-%s-service-account-prod.key))", man.Team),
			"KATEE_TEAM":            man.Team,
		},
	}
	if task.Tag == "gitref" {
		createDeploymentStatus.Env["PUBLISHED_VERSION"] = "${{ env.GIT_REVISION }}"
	} else if task.Tag == "version" {
		createDeploymentStatus.Env["PUBLISHED_VERSION"] = "${{ env.BUILD_VERSION }}"
	}

	return createDeploymentStatus
}
