package pipeline

import (
	"fmt"
	"testing"

	"github.com/concourse/atc"
	"github.com/springernature/halfpipe/model"
	"github.com/stretchr/testify/assert"
)

func testPipeline() Pipeline {
	return Pipeline{}
}

func TestRendersHttpGitResource(t *testing.T) {
	name := "yolo"
	gitUri := fmt.Sprintf("git@github.com:springernature/%s.git", name)

	manifest := model.Manifest{}
	manifest.Repo.Uri = gitUri

	expected := atc.Config{
		Resources: atc.ResourceConfigs{
			atc.ResourceConfig{
				Name: name,
				Type: "git",
				Source: atc.Source{
					"uri": gitUri,
				},
			},
		},
	}
	assert.Equal(t, expected, testPipeline().Render(model.Project{}, manifest))
}

func TestRendersSshGitResource(t *testing.T) {
	name := "asdf"
	gitUri := fmt.Sprintf("git@github.com:springernature/%s.git/", name)
	privateKey := "blurgh"

	manifest := model.Manifest{}
	manifest.Repo.Uri = gitUri
	manifest.Repo.PrivateKey = privateKey

	expected := atc.Config{
		Resources: atc.ResourceConfigs{
			atc.ResourceConfig{
				Name: name,
				Type: "git",
				Source: atc.Source{
					"uri":         gitUri,
					"private_key": privateKey,
				},
			},
		},
	}
	assert.Equal(t, expected, testPipeline().Render(model.Project{}, manifest))
}

func TestRendersGitResourceWithWatchesAndIgnores(t *testing.T) {
	name := "asdf"
	gitUri := fmt.Sprintf("git@github.com:springernature/%s.git/", name)
	privateKey := "blurgh"

	manifest := model.Manifest{}
	manifest.Repo.Uri = gitUri
	manifest.Repo.PrivateKey = privateKey

	watches := []string{"watch1", "watch2"}
	ignores := []string{"ignore1", "ignore2"}
	manifest.Repo.WatchedPaths = watches
	manifest.Repo.IgnoredPaths = ignores

	expected := atc.Config{
		Resources: atc.ResourceConfigs{
			atc.ResourceConfig{
				Name: name,
				Type: "git",
				Source: atc.Source{
					"uri":          gitUri,
					"private_key":  privateKey,
					"paths":        watches,
					"ignore_paths": ignores,
				},
			},
		},
	}
	assert.Equal(t, expected, testPipeline().Render(model.Project{}, manifest))
}

func TestRenderRunTask(t *testing.T) {
	manifest := model.Manifest{}
	manifest.Repo.Uri = "git@github.com:/springernature/foo.git"
	manifest.Tasks = []model.Task{
		model.Run{
			Script: "./yolo.sh",
			Image:  "imagename:TAG",
			Vars: map[string]string{
				"VAR1": "Value1",
				"VAR2": "Value2",
			},
		},
	}

	expected := atc.JobConfig{
		Name:   "run yolo.sh",
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{Get: manifest.Repo.GetName(), Trigger: true},
			atc.PlanConfig{Task: "./yolo.sh", TaskConfig: &atc.TaskConfig{
				Platform: "linux",
				Params: map[string]string{
					"VAR1": "Value1",
					"VAR2": "Value2",
				},
				ImageResource: &atc.ImageResource{
					Type: "docker-image",
					Source: atc.Source{
						"repository": "imagename",
						"tag":        "TAG",
					},
				},
				Run: atc.TaskRunConfig{
					Path: "/bin/sh",
					Dir:  manifest.Repo.GetName(),
					Args: []string{"-ec", fmt.Sprintf("./yolo.sh")},
				},
				Inputs: []atc.TaskInputConfig{
					{Name: manifest.Repo.GetName()},
				},
			}},
		}}

	assert.Equal(t, expected, testPipeline().Render(model.Project{}, manifest).Jobs[0])
}

func TestRenderRunTaskFromHalfpipeNotInRoot(t *testing.T) {
	manifest := model.Manifest{}
	manifest.Repo.Uri = "git@github.com:/springernature/foo.git"
	project := model.Project{BasePath: "subapp"}

	manifest.Tasks = []model.Task{
		model.Run{
			Script: "./yolo.sh",
			Image:  "imagename:TAG",
			Vars: map[string]string{
				"VAR1": "Value1",
				"VAR2": "Value2",
			},
		},
	}

	expected := atc.JobConfig{
		Name:   "run yolo.sh",
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{Get: manifest.Repo.GetName(), Trigger: true},
			atc.PlanConfig{Task: "./yolo.sh", TaskConfig: &atc.TaskConfig{
				Platform: "linux",
				Params: map[string]string{
					"VAR1": "Value1",
					"VAR2": "Value2",
				},
				ImageResource: &atc.ImageResource{
					Type: "docker-image",
					Source: atc.Source{
						"repository": "imagename",
						"tag":        "TAG",
					},
				},
				Run: atc.TaskRunConfig{
					Path: "/bin/sh",
					Dir:  manifest.Repo.GetName() + "/" + project.BasePath,
					Args: []string{"-ec", fmt.Sprintf("./yolo.sh")},
				},
				Inputs: []atc.TaskInputConfig{
					{Name: manifest.Repo.GetName()},
				},
			}},
		}}

	assert.Equal(t, expected, testPipeline().Render(project, manifest).Jobs[0])
}

