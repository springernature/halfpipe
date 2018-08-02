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

func TestRestoreArtifactsComplainIfNoPreviousTaskHaveSaved(t *testing.T) {
	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.Run{},
			manifest.Run{RestoreArtifacts: true},
			manifest.DockerCompose{RestoreArtifacts: true},
			manifest.DockerPush{RestoreArtifacts: true},
		},
	}

	result := artifactsLinter{}.Lint(man)
	assertInvalidFieldInErrors(t, "run.restore_artifacts", result.Errors)
	assertInvalidFieldInErrors(t, "docker-compose.restore_artifacts", result.Errors)
	assertInvalidFieldInErrors(t, "docker-push.restore_artifacts", result.Errors)
}

func TestRestoreArtifactsWorksIfPreviousTaskHaveSaved(t *testing.T) {
	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.Run{SaveArtifacts: []string{"some/path"}},
			manifest.Run{RestoreArtifacts: true},
			manifest.DockerCompose{RestoreArtifacts: true},
			manifest.DockerPush{RestoreArtifacts: true},
		},
	}

	result := artifactsLinter{}.Lint(man)
	assertInvalidFieldShouldNotBeInErrors(t, "run.restore_artifacts", result.Errors)
	assertInvalidFieldShouldNotBeInErrors(t, "docker-compose.restore_artifacts", result.Errors)
	assertInvalidFieldShouldNotBeInErrors(t, "docker-push.restore_artifacts", result.Errors)
}

func TestDeployMLZipRequiresASavedArtifact(t *testing.T) {

	// No previous tasks have defined a SaveArtifact
	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DeployMLZip{
				DeployZip: "b",
			},
		},
	}

	result := artifactsLinter{}.Lint(man)
	assertInvalidFieldInErrors(t, "deploy-ml-zip.deploy_zip", result.Errors)

	// A previous task has saved something
	man = manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.Run{
				SaveArtifacts: []string{
					"a",
				},
			},
			manifest.Run{},
			manifest.DeployMLZip{
				DeployZip: "b",
			},
		},
	}

	result = artifactsLinter{}.Lint(man)
	assertInvalidFieldShouldNotBeInErrors(t, "deploy-ml-zip.deploy_zip", result.Errors)
}
