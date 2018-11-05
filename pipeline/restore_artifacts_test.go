package pipeline

import (
	"fmt"
	"testing"

	"path"

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

	artifactResourceName := GenerateArtifactsResourceName(man.Team, man.Pipeline)
	renderedPipeline := testPipeline().Render(man)
	assert.Equal(t, artifactResourceName, (*renderedPipeline.Jobs[0].Plan[0].Aggregate)[1].Resource)
	assert.Contains(t, renderedPipeline.Jobs[0].Plan[1].TaskConfig.Inputs, atc.TaskInputConfig{Name: artifactsName})
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

	artifactResourceName := GenerateArtifactsResourceName(man.Team, man.Pipeline)
	renderedPipeline := testPipeline().Render(man)
	assert.Equal(t, artifactResourceName, (*renderedPipeline.Jobs[0].Plan[0].Aggregate)[1].Resource)
	assert.Contains(t, renderedPipeline.Jobs[0].Plan[1].TaskConfig.Inputs, atc.TaskInputConfig{Name: artifactsName})
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
				RestoreArtifacts: true,
			},
		},
	}

	artifactResourceName := GenerateArtifactsResourceName(man.Team, man.Pipeline)
	renderedPipeline := testPipeline().Render(man)
	assert.Equal(t, artifactResourceName, (*renderedPipeline.Jobs[0].Plan[0].Aggregate)[1].Resource)

	runtTaskArgs := renderedPipeline.Jobs[0].Plan[1].TaskConfig.Run.Args[1]
	assert.Contains(t, runtTaskArgs, "cp -r git/. docker_build")
	assert.Contains(t, runtTaskArgs, "cp -r artifacts/. docker_build")

	assert.Equal(t, dockerPushResourceName, renderedPipeline.Jobs[0].Plan[2].Put)
	assert.Equal(t, dockerBuildTmpDir, renderedPipeline.Jobs[0].Plan[2].Params["build"])

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
				RestoreArtifacts: true,
			},
		},
	}

	artifactResourceName = GenerateArtifactsResourceName(man.Team, man.Pipeline)
	renderedPipeline = testPipeline().Render(man)
	assert.Equal(t, artifactResourceName, (*renderedPipeline.Jobs[0].Plan[0].Aggregate)[1].Resource)

	runtTaskArgs = renderedPipeline.Jobs[0].Plan[1].TaskConfig.Run.Args[1]
	assert.Contains(t, runtTaskArgs, "cp -r git/. docker_build")
	assert.Contains(t, runtTaskArgs, "cp -r artifacts/. docker_build/some/random/path")

	assert.Equal(t, dockerPushResourceName, renderedPipeline.Jobs[0].Plan[2].Put)
	assert.Equal(t, path.Join(dockerBuildTmpDir, man.Repo.BasePath), renderedPipeline.Jobs[0].Plan[2].Params["build"])
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

	runtTaskArgs := renderedPipeline.Jobs[0].Plan[1].TaskConfig.Run.Args[1]
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

	runtTaskArgs = renderedPipeline.Jobs[0].Plan[1].TaskConfig.Run.Args[1]
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

	runtTaskArgs := renderedPipeline.Jobs[0].Plan[1].TaskConfig.Run.Args[1]
	assert.Contains(t, runtTaskArgs, "cp -r ../artifacts/. .")
}