func TestRenderDockerPushTask(t *testing.T) {
	manifest := model.Manifest{}
	manifest.Repo.Uri = "git@github.com:/springernature/foo.git"

	username := "halfpipe"
	password := "secret"
	repo := "halfpipe/halfpipe-cli"
	manifest.Tasks = []model.Task{
		model.DockerPush{
			Username: username,
			Password: password,
			Repo:     repo,
		},
	}

	expectedResource := atc.ResourceConfig{
		Name: "Docker Registry",
		Type: "docker-image",
		Source: atc.Source{
			"username":   username,
			"password":   password,
			"repository": repo,
		},
	}

	expectedJobConfig := atc.JobConfig{
		Name:   "docker-push",
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{Get: manifest.Repo.GetName(), Trigger: true},
			atc.PlanConfig{Put: "Docker Registry", Params: atc.Params{"build": manifest.Repo.GetName()}},
		},
	}

	// First resource will always be the git resource.
	assert.Equal(t, expectedResource, testPipeline().Render(model.Project{}, manifest).Resources[1])
	assert.Equal(t, expectedJobConfig, testPipeline().Render(model.Project{}, manifest).Jobs[0])
}

func TestRenderDockerPushTaskNotInRoot(t *testing.T) {
	manifest := model.Manifest{}
	manifest.Repo.Uri = "git@github.com:/springernature/foo.git"
	project := model.Project{BasePath: "subapp"}

	username := "halfpipe"
	password := "secret"
	repo := "halfpipe/halfpipe-cli"
	manifest.Tasks = []model.Task{
		model.DockerPush{
			Username: username,
			Password: password,
			Repo:     repo,
		},
	}

	expectedResource := atc.ResourceConfig{
		Name: "Docker Registry",
		Type: "docker-image",
		Source: atc.Source{
			"username":   username,
			"password":   password,
			"repository": repo,
		},
	}

	expectedJobConfig := atc.JobConfig{
		Name:   "docker-push",
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{Get: manifest.Repo.GetName(), Trigger: true},
			atc.PlanConfig{Put: "Docker Registry", Params: atc.Params{"build": manifest.Repo.GetName() + "/" + project.BasePath}},
		},
	}

	// First resource will always be the git resource.
	assert.Equal(t, expectedResource, testPipeline().Render(project, manifest).Resources[1])
	assert.Equal(t, expectedJobConfig, testPipeline().Render(project, manifest).Jobs[0])
}

func TestRenderWithTriggerTrueAndPassedOnPreviousTask(t *testing.T) {
	manifest := model.Manifest{
		Tasks: []model.Task{
			model.Run{Script: "asd.sh"},
			model.DeployCF{},
			model.DockerPush{},
		},
	}
	config := testPipeline().Render(model.Project{}, manifest)

	assert.Nil(t, config.Jobs[0].Plan[0].Passed)
	assert.Equal(t, config.Jobs[1].Plan[0].Passed[0], config.Jobs[0].Name)
	assert.Equal(t, config.Jobs[2].Plan[0].Passed[0], config.Jobs[1].Name)
}

func TestRendersHttpGitResourceWithGitCrypt(t *testing.T) {
	name := "yolo"
	gitUri := fmt.Sprintf("git@github.com:springernature/%s.git", name)
	gitCrypt := "AABBFF66"

	manifest := model.Manifest{}
	manifest.Repo.Uri = gitUri
	manifest.Repo.GitCryptKey = gitCrypt

	expected := atc.Config{
		Resources: atc.ResourceConfigs{
			atc.ResourceConfig{
				Name: name,
				Type: "git",
				Source: atc.Source{
					"uri":           gitUri,
					"git_crypt_key": gitCrypt,
				},
			},
		},
	}
	assert.Equal(t, expected, testPipeline().Render(model.Project{}, manifest))
}

func TestRendersPipelineWithOutputFolderAndFileCopyIfSaveArtifact(t *testing.T) {
	// Without any save artifact there should not be a copy and a output
	name := "yolo"
	gitUri := fmt.Sprintf("git@github.com:springernature/%s.git", name)
	manifest := model.Manifest{}
	manifest.Repo.Uri = gitUri
	manifest.Tasks = []model.Task{
		model.Run{
			Script:        "./build.sh",
			SaveArtifacts: []string{"build/lib/artifact.jar"},
		},
	}

	renderedPipeline := testPipeline().Render(model.Project{}, manifest)
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
	manifest := model.Manifest{}
	manifest.Repo.Uri = gitUri
	manifest.Tasks = []model.Task{
		model.Run{
			Script:        "./build.sh",
			SaveArtifacts: []string{"build/lib/artifact.jar"},
		},
	}

	project := model.Project{
		BasePath: "apps/subapp1",
	}
	renderedPipeline := testPipeline().Render(project, manifest)
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
