package linters

import (
	"testing"

	"github.com/springernature/halfpipe/manifest"
)

func TestDeployArtifactRequiresASavedArtifact(t *testing.T) {

	// No previous tasks have defined a SaveArtifact
	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DeployCF{
				DeployArtifact: "b",
			},
		},
	}

	result := artifactsLinter{}.Lint(man)
	assertInvalidFieldInErrors(t, "deploy-cf.deploy_artifact", result.Errors)

	// A previous task has saved something
	man = manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.Run{
				SaveArtifacts: []string{
					"a",
				},
			},
			manifest.Run{},
			manifest.DeployCF{
				DeployArtifact: "b",
			},
		},
	}

	result = artifactsLinter{}.Lint(man)
	assertInvalidFieldShouldNotBeInErrors(t, "deploy-cf.deploy_artifact", result.Errors)
}

func TestItWorksWithDockerCompose(t *testing.T) {
	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DockerCompose{
				SaveArtifacts: []string{
					"a",
				},
			},
			manifest.Run{},
			manifest.DeployCF{
				DeployArtifact: "a",
			},
		},
	}

	result := artifactsLinter{}.Lint(man)
	assertInvalidFieldShouldNotBeInErrors(t, "deploy-cf.deploy_artifact", result.Errors)
}
