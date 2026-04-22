package shared

import (
	"fmt"
	"path"
	"strings"

	"github.com/springernature/halfpipe/config"

	"github.com/springernature/halfpipe/manifest"
)

func ConvertDeployMLModules(mlTask manifest.DeployMLModules, man manifest.Manifest) manifest.Run {
	runTask := manifest.Run{
		Name:   mlTask.Name,
		Script: "/ml-deploy/deploy-ml-modules",
		Docker: manifest.Docker{
			Image:    path.Join(config.DockerRegistry, "halfpipe-ml-deploy"),
			Username: "oauth2accesstoken",
			Password: "((gcp:platform-gar/token.token))",
		},
		Vars: manifest.Vars{
			"MARKLOGIC_HOST":       strings.Join(mlTask.Targets, ","),
			"MARKLOGIC_USERNAME":   mlTask.Username,
			"MARKLOGIC_PASSWORD":   mlTask.Password,
			"APP_NAME":             defaultValue(mlTask.AppName, man.Pipeline),
			"ARTIFACTORY_USERNAME": "((artifactory.username))",
			"ARTIFACTORY_PASSWORD": "((artifactory.password))",
			"ML_MODULES_VERSION":   mlTask.MLModulesVersion,
			"USE_BUILD_VERSION":    fmt.Sprint(mlTask.UseBuildVersion),
		},
		TaskBase: mlTask.TaskBase,
	}

	if mlTask.AppVersion != "" {
		runTask.Vars["APP_VERSION"] = mlTask.AppVersion
	}
	return runTask
}

func ConvertDeployMLZip(mlTask manifest.DeployMLZip, man manifest.Manifest) manifest.Run {
	runTask := manifest.Run{
		Name:   mlTask.Name,
		Script: "/ml-deploy/deploy-local-zip",
		Docker: manifest.Docker{
			Image:    path.Join(config.DockerRegistry, "halfpipe-ml-deploy"),
			Username: "oauth2accesstoken",
			Password: "((gcp:platform-gar/token.token))",
		},
		Vars: manifest.Vars{
			"MARKLOGIC_HOST":     strings.Join(mlTask.Targets, ","),
			"MARKLOGIC_USERNAME": mlTask.Username,
			"MARKLOGIC_PASSWORD": mlTask.Password,
			"APP_NAME":           defaultValue(mlTask.AppName, man.Pipeline),
			"DEPLOY_ZIP":         mlTask.DeployZip,
			"USE_BUILD_VERSION":  fmt.Sprint(mlTask.UseBuildVersion),
		},
		RestoreArtifacts: true,
		TaskBase:         mlTask.TaskBase,
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
