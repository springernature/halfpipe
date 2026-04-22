package concourse

import (
	"maps"
	"strconv"

	"github.com/concourse/concourse/atc"
	"github.com/springernature/halfpipe/manifest"
)

func (c Concourse) deployKateeJob(task manifest.DeployKatee, man manifest.Manifest, basePath string) (job atc.JobConfig) {
	job.Name = task.GetName()
	job.Serial = true

	deployKateeRunJob := c.runJob(createDeployKateeRunTask(task, man), man, basePath)
	job.PlanSequence = deployKateeRunJob.PlanSequence

	return
}

func createDeployKateeRunTask(task manifest.DeployKatee, man manifest.Manifest) manifest.Run {
	run := manifest.Run{
		Type: "run",
		Name: "Deploy to Katee",
		Script: `\export TAG="${BUILD_VERSION:-$GIT_REVISION}-$(date +%s)"
if [ "$REVISION_FORMAT" == "gitref" ]; then
  export TAG="$GIT_REVISION"
elif [ "$REVISION_FORMAT" == "version" ]; then
  export TAG="$BUILD_VERSION"
fi

halfpipe-deploy`,
		Docker: manifest.Docker{
			Image:    "eu.gcr.io/halfpipe-io/ee-run/docker/ee-katee-vela-cli:latest",
			Username: "oauth2accesstoken",
			Password: secrets.GARToken,
		},
		Privileged: false,
		Vars: manifest.Vars{
			"CHECK_INTERVAL":        strconv.Itoa(task.CheckInterval),
			"KATEE_NAMESPACE":       task.Namespace,
			"KATEE_APPFILE":         task.VelaManifest,
			"KATEE_GKE_CREDENTIALS": secrets.KateeKey(task.Namespace),
			"MAX_CHECKS":            strconv.Itoa(task.MaxChecks),
			"REVISION_FORMAT":       task.Tag,
		},
		TaskBase: task.TaskBase,
	}

	maps.Copy(run.Vars, task.Vars)
	if man.OpsLevel.System != "" {
		run.Vars["EAID"] = man.OpsLevel.System
	}

	return run
}
