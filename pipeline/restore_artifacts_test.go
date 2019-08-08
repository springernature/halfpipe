package pipeline

import (
	"fmt"
	"testing"

	"path"

	"github.com/concourse/concourse/atc"
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

	renderedPipeline := testPipeline().Render(man)
	getArtifact := renderedPipeline.Jobs[0].Plan[1]
	assert.Equal(t, "get artifact", getArtifact.Name())
	assert.Equal(t, "git", getArtifact.TaskConfig.Inputs[0].Name)
	assert.Equal(t, artifactsInDir, getArtifact.TaskConfig.Outputs[0].Name)

	assert.Contains(t, renderedPipeline.Jobs[0].Plan[2].TaskConfig.Inputs, atc.TaskInputConfig{Name: artifactsInDir})
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

	renderedPipeline := testPipeline().Render(man)

	assert.Equal(t, "get artifact", renderedPipeline.Jobs[0].Plan[1].Name())
	assert.Contains(t, renderedPipeline.Jobs[0].Plan[2].TaskConfig.Inputs, atc.TaskInputConfig{Name: artifactsName})
}

func TestRendersPipelineWithArtifactsAsInputForDockerPushTask(t *testing.T) {
	// Single app repo
	gitURI := fmt.Sprintf("git@github.com:springernature/%s.git", "yolo")
	man := manifest.Manifest{
		Team:     "kehe",
		Pipeline: "Yolo",
		Repo: manifest.Repo{
			URI: gitURI,
		},
		Tasks: []manifest.Task{
			manifest.DockerPush{
				Image:            "somethigs/halfpipe-cli",
				RestoreArtifacts: true,
			},
		},
	}

	renderedPipeline := testPipeline().Render(man)
	assert.Equal(t, restoreArtifactTask(man), renderedPipeline.Jobs[0].Plan[1])

	runtTaskArgs := renderedPipeline.Jobs[0].Plan[2].TaskConfig.Run.Args[1]
	assert.Contains(t, runtTaskArgs, "cp -r git/. docker_build")
	assert.Contains(t, runtTaskArgs, "cp -r artifacts/. docker_build")

	assert.Equal(t, "halfpipe-cli", renderedPipeline.Jobs[0].Plan[3].Put)
	assert.Equal(t, dockerBuildTmpDir, renderedPipeline.Jobs[0].Plan[3].Params["build"])

	// Mono repo
	man = manifest.Manifest{
		Team:     "kehe",
		Pipeline: "Yolo",
		Repo: manifest.Repo{
			URI:      gitURI,
			BasePath: "some/random/path",
		},
		Tasks: []manifest.Task{
			manifest.DockerPush{
				Image:            "something/halfpipe-cli",
				RestoreArtifacts: true,
			},
		},
	}

	renderedPipeline = testPipeline().Render(man)
	assert.Equal(t, restoreArtifactTask(man), renderedPipeline.Jobs[0].Plan[1])

	runtTaskArgs = renderedPipeline.Jobs[0].Plan[2].TaskConfig.Run.Args[1]
	assert.Contains(t, runtTaskArgs, "cp -r git/. docker_build")
	assert.Contains(t, runtTaskArgs, "cp -r artifacts/. docker_build/some/random/path")

	assert.Equal(t, "halfpipe-cli", renderedPipeline.Jobs[0].Plan[3].Put)
	assert.Equal(t, path.Join(dockerBuildTmpDir, man.Repo.BasePath), renderedPipeline.Jobs[0].Plan[3].Params["build"])
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
	assert.Contains(t, runtTaskArgs, "cp -r ../artifacts/. .")

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
	assert.Contains(t, runtTaskArgs, "cp -r ../../../artifacts/. .")
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
	assert.Contains(t, runtTaskArgs, "cp -r ../artifacts/. .")
}
