package concourse

import (
	"fmt"
	"strconv"

	"github.com/concourse/concourse/atc"
	"github.com/springernature/halfpipe/manifest"
)

func (c Concourse) deployKateeJob(task manifest.DeployKatee, man manifest.Manifest, basePath string) (job atc.JobConfig) {
	job.Name = task.GetName()
	job.Serial = true

	deployKateeRunJob := c.runJob(createDeployKateeRunTask(task), man, false, basePath)
	job.PlanSequence = deployKateeRunJob.PlanSequence

	return
}

func createDeployKateeRunTask(task manifest.DeployKatee) manifest.Run {
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

halfpipe-deploy`,
		Docker: manifest.Docker{
			Image:    "eu.gcr.io/halfpipe-io/ee-katee-vela-cli:latest",
			Username: "_json_key",
			Password: "((halfpipe-gcr.private_key))",
		},
		Privileged: false,
		Vars: manifest.Vars{
			"KATEE_ENVIRONMENT":      task.Environment,
			"KATEE_NAMESPACE":        task.Namespace,
			"KATEE_PLATFORM_VERSION": task.PlatformVersion,
			"KATEE_APPFILE":          task.VelaManifest,
			"KATEE_GKE_CREDENTIALS":  fmt.Sprintf(`((%s-service-account-prod.key))`, task.Namespace),
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

	if task.DeploymentCheckTimeout != 0 {
		run.Vars["MAX_CHECKS"] = strconv.Itoa(task.DeploymentCheckTimeout)
	}

	for k, v := range task.Vars {
		run.Vars[k] = v
	}

	return run
}
