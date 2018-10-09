package pipeline

import (
	"fmt"
	"testing"

	"path"

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
	expectedRunScript := runScriptArgs("./build.sh", true, "", "../artifacts-out", false, []string{"build/lib"}, ".git/ref")
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
	expectedRunScript := runScriptArgs("./build.sh", true, "", "../../../artifacts-out", false, []string{"build/lib"}, "../../.git/ref")
	assert.Equal(t, expectedRunScript, renderedPipeline.Jobs[0].Plan[1].TaskConfig.Run.Args)
}

func TestRendersPipelineWithCorrectResourceIfOverridingArtifactoryConfig(t *testing.T) {
	gitURI := "git@github.com:springernature/myRepo.git"

	man := manifest.Manifest{
		Team:     "team",
		Pipeline: "pipeline",
		ArtifactConfig: manifest.ArtifactConfig{
			Bucket:  "((override.Bucket))",
			JsonKey: "((override.JsonKey))",
		},
		Repo: manifest.Repo{
			URI:      gitURI,
			BasePath: "apps/subapp1",
		},
		Tasks: []manifest.Task{
			manifest.Run{
				Script:        "./build.sh",
				SaveArtifacts: []string{"build/lib/artifact.jar"},
			},
		},
	}
	artifactsResource := fmt.Sprintf("%s-%s-%s", artifactsDir, man.Team, man.Pipeline)

	renderedPipeline := testPipeline().Render(man)
	assert.Len(t, renderedPipeline.Jobs[0].Plan, 3)
	assert.Equal(t, artifactsName, renderedPipeline.Jobs[0].Plan[2].Put)
	assert.Equal(t, artifactsResource, renderedPipeline.Jobs[0].Plan[2].Resource)
	assert.Equal(t, artifactsOutDir, renderedPipeline.Jobs[0].Plan[2].Params["folder"])
	assert.Equal(t, gitDir+"/.git/ref", renderedPipeline.Jobs[0].Plan[2].Params["version_file"])

	resourceType, _ := renderedPipeline.ResourceTypes.Lookup("gcp-resource")
	assert.NotNil(t, resourceType)
	assert.Equal(t, "platformengineering/gcp-resource", resourceType.Source["repository"])
	assert.NotEmpty(t, resourceType.Source["tag"])

	resource, _ := renderedPipeline.Resources.Lookup(GenerateArtifactsResourceName(man.Team, man.Pipeline))
	assert.NotNil(t, resource)
	assert.Equal(t, man.ArtifactConfig.Bucket, resource.Source["bucket"])
	assert.Equal(t, man.ArtifactConfig.JsonKey, resource.Source["json_key"])
}

func TestRendersPipelineWithDeployArtifacts(t *testing.T) {
	gitURI := "git@github.com:springernature/myRepo.git"
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
	artifactsResource := fmt.Sprintf("%s-%s-%s", artifactsDir, man.Team, man.Pipeline)

	assert.Len(t, renderedPipeline.Jobs, 1)
	assert.Len(t, renderedPipeline.Jobs[0].Plan, 3)

	assert.Equal(t, artifactsDir, (*renderedPipeline.Jobs[0].Plan[0].Aggregate)[1].Get)
	assert.Equal(t, artifactsResource, (*renderedPipeline.Jobs[0].Plan[0].Aggregate)[1].Resource)
	assert.Equal(t, gitDir+"/.git/ref", (*renderedPipeline.Jobs[0].Plan[0].Aggregate)[1].Params["version_file"])

	resourceType, _ := renderedPipeline.ResourceTypes.Lookup("gcp-resource")
	assert.NotNil(t, resourceType)
	assert.Equal(t, "platformengineering/gcp-resource", resourceType.Source["repository"])
	assert.NotEmpty(t, resourceType.Source["tag"])

	resource, _ := renderedPipeline.Resources.Lookup(GenerateArtifactsResourceName(man.Team, man.Pipeline))
	assert.NotNil(t, resource)
	assert.Equal(t, "halfpipe-io-artifacts", resource.Source["bucket"])
	assert.Equal(t, path.Join(man.Team, man.Pipeline), resource.Source["folder"])
	assert.Equal(t, "((gcr.private_key))", resource.Source["json_key"])
}

func TestRenderPipelineWithSaveAndDeploy(t *testing.T) {
	repoName := "yolo"
	gitURI := fmt.Sprintf("git@github.com:springernature/%s.git", repoName)
	man := manifest.Manifest{}
	man.Team = "team"
	man.Pipeline = "pipeline"
	man.Repo.URI = gitURI
	man.Repo.BasePath = "apps/subapp1"

	deployArtifactPath := "build/lib/artifact.jar"
	man.Tasks = []manifest.Task{
		manifest.Run{
			Script:        "./build.sh",
			SaveArtifacts: []string{"build/lib"},
		},
		manifest.DeployCF{
			DeployArtifact: deployArtifactPath,
		},
	}

	renderedPipeline := testPipeline().Render(man)
	artifactsResource := fmt.Sprintf("%s-%s-%s", artifactsDir, man.Team, man.Pipeline)

	assert.Len(t, renderedPipeline.Jobs, 2)
	assert.Len(t, renderedPipeline.Jobs[0].Plan, 3)
	assert.Len(t, renderedPipeline.Jobs[1].Plan, 3)

	// order of the plans is important
	assert.Equal(t, artifactsResource, (*renderedPipeline.Jobs[1].Plan[0].Aggregate)[1].Resource)
	assert.Equal(t, "cf halfpipe-push", renderedPipeline.Jobs[1].Plan[1].Put)

	expectedAppPath := fmt.Sprintf("%s/%s", artifactsDir, deployArtifactPath)
	assert.Equal(t, expectedAppPath, renderedPipeline.Jobs[1].Plan[1].Params["appPath"])
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
	artifactsResource := fmt.Sprintf("%s-%s-%s", artifactsDir, man.Team, man.Pipeline)

	assert.Len(t, renderedPipeline.Jobs, 2)
	assert.Len(t, renderedPipeline.Jobs[0].Plan, 3)
	assert.Len(t, renderedPipeline.Jobs[1].Plan, 3)

	// order if the plans is important
	assert.Equal(t, artifactsResource, (*renderedPipeline.Jobs[1].Plan[0].Aggregate)[1].Resource)
	assert.Equal(t, "cf halfpipe-push", renderedPipeline.Jobs[1].Plan[1].Put)
	assert.Equal(t, artifactsDir+"/build/lib/artifact.jar", renderedPipeline.Jobs[1].Plan[1].Params["appPath"])
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

func TestRenderRunWithBothRestoreAndSave(t *testing.T) {
	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.Run{
				RestoreArtifacts: true,
				SaveArtifacts: []string{
					".",
				},
			},
		},
	}

	config := testPipeline().Render(man)

	hasArtifactGet := func() bool {
		for _, task := range *config.Jobs[0].Plan[0].Aggregate {
			if task.Get == artifactsDir {
				return true
			}
		}
		return false
	}

	assert.True(t, hasArtifactGet())
	assert.Equal(t, "artifacts-out", config.Jobs[0].Plan[1].TaskConfig.Outputs[0].Name)
}
