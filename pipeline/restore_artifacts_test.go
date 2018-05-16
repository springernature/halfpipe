package pipeline

import (
	"fmt"
	"testing"

	"github.com/concourse/atc"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestRendersPipelineWithArtifactsAsInputForRunTask(t *testing.T) {
	gitURI := fmt.Sprintf("git@github.com:springernature/%s.git", "yolo")
	man := manifest.Manifest{
		Team:     "kehe",
		Pipeline: "Yolo",
		Repo: manifest.Repo{
			URI: gitURI,
		},
		Tasks: []manifest.Task{
			manifest.Run{
				Script:           "./build.sh",
				RestoreArtifacts: true,
			},
		},
	}

	artifactResourceName := GenerateArtifactsFolderName(man.Team, man.Pipeline)
	renderedPipeline := testPipeline().Render(man)
	assert.Equal(t, artifactResourceName, renderedPipeline.Jobs[0].Plan[1].Get)
	assert.Contains(t, renderedPipeline.Jobs[0].Plan[2].TaskConfig.Inputs, atc.TaskInputConfig{Name: GenerateArtifactsFolderName(man.Team, man.Pipeline), Path: artifactsFolderName})
}

func TestRendersPipelineWithArtifactsAsInputForDockerComposeTask(t *testing.T) {
	gitURI := fmt.Sprintf("git@github.com:springernature/%s.git", "yolo")
	man := manifest.Manifest{
		Team:     "kehe",
		Pipeline: "Yolo",
		Repo: manifest.Repo{
			URI: gitURI,
		},
		Tasks: []manifest.Task{
			manifest.DockerCompose{
				RestoreArtifacts: true,
			},
		},
	}

	artifactResourceName := GenerateArtifactsFolderName(man.Team, man.Pipeline)
	renderedPipeline := testPipeline().Render(man)
	assert.Equal(t, artifactResourceName, renderedPipeline.Jobs[0].Plan[1].Get)
	assert.Contains(t, renderedPipeline.Jobs[0].Plan[2].TaskConfig.Inputs, atc.TaskInputConfig{Name: GenerateArtifactsFolderName(man.Team, man.Pipeline), Path: artifactsFolderName})
}

func TestRendersPipelineWithArtifactsBeingCopiedIntoTheWorkingDirForRunTask(t *testing.T) {
	// Single app repo
	gitURI := fmt.Sprintf("git@github.com:springernature/%s.git", "yolo")
	man := manifest.Manifest{
		Team:     "kehe",
		Pipeline: "Yolo",
		Repo: manifest.Repo{
			URI: gitURI,
		},
		Tasks: []manifest.Task{
			manifest.Run{
				Script:           "./build.sh",
				RestoreArtifacts: true,
			},
		},
	}

	renderedPipeline := testPipeline().Render(man)

	runtTaskArgs := renderedPipeline.Jobs[0].Plan[2].TaskConfig.Run.Args[1]
	assert.Contains(t, runtTaskArgs, "cp -r ../artifacts/* .")

	// Monorepo
	man = manifest.Manifest{
		Team:     "kehe",
		Pipeline: "Yolo",
		Repo: manifest.Repo{
			URI:      gitURI,
			BasePath: "some/subfolder",
		},
		Tasks: []manifest.Task{
			manifest.Run{
				Script:           "./build.sh",
				RestoreArtifacts: true,
			},
		},
	}

	renderedPipeline = testPipeline().Render(man)

	runtTaskArgs = renderedPipeline.Jobs[0].Plan[2].TaskConfig.Run.Args[1]
	assert.Contains(t, runtTaskArgs, "cp -r ../../../artifacts/* .")
}

func TestRendersPipelineWithArtifactsBeingCopiedIntoTheWorkingDirForDockerCompose(t *testing.T) {

	gitURI := fmt.Sprintf("git@github.com:springernature/%s.git", "yolo")
	man := manifest.Manifest{
		Team:     "kehe",
		Pipeline: "Yolo",
		Repo: manifest.Repo{
			URI: gitURI,
		},
		Tasks: []manifest.Task{
			manifest.DockerCompose{
				RestoreArtifacts: true,
			},
		},
	}

	renderedPipeline := testPipeline().Render(man)

	runtTaskArgs := renderedPipeline.Jobs[0].Plan[2].TaskConfig.Run.Args[1]
	assert.Contains(t, runtTaskArgs, "cp -r ../artifacts/* .")
}
