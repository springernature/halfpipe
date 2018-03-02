package linters

import (
	"testing"

	"github.com/springernature/halfpipe/manifest"
)

func TestCanOnlyHaveOneTaskThatSavesArtifactsInPipeline(t *testing.T) {
	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.Run{
				SaveArtifacts: []string{
					"a",
				},
			},
			manifest.Run{
				SaveArtifacts: []string{
					"b",
				},
			},
		},
	}

	result := artifactsLinter{}.Lint(man)
	assertInvalidFieldInErrors(t, "run.save_artifact", result.Errors)

	// Good!

	man = manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.Run{},
			manifest.Run{
				SaveArtifacts: []string{
					"b",
				},
			},
		},
	}

	result = artifactsLinter{}.Lint(man)
	assertInvalidFieldShouldNotBeInErrors(t, "run.save_artifact", result.Errors)
}

func TestWeOnlySupportSavingOfOneArtifactInPipeline(t *testing.T) {
	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.Run{
				SaveArtifacts: []string{
					"a",
					"b",
				},
			},
		},
	}

	result := artifactsLinter{}.Lint(man)
	assertInvalidFieldInErrors(t, "run.save_artifact", result.Errors)
}

func TestDeployArtifact(t *testing.T) {

	// No previous jobs have defined a SaveArtifact
	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DeployCF{
				DeployArtifact: "b",
			},
		},
	}

	result := artifactsLinter{}.Lint(man)
	assertInvalidFieldInErrors(t, "deploy-cf.deploy_artifact", result.Errors)

	// Different name of the artifacts
	man = manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.Run{
				SaveArtifacts: []string{
					"a",
				},
			},
			manifest.DeployCF{
				DeployArtifact: "b",
			},
		},
	}

	result = artifactsLinter{}.Lint(man)
	assertInvalidFieldInErrors(t, "deploy-cf.deploy_artifact", result.Errors)

	// Alles OK!
	man = manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.Run{
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

	result = artifactsLinter{}.Lint(man)
	assertInvalidFieldShouldNotBeInErrors(t, "deploy-cf.deploy_artifact", result.Errors)

}
