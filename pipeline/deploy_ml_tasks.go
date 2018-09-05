package pipeline

import (
	"strings"

	"github.com/springernature/halfpipe/manifest"
)

func ConvertDeployMLModulesToRunTask(mlTask manifest.DeployMLModules, man manifest.Manifest) manifest.Run {
	runTask := manifest.Run{
		Retries: mlTask.Retries,
		Name:     mlTask.Name,
		Script:   "/ml-deploy/deploy-ml-modules",
		Docker: manifest.Docker{
			Image:    "eu.gcr.io/halfpipe-io/halfpipe-ml-deploy",
			Username: "_json_key",
			Password: "((gcr.private_key))",
		},
		Vars: manifest.Vars{
			"MARKLOGIC_HOST":       strings.Join(mlTask.Targets, ","),
			"APP_NAME":             defaultValue(mlTask.AppName, man.Pipeline),
			"ARTIFACTORY_USERNAME": "((artifactory.username))",
			"ARTIFACTORY_PASSWORD": "((artifactory.password))",
			"ML_MODULES_VERSION":   mlTask.MLModulesVersion,
		},
		Parallel:      mlTask.Parallel,
		ManualTrigger: mlTask.ManualTrigger,
	}

	if mlTask.AppVersion != "" {
		runTask.Vars["APP_VERSION"] = mlTask.AppVersion
	}
	return runTask
}

func ConvertDeployMLZipToRunTask(mlTask manifest.DeployMLZip, man manifest.Manifest) manifest.Run {
	runTask := manifest.Run{
		Retries: mlTask.Retries,
		Name:     mlTask.Name,
		Script:   "/ml-deploy/deploy-local-zip",
		Docker: manifest.Docker{
			Image:    "eu.gcr.io/halfpipe-io/halfpipe-ml-deploy",
			Username: "_json_key",
			Password: "((gcr.private_key))",
		},
		Vars: manifest.Vars{
			"MARKLOGIC_HOST": strings.Join(mlTask.Targets, ","),
			"APP_NAME":       defaultValue(mlTask.AppName, man.Pipeline),
			"DEPLOY_ZIP":     mlTask.DeployZip,
		},
		Parallel:         mlTask.Parallel,
		ManualTrigger:    mlTask.ManualTrigger,
		RestoreArtifacts: true,
	}

	if mlTask.AppVersion != "" {
		runTask.Vars["APP_VERSION"] = mlTask.AppVersion
	}
	return runTask
}

func defaultValue(value, defaultValue string) string {
	if value != "" {
		return value
	}
	return defaultValue
}
