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
			SaveArtifacts: []string{"build/lib"},
		},
	}

	renderedPipeline := testPipeline().Render(man)
	assert.Len(t, renderedPipeline.Jobs[0].Plan[1].TaskConfig.Outputs, 1) // Plan[0] is always the git get, Plan[1] is the task
	expectedRunScript := runScriptArgs("./build.sh", "../artifacts", []string{"build/lib"}, ".git/ref")
	assert.Equal(t, expectedRunScript, renderedPipeline.Jobs[0].Plan[1].TaskConfig.Run.Args)
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
			SaveArtifacts: []string{"build/lib"},
		},
	}

	renderedPipeline := testPipeline().Render(man)
	assert.Len(t, renderedPipeline.Jobs[0].Plan[1].TaskConfig.Outputs, 1) // Plan[0] is always the git get, Plan[1] is the task
	expectedRunScript := runScriptArgs("./build.sh", "../../../artifacts", []string{"build/lib"}, "../../.git/ref")
	assert.Equal(t, expectedRunScript, renderedPipeline.Jobs[0].Plan[1].TaskConfig.Run.Args)
}

func TestRendersPipelineWithSaveArtifacts(t *testing.T) {
	name := "yolo"
	gitURI := fmt.Sprintf("git@github.com:springernature/%s.git", name)
	man := manifest.Manifest{}
	man.Team = "team"
	man.Pipeline = "pipeline"
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
	assert.Equal(t, "artifacts-team-pipeline", renderedPipeline.Jobs[0].Plan[2].Put)
	assert.Equal(t, "artifacts", renderedPipeline.Jobs[0].Plan[2].Params["folder"])
	assert.Equal(t, name+"/.git/ref", renderedPipeline.Jobs[0].Plan[2].Params["version_file"])

	resourceType, _ := renderedPipeline.ResourceTypes.Lookup("gcp-resource")
	assert.NotNil(t, resourceType)
	assert.Equal(t, "platformengineering/gcp-resource", resourceType.Source["repository"])
	assert.Equal(t, "stable", resourceType.Source["tag"])

	resource, _ := renderedPipeline.Resources.Lookup(GenerateArtifactsFolderName(man.Team, man.Pipeline))
	assert.NotNil(t, resource)
	assert.Equal(t, "halfpipe-artifacts", resource.Source["bucket"])
	assert.Equal(t, "((gcr.private_key))", resource.Source["json_key"])
}

func TestRendersPipelineWithDeployArtifacts(t *testing.T) {
	name := "yolo"
	gitURI := fmt.Sprintf("git@github.com:springernature/%s.git", name)
	man := manifest.Manifest{}
	man.Team = "team"
	man.Pipeline = "pipeline"
	man.Repo.URI = gitURI
	man.Repo.BasePath = "apps/subapp1"
	man.Tasks = []manifest.Task{
		manifest.DeployCF{
			DeployArtifact: "build/lib/artifact.jar",
		},
	}

	renderedPipeline := testPipeline().Render(man)
	assert.Len(t, renderedPipeline.Jobs, 1)
	assert.Len(t, renderedPipeline.Jobs[0].Plan, 5)

	assert.Equal(t, "artifacts-team-pipeline", renderedPipeline.Jobs[0].Plan[1].Get)
	assert.Equal(t, name+"/.git/ref", renderedPipeline.Jobs[0].Plan[1].Params["version_file"])

	resourceType, _ := renderedPipeline.ResourceTypes.Lookup("gcp-resource")
	assert.NotNil(t, resourceType)
	assert.Equal(t, "platformengineering/gcp-resource", resourceType.Source["repository"])
	assert.Equal(t, "stable", resourceType.Source["tag"])

	resource, _ := renderedPipeline.Resources.Lookup(GenerateArtifactsFolderName(man.Team, man.Pipeline))
	assert.NotNil(t, resource)
	assert.Equal(t, "halfpipe-artifacts", resource.Source["bucket"])
	assert.Equal(t, "((gcr.private_key))", resource.Source["json_key"])
}

func TestRenderPipelineWithSaveAndDeploy(t *testing.T) {
	name := "yolo"
	gitURI := fmt.Sprintf("git@github.com:springernature/%s.git", name)
	man := manifest.Manifest{}
	man.Team = "team"
	man.Pipeline = "pipeline"
	man.Repo.URI = gitURI
	man.Repo.BasePath = "apps/subapp1"
	man.Tasks = []manifest.Task{
		manifest.Run{
			Script:        "./build.sh",
			SaveArtifacts: []string{"build/lib"},
		},
		manifest.DeployCF{
			DeployArtifact: "build/lib/artifact.jar",
		},
	}

	renderedPipeline := testPipeline().Render(man)
	assert.Len(t, renderedPipeline.Jobs, 2)
	assert.Len(t, renderedPipeline.Jobs[0].Plan, 3)
	assert.Len(t, renderedPipeline.Jobs[1].Plan, 5)

	// order if the plans is important
	assert.Equal(t, "artifacts-team-pipeline", renderedPipeline.Jobs[1].Plan[1].Get)
	assert.Equal(t, "CF   ", renderedPipeline.Jobs[1].Plan[2].Put)
	assert.Equal(t, "artifacts-team-pipeline/build/lib/artifact.jar", renderedPipeline.Jobs[1].Plan[2].Params["appPath"])
}

func TestRenderPipelineWithSaveAndDeployInSingleAppRepo(t *testing.T) {
	name := "yolo"
	gitURI := fmt.Sprintf("git@github.com:springernature/%s.git", name)
	man := manifest.Manifest{}
	man.Team = "team"
	man.Pipeline = "pipeline"
	man.Repo.URI = gitURI
	man.Tasks = []manifest.Task{
		manifest.Run{
			Script:        "./build.sh",
			SaveArtifacts: []string{"build/lib"},
		},
		manifest.DeployCF{
			DeployArtifact: "build/lib/artifact.jar",
		},
	}

	renderedPipeline := testPipeline().Render(man)
	assert.Len(t, renderedPipeline.Jobs, 2)
	assert.Len(t, renderedPipeline.Jobs[0].Plan, 3)
	assert.Len(t, renderedPipeline.Jobs[1].Plan, 5)

	// order if the plans is important
	assert.Equal(t, "artifacts-team-pipeline", renderedPipeline.Jobs[1].Plan[1].Get)
	assert.Equal(t, "CF   ", renderedPipeline.Jobs[1].Plan[2].Put)
	assert.Equal(t, "artifacts-team-pipeline/build/lib/artifact.jar", renderedPipeline.Jobs[1].Plan[2].Params["appPath"])
}

func TestCopyArtifactScript(t *testing.T) {
	actual := copyArtifactScript("../../artifacts", "target/dist/artifact.jar")

	expected := `
if [ -d target/dist/artifact.jar ]
then
  mkdir -p ../../artifacts/target/dist/artifact.jar
  cp -r target/dist/artifact.jar/. ../../artifacts/target/dist/artifact.jar/
elif [ -f target/dist/artifact.jar ]
then
  artifactDir=$(dirname target/dist/artifact.jar)
  mkdir -p ../../artifacts/$artifactDir
  cp target/dist/artifact.jar ../../artifacts/$artifactDir
else
  echo "ERROR: Artifact 'target/dist/artifact.jar' not found. Try fly hijack to check the filesystem."
  exit 1
fi
`
	assert.Equal(t, expected, actual)
}
