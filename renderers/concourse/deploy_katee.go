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

export TAG="${BUILD_VERSION:-$GIT_REVISION}-$(date +%s)"
if [ "$REVISION_FORMAT" == "gitref" ]; then
  export TAG="$GIT_REVISION"
elif [ "$REVISION_FORMAT" == "version" ]; then
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
			"KATEE_NAMESPACE":       task.Namespace,
			"KATEE_APPFILE":         task.VelaManifest,
			"KATEE_GKE_CREDENTIALS": fmt.Sprintf(`((%s-service-account-prod.key))`, strings.Replace(task.Namespace, "katee", "katee-v2", 1)),
			"MAX_CHECKS":            strconv.Itoa(task.MaxChecks),
			"REVISION_FORMAT":       task.Tag,
		},
		Retries:         task.Retries,
		NotifyOnSuccess: task.NotifyOnSuccess,
		Notifications:   task.Notifications,
		Timeout:         task.Timeout,
	}

	for k, v := range task.Vars {
		run.Vars[k] = v
	}

	return run
}
