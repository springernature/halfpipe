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
	gitUri := fmt.Sprintf("git@github.com:springernature/%s.git", name)
	man := manifest.Manifest{}
	man.Repo.Uri = gitUri
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
	gitUri := fmt.Sprintf("git@github.com:springernature/%s.git", name)
	man := manifest.Manifest{}
	man.Repo.Uri = gitUri
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
