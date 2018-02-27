package linters

import (
	"testing"

	"github.com/springernature/halfpipe/model"
)

func TestCanOnlyHaveOneTaskThatSavesArtifactsInPipeline(t *testing.T) {
	man := model.Manifest{
		Tasks: []model.Task{
			model.Run{
				SaveArtifacts: []string{
					"a",
				},
			},
			model.Run{
				SaveArtifacts: []string{
					"b",
				},
			},
		},
	}

	result := ArtifactsLinter{}.Lint(man)
	assertInvalidFieldInErrors(t, "run.save_artifact", result.Errors)

	// Good!

	man = model.Manifest{
		Tasks: []model.Task{
			model.Run{},
			model.Run{
				SaveArtifacts: []string{
					"b",
				},
			},
		},
	}

	result = ArtifactsLinter{}.Lint(man)
	assertInvalidFieldShouldNotBeInErrors(t, "run.save_artifact", result.Errors)
}

func TestWeOnlySupportSavingOfOneArtifactInPipeline(t *testing.T) {
	man := model.Manifest{
		Tasks: []model.Task{
			model.Run{
				SaveArtifacts: []string{
					"a",
					"b",
				},
			},
		},
	}

	result := ArtifactsLinter{}.Lint(man)
	assertInvalidFieldInErrors(t, "run.save_artifact", result.Errors)
}

func TestDeployArtifact(t *testing.T) {

	// No previous jobs have defined a SaveArtifact
	man := model.Manifest{
		Tasks: []model.Task{
			model.DeployCF{
				DeployArtifact: "b",
			},
		},
	}

	result := ArtifactsLinter{}.Lint(man)
	assertInvalidFieldInErrors(t, "deploy-cf.deploy_artifact", result.Errors)

	// Different name of the artifacts
	man = model.Manifest{
		Tasks: []model.Task{
			model.Run{
				SaveArtifacts: []string{
					"a",
				},
			},
			model.DeployCF{
				DeployArtifact: "b",
			},
		},
	}

	result = ArtifactsLinter{}.Lint(man)
	assertInvalidFieldInErrors(t, "deploy-cf.deploy_artifact", result.Errors)

	// Alles OK!
	man = model.Manifest{
		Tasks: []model.Task{
			model.Run{
				SaveArtifacts: []string{
					"a",
				},
			},
			model.Run{},
			model.DeployCF{
				DeployArtifact: "a",
			},
		},
	}

	result = ArtifactsLinter{}.Lint(man)
	assertInvalidFieldShouldNotBeInErrors(t, "deploy-cf.deploy_artifact", result.Errors)

}
