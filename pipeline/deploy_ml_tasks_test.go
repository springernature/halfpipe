package pipeline

import (
	"github.com/springernature/halfpipe/config"
	"testing"

	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestConvertDeployMLZipToRunTask(t *testing.T) {
	deployMl := manifest.DeployMLZip{
		Name:            "foobar",
		Parallel:        "true",
		DeployZip:       "d-artifact",
		AppName:         "a-name",
		AppVersion:      "a-version",
		Targets:         []string{"blah", "blah1"},
		ManualTrigger:   true,
		UseBuildVersion: true,
	}

	man := manifest.Manifest{}

	expected := manifest.Run{
		Type:          "",
		Name:          "foobar",
		ManualTrigger: true,
		Script:        "/ml-deploy/deploy-local-zip",
		Docker: manifest.Docker{
			Image:    "eu.gcr.io/" + config.Project + "/halfpipe-ml-deploy",
			Username: "_json_key",
			Password: "((halfpipe-gcr.private_key))",
		},
		Vars: manifest.Vars{
			"MARKLOGIC_HOST":    "blah,blah1",
			"APP_NAME":          "a-name",
			"APP_VERSION":       "a-version",
			"DEPLOY_ZIP":        "d-artifact",
			"USE_BUILD_VERSION": "true",
		},
		RestoreArtifacts: true,
		Parallel:         "true",
	}

	actual := ConvertDeployMLZipToRunTask(deployMl, man)

	assert.Equal(t, expected, actual)
}

func TestConvertDeployMLModulesToRunTask(t *testing.T) {
	deployMl := manifest.DeployMLModules{
		Name:             "foobar",
		Parallel:         "true",
		MLModulesVersion: "1.2345",
		AppName:          "a-name",
		AppVersion:       "a-version",
		Targets:          []string{"blah", "blah1"},
		ManualTrigger:    true,
	}

	man := manifest.Manifest{}

	expected := manifest.Run{
		Type:          "",
		Name:          "foobar",
		ManualTrigger: true,
		Script:        "/ml-deploy/deploy-ml-modules",
		Docker: manifest.Docker{
			Image:    "eu.gcr.io/" + config.Project + "/halfpipe-ml-deploy",
			Username: "_json_key",
			Password: "((halfpipe-gcr.private_key))",
		},
		Vars: manifest.Vars{
			"ARTIFACTORY_USERNAME": "((artifactory.username))",
			"ARTIFACTORY_PASSWORD": "((artifactory.password))",
			"MARKLOGIC_HOST":       "blah,blah1",
			"APP_NAME":             "a-name",
			"APP_VERSION":          "a-version",
			"ML_MODULES_VERSION":   "1.2345",
			"USE_BUILD_VERSION":    "false",
		},
		RestoreArtifacts: false,
		Parallel:         "true",
	}

	actual := ConvertDeployMLModulesToRunTask(deployMl, man)

	assert.Equal(t, expected, actual)
}

func TestDefaultAppNameToPipelineName(t *testing.T) {
	man := manifest.Manifest{Pipeline: "my-pipe"}

	tests := []struct {
		task            manifest.Run
		expectedAppName string
	}{
		{ConvertDeployMLModulesToRunTask(manifest.DeployMLModules{}, man), "my-pipe"},
		{ConvertDeployMLModulesToRunTask(manifest.DeployMLModules{AppName: "foo"}, man), "foo"},
		{ConvertDeployMLZipToRunTask(manifest.DeployMLZip{}, man), "my-pipe"},
		{ConvertDeployMLZipToRunTask(manifest.DeployMLZip{AppName: "foo"}, man), "foo"},
	}

	for _, test := range tests {
		assert.Equal(t, test.expectedAppName, test.task.Vars["APP_NAME"])
	}

}

func TestAppVersionOnlySetIfNotEmpty(t *testing.T) {
	man := manifest.Manifest{Pipeline: "my-pipe"}

	tests := []struct {
		task          manifest.Run
		appVersionSet bool
	}{
		{ConvertDeployMLModulesToRunTask(manifest.DeployMLModules{}, man), false},
		{ConvertDeployMLModulesToRunTask(manifest.DeployMLModules{AppVersion: "1.1"}, man), true},
		{ConvertDeployMLZipToRunTask(manifest.DeployMLZip{}, man), false},
		{ConvertDeployMLZipToRunTask(manifest.DeployMLZip{AppVersion: "1.1"}, man), true},
	}

	for _, test := range tests {
		_, exists := test.task.Vars["APP_VERSION"]
		assert.Equal(t, test.appVersionSet, exists)
	}

}
