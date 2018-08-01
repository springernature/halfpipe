package pipeline

import (
	"testing"

	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestConvertDeployMLLocalArtifactToRunTask(t *testing.T) {
	deployMl := manifest.DeployML{
		Name:           "foobar",
		Parallel:       true,
		DeployArtifact: "d-artifact",
		AppName:        "a-name",
		AppVersion:     "a-version",
		Targets:        []string{"blah", "blah1"},
		ManualTrigger:  true,
	}

	manif := manifest.Manifest{}

	exp := manifest.Run{
		Type:          "",
		Name:          "foobar",
		ManualTrigger: true,
		Script:        "/ml-deploy/deploy-local-zip",
		Docker: manifest.Docker{
			Image:    "eu.gcr.io/halfpipe-io/halfpipe-ml-deploy",
			Username: "_json_key",
			Password: "((gcr.private_key))",
		},
		Vars: manifest.Vars{
			"MARKLOGIC_HOST": "blah,blah1",
			"APP_NAME":       "a-name",
			"APP_VERSION":    "a-version",
			"DEPLOY_ZIP":     "d-artifact",
		},
		RestoreArtifacts: true,
		Parallel:         true,
	}

	act := ConvertDeployMLToRunTask(deployMl, manif)

	assert.Equal(t, exp, act)
}

func TestConvertDeployMLModulesToRunTask(t *testing.T) {
	deployMl := manifest.DeployML{
		Name:             "foobar",
		Parallel:         true,
		MLModulesVersion: "1.2345",
		AppName:          "a-name",
		AppVersion:       "a-version",
		Targets:          []string{"blah", "blah1"},
		ManualTrigger:    true,
	}

	manif := manifest.Manifest{}

	exp := manifest.Run{
		Type:          "",
		Name:          "foobar",
		ManualTrigger: true,
		Script:        "/ml-deploy/deploy-ml-modules",
		Docker: manifest.Docker{
			Image:    "eu.gcr.io/halfpipe-io/halfpipe-ml-deploy",
			Username: "_json_key",
			Password: "((gcr.private_key))",
		},
		Vars: manifest.Vars{
			"ARTIFACTORY_USER":     "((artifactory.username))",
			"ARTIFACTORY_PASSWORD": "((artifactory.password))",
			"MARKLOGIC_HOST":       "blah,blah1",
			"APP_NAME":             "a-name",
			"APP_VERSION":          "a-version",
			"ML_MODULES_VERSION":   "1.2345",
		},
		RestoreArtifacts: false,
		Parallel:         true,
	}

	act := ConvertDeployMLToRunTask(deployMl, manif)

	assert.Equal(t, exp, act)
}
