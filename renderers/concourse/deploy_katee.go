package concourse

import (
	"fmt"
	"strconv"
	"strings"

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
			"CHECK_INTERVAL":        strconv.Itoa(task.CheckInterval),
			"KATEE_ENVIRONMENT":     strings.TrimPrefix(task.Namespace, "katee-"),
			"KATEE_NAMESPACE":       task.Namespace,
			"KATEE_APPFILE":         task.VelaManifest,
			"KATEE_GKE_CREDENTIALS": fmt.Sprintf(`((%s-service-account-prod.key))`, task.Namespace),
			"KATEE_V2_GKE_CREDS":    fmt.Sprintf(`((%s-service-account-prod.key))`, strings.Replace(task.Namespace, "katee", "katee-v2", 1)),
			"MAX_CHECKS":            strconv.Itoa(task.MaxChecks),
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
