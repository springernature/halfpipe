package pipeline

import (
	"strings"

	"github.com/springernature/halfpipe/manifest"
)

func ConvertDeployMLToRunTask(mlTask manifest.DeployML, man manifest.Manifest) manifest.Run {
	runTask := manifest.Run{
		Name: mlTask.Name,
		Docker: manifest.Docker{
			Image:    "eu.gcr.io/halfpipe-io/halfpipe-ml-deploy",
			Username: "_json_key",
			Password: "((gcr.private_key))",
		},
		Vars: manifest.Vars{
			"MARKLOGIC_HOST": strings.Join(mlTask.Targets, ","),
			"APP_NAME":       mlTask.AppName,
			"APP_VERSION":    mlTask.AppVersion,
		},
		Parallel:      mlTask.Parallel,
		ManualTrigger: mlTask.ManualTrigger,
	}

	if mlTask.MLModulesVersion != "" {
		runTask.Script = "/ml-deploy/deploy-ml-modules"
		runTask.Vars["ML_MODULES_VERSION"] = mlTask.MLModulesVersion
		runTask.Vars["ARTIFACTORY_USER"] = "((artifactory.username))"
		runTask.Vars["ARTIFACTORY_PASSWORD"] = "((artifactory.password))"
	} else {
		runTask.Script = "/ml-deploy/deploy-local-zip"
		runTask.Vars["DEPLOY_ZIP"] = mlTask.DeployArtifact
		runTask.RestoreArtifacts = true
	}

	return runTask
}
