package concourse

import (
	"fmt"
	"github.com/concourse/concourse/atc"
	"github.com/springernature/halfpipe/manifest"
)

func (c Concourse) deployKateeJob(task manifest.DeployKatee, man manifest.Manifest, basePath string) atc.JobConfig {
	deployKatee := deployKatee{}
	deployKatee.task = task
	deployKatee.halfpipeManifest = man
	deployKatee.basePath = basePath
	deployKatee.vars = convertVars(task.Vars)

	job := atc.JobConfig{
		Name:   task.GetName(),
		Serial: true,
	}

	deployKateeRunJob := c.runJob(createDeployKateeRunTask(task, man, basePath), man, false, basePath)

	var steps []atc.Step
	steps = append(steps, deployKateeRunJob.PlanSequence...)

	job.PlanSequence = steps
	return job
}

func createDeployKateeRunTask(task manifest.DeployKatee, man manifest.Manifest, basePath string) manifest.Run {
	run := manifest.Run{
		Type:          "run",
		Name:          "Deploy to Katee",
		ManualTrigger: false,
		Script: `\echo "Running vela up...";
if [ "$DOCKER_TAG" == "gitref" ]
then
  export KATEE_APPLICATION_IMAGE=$KATEE_IMAGE:$GIT_REVISION;
else
  export KATEE_APPLICATION_IMAGE=$KATEE_IMAGE:$BUILD_VERSION;
fi; /exe vela up -f $KATEE_APPFILE --publish-version $DOCKER_TAG`,
		Docker: manifest.Docker{
			Image:    "eu.gcr.io/halfpipe-io/ee-katee-vela-cli:latest",
			Username: "_json_key",
			Password: "((halfpipe-gcr.private_key))",
		},
		Privileged: false,
		Vars: manifest.Vars{
			"KATEE_TEAM":             man.Team,
			"KATEE_APPFILE":          task.VelaAppFile,
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

	return run
}

type deployKatee struct {
	task             manifest.DeployKatee
	halfpipeManifest manifest.Manifest
	basePath         string
	vars             map[string]interface{}
}
