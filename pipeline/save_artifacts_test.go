package pipeline

import (
	"fmt"
	"testing"

	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestRendersPipelineWithOutputFolderAndFileCopyIfSaveArtifact(t *testing.T) {
	// Without any save artifact there should not be a copy and a output
	name := "yolo"
	gitURI := fmt.Sprintf("git@github.com:springernature/%s.git", name)
	man := manifest.Manifest{}
	man.Repo.URI = gitURI
	man.Tasks = []manifest.Task{
		manifest.Run{
			Script:        "./build.sh",
			SaveArtifacts: []string{"build/lib/artifact.jar"},
		},
	}

	renderedPipeline := testPipeline().Render(man)
	assert.Len(t, renderedPipeline.Jobs[0].Plan[1].TaskConfig.Outputs, 1) // Plan[0] is always the git get, Plan[1] is the task
	expected := `ARTIFACTS_DIR=../artifacts
./build.sh
if [ ! -e build/lib/artifact.jar ]; then
    echo "Artifact that should be at path 'build/lib/artifact.jar' not found! Bailing out"
    exit -1
fi

ARTIFACT_DIR_NAME=$(dirname build/lib/artifact.jar)
mkdir -p $ARTIFACTS_DIR/$ARTIFACT_DIR_NAME
cp build/lib/artifact.jar $ARTIFACTS_DIR/$ARTIFACT_DIR_NAME
`
	assert.Equal(t, expected, renderedPipeline.Jobs[0].Plan[1].TaskConfig.Run.Args[1])
}

func TestRendersPipelineWithOutputFolderAndFileCopyIfSaveArtifactInMonoRepo(t *testing.T) {
	// Without any save artifact there should not be a copy and a output
	name := "yolo"
	gitURI := fmt.Sprintf("git@github.com:springernature/%s.git", name)
	man := manifest.Manifest{}
	man.Repo.URI = gitURI
	man.Repo.BasePath = "apps/subapp1"
	man.Tasks = []manifest.Task{
		manifest.Run{
			Script:        "./build.sh",
			SaveArtifacts: []string{"build/lib/artifact.jar"},
		},
	}

	renderedPipeline := testPipeline().Render(man)
	assert.Len(t, renderedPipeline.Jobs[0].Plan[1].TaskConfig.Outputs, 1) // Plan[0] is always the git get, Plan[1] is the task
	expected := `ARTIFACTS_DIR=../../../artifacts
./build.sh
if [ ! -e build/lib/artifact.jar ]; then
    echo "Artifact that should be at path 'build/lib/artifact.jar' not found! Bailing out"
    exit -1
fi

ARTIFACT_DIR_NAME=$(dirname build/lib/artifact.jar)
mkdir -p $ARTIFACTS_DIR/$ARTIFACT_DIR_NAME
cp build/lib/artifact.jar $ARTIFACTS_DIR/$ARTIFACT_DIR_NAME
`
	assert.Equal(t, expected, renderedPipeline.Jobs[0].Plan[1].TaskConfig.Run.Args[1])
}

func TestRendersPipelineWithSaveArtifacts(t *testing.T) {
	// Without any save artifact there should not be a copy and a output
	name := "yolo"
	gitURI := fmt.Sprintf("git@github.com:springernature/%s.git", name)
	man := manifest.Manifest{}
	man.Repo.URI = gitURI
	man.Repo.BasePath = "apps/subapp1"
	man.Tasks = []manifest.Task{
		manifest.Run{
			Script:        "./build.sh",
			SaveArtifacts: []string{"build/lib/artifact.jar"},
		},
	}

	renderedPipeline := testPipeline().Render(man)
	assert.Len(t, renderedPipeline.Jobs[0].Plan, 3)
	assert.Equal(t, "artifact-storage", renderedPipeline.Jobs[0].Plan[2].Put)
	assert.Equal(t, "artifacts", renderedPipeline.Jobs[0].Plan[2].Params["folder"])
	assert.Equal(t, name+"/.git/ref", renderedPipeline.Jobs[0].Plan[2].Params["version_file"])

	resourceType, _ := renderedPipeline.ResourceTypes.Lookup("gcp-resource")
	assert.NotNil(t, resourceType)
	assert.Equal(t, "platformengineering/gcp-resource", resourceType.Source["repository"])
	assert.Equal(t, "latest", resourceType.Source["tag"])

	resource, _ := renderedPipeline.Resources.Lookup("artifact-storage")
	assert.NotNil(t, resource)
	assert.Equal(t, "halfpipe-artifacts", resource.Source["bucket"])
	assert.Equal(t, "((gcr.private_key))", resource.Source["json_key"])
}

func TestRendersPipelineWithDeployArtifacts(t *testing.T) {
	// Without any save artifact there should not be a copy and a output
	name := "yolo"
	gitURI := fmt.Sprintf("git@github.com:springernature/%s.git", name)
	man := manifest.Manifest{}
	man.Repo.URI = gitURI
	man.Repo.BasePath = "apps/subapp1"
	man.Tasks = []manifest.Task{
		manifest.DeployCF{
			DeployArtifact: "build/lib/artifact.jar",
		},
	}

	renderedPipeline := testPipeline().Render(man)
	assert.Len(t, renderedPipeline.Jobs, 1)
	assert.Len(t, renderedPipeline.Jobs[0].Plan, 3)

	assert.Equal(t, "artifact-storage", renderedPipeline.Jobs[0].Plan[2].Get)
	assert.Equal(t, name+"/.git/ref", renderedPipeline.Jobs[0].Plan[2].Params["version_file"])

	resourceType, _ := renderedPipeline.ResourceTypes.Lookup("gcp-resource")
	assert.NotNil(t, resourceType)
	assert.Equal(t, "platformengineering/gcp-resource", resourceType.Source["repository"])
	assert.Equal(t, "latest", resourceType.Source["tag"])

	resource, _ := renderedPipeline.Resources.Lookup("artifact-storage")
	assert.NotNil(t, resource)
	assert.Equal(t, "halfpipe-artifacts", resource.Source["bucket"])
	assert.Equal(t, "((gcr.private_key))", resource.Source["json_key"])
}
