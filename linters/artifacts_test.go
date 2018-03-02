package linters

import (
	"testing"

	"github.com/springernature/halfpipe/parser"
)

func TestCanOnlyHaveOneTaskThatSavesArtifactsInPipeline(t *testing.T) {
	man := parser.Manifest{
		Tasks: []parser.Task{
			parser.Run{
				SaveArtifacts: []string{
					"a",
				},
			},
			parser.Run{
				SaveArtifacts: []string{
					"b",
				},
			},
		},
	}

	result := ArtifactsLinter{}.Lint(man)
	assertInvalidFieldInErrors(t, "run.save_artifact", result.Errors)

	// Good!

	man = parser.Manifest{
		Tasks: []parser.Task{
			parser.Run{},
			parser.Run{
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
	man := parser.Manifest{
		Tasks: []parser.Task{
			parser.Run{
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
	man := parser.Manifest{
		Tasks: []parser.Task{
			parser.DeployCF{
				DeployArtifact: "b",
			},
		},
	}

	result := ArtifactsLinter{}.Lint(man)
	assertInvalidFieldInErrors(t, "deploy-cf.deploy_artifact", result.Errors)

	// Different name of the artifacts
	man = parser.Manifest{
		Tasks: []parser.Task{
			parser.Run{
				SaveArtifacts: []string{
					"a",
				},
			},
			parser.DeployCF{
				DeployArtifact: "b",
			},
		},
	}

	result = ArtifactsLinter{}.Lint(man)
	assertInvalidFieldInErrors(t, "deploy-cf.deploy_artifact", result.Errors)

	// Alles OK!
	man = parser.Manifest{
		Tasks: []parser.Task{
			parser.Run{
				SaveArtifacts: []string{
					"a",
				},
			},
			parser.Run{},
			parser.DeployCF{
				DeployArtifact: "a",
			},
		},
	}

	result = ArtifactsLinter{}.Lint(man)
	assertInvalidFieldShouldNotBeInErrors(t, "deploy-cf.deploy_artifact", result.Errors)

}
