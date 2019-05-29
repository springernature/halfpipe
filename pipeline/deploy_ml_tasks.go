package pipeline

import (
	"fmt"
	"github.com/springernature/halfpipe/config"
	"strings"

	"github.com/springernature/halfpipe/manifest"
)

func ConvertDeployMLModulesToRunTask(mlTask manifest.DeployMLModules, man manifest.Manifest) manifest.Run {
	runTask := manifest.Run{
		Retries: mlTask.Retries,
		Name:    mlTask.Name,
		Script:  "/ml-deploy/deploy-ml-modules",
		Docker: manifest.Docker{
			Image:    "eu.gcr.io/" + config.Project + "/halfpipe-ml-deploy",
			Username: "_json_key",
			Password: "((halfpipe-gcr.private_key))",
		},
		Vars: manifest.Vars{
			"MARKLOGIC_HOST":       strings.Join(mlTask.Targets, ","),
			"APP_NAME":             defaultValue(mlTask.AppName, man.Pipeline),
			"ARTIFACTORY_USERNAME": "((artifactory.username))",
			"ARTIFACTORY_PASSWORD": "((artifactory.password))",
			"ML_MODULES_VERSION":   mlTask.MLModulesVersion,
			"USE_BUILD_VERSION":    fmt.Sprint(mlTask.UseBuildVersion),
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
		Name:    mlTask.Name,
		Script:  "/ml-deploy/deploy-local-zip",
		Docker: manifest.Docker{
			Image:    "eu.gcr.io/" + config.Project + "/halfpipe-ml-deploy",
			Username: "_json_key",
			Password: "((halfpipe-gcr.private_key))",
		},
		Vars: manifest.Vars{
			"MARKLOGIC_HOST":    strings.Join(mlTask.Targets, ","),
			"APP_NAME":          defaultValue(mlTask.AppName, man.Pipeline),
			"DEPLOY_ZIP":        mlTask.DeployZip,
			"USE_BUILD_VERSION": fmt.Sprint(mlTask.UseBuildVersion),
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
