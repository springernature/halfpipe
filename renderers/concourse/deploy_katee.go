package concourse

import (
	"fmt"
	"github.com/concourse/concourse/atc"
	"github.com/springernature/halfpipe/manifest"
)

func (c Concourse) deployKateeJob(task manifest.DeployKatee, man manifest.Manifest, basePath string) atc.JobConfig {
	job := atc.JobConfig{
		Name:   task.GetName(),
		Serial: true,
	}

	deployKateeRunJob := c.runJob(createDeployKateeRunTask(task, man), man, false, basePath)
	deploymentStatusRunJob := c.runJob(createDeploymentStatusTask(task, man), man, false, basePath)

	var steps []atc.Step
	steps = append(steps, deployKateeRunJob.PlanSequence...)
	steps = append(steps, deploymentStatusRunJob.PlanSequence...)

	job.PlanSequence = steps
	return job
}

func createDeployKateeRunTask(task manifest.DeployKatee, man manifest.Manifest) manifest.Run {
	run := manifest.Run{
		Type:          "run",
		Name:          "Deploy to Katee",
		ManualTrigger: false,
		Script: `\echo "Running vela up..."

if [ "$DOCKER_TAG" == "gitref" ]
then
  export TAG="$GIT_REVISION"
else
  export TAG="$BUILD_VERSION"
fi

export KATEE_APPLICATION_IMAGE=$KATEE_IMAGE:$TAG

/exe vela up -f $KATEE_APPFILE --publish-version $TAG`,
		Docker: manifest.Docker{
			Image:    "eu.gcr.io/halfpipe-io/ee-katee-vela-cli:latest",
			Username: "_json_key",
			Password: "((halfpipe-gcr.private_key))",
		},
		Privileged: false,
		Vars: manifest.Vars{
			"KATEE_TEAM":             man.Team,
			"KATEE_APPFILE":          task.VelaManifest,
			"KATEE_APPLICATION_NAME": task.ApplicationName,
			"KATEE_IMAGE":            task.Image,
			"KATEE_GKE_CREDENTIALS": fmt.Sprintf(
				`((katee-%s-service-account-prod.key))`, man.Team),
		},
		Retries:         task.Retries,
		NotifyOnSuccess: task.NotifyOnSuccess,
		Notifications:   task.Notifications,
		Timeout:         task.Timeout,
		BuildHistory:    task.BuildHistory,
	}

	if task.Tag == "gitref" {
		run.Vars["DOCKER_TAG"] = "gitref"
	} else if task.Tag == "version" {
		run.Vars["DOCKER_TAG"] = "buildVersion"
	}

	for k, v := range task.Vars {
		run.Vars[k] = v
	}

	return run
}

func createDeploymentStatusTask(task manifest.DeployKatee, man manifest.Manifest) manifest.Run {
	deploymentStatus := manifest.Run{
		Type:          "run",
		Name:          "Check Deployment Status",
		ManualTrigger: false,
		Docker: manifest.Docker{
			Image:    "eu.gcr.io/halfpipe-io/ee-katee-vela-cli:latest",
			Username: "_json_key",
			Password: "((halfpipe-gcr.private_key))",
		},
		Privileged: false,
		Vars: manifest.Vars{
			"KATEE_TEAM":       man.Team,
			"APPLICATION_NAME": task.ApplicationName,
			"KATEE_APPFILE":    task.VelaManifest,
			"KATEE_GKE_CREDENTIALS": fmt.Sprintf(
				`((katee-%s-service-account-prod.key))`, man.Team),
		},
		Retries:         1,
		NotifyOnSuccess: task.NotifyOnSuccess,
		Script: `\echo "Checking Deployment Status.."
if [ "$DOCKER_TAG" == "gitref" ]
then
  export PUBLISHED_VERSION=$GIT_REVISION
else
  export PUBLISHED_VERSION=$BUILD_VERSION
fi

/exe deployment-status katee-$KATEE_TEAM $APPLICATION_NAME $PUBLISHED_VERSION`,
	}

	if task.Tag == "gitref" {
		deploymentStatus.Vars["DOCKER_TAG"] = "gitref"
	} else if task.Tag == "version" {
		deploymentStatus.Vars["DOCKER_TAG"] = "buildVersion"
	}
	return deploymentStatus
}