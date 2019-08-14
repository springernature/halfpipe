package pipeline

import (
	"fmt"
	"github.com/springernature/halfpipe/config"
	"strings"
	"testing"

	"path"

	"github.com/concourse/concourse/atc"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestRendersPipelineWithOutputFolderAndFileCopyIfSaveArtifact(t *testing.T) {
	// Without any save artifact there should not be a copy and a output
	man := manifest.Manifest{}
	runTask := manifest.Run{
		Script:        "./build.sh",
		SaveArtifacts: []string{"build/lib"},
	}
	man.Tasks = []manifest.Task{
		runTask,
	}

	renderedPipeline := testPipeline().Render(man)
	assert.Len(t, renderedPipeline.Jobs[0].Plan[1].TaskConfig.Outputs, 1) // Plan[0] is always the git get, Plan[1] is the task
	expectedRunScript := runScriptArgs(runTask, manifest.Manifest{}, true)
	assert.Equal(t, expectedRunScript, renderedPipeline.Jobs[0].Plan[1].TaskConfig.Run.Args)
}

func TestRendersPipelineFailureOutputFolderAndPut(t *testing.T) {
	man := manifest.Manifest{}

	run1 := "run1"
	run2 := "run2"
	run3 := "run3"
	dockerCompose1 := "dockerCompose1"
	dockerCompose2 := "dockerCompose2"
	dockerCompose3 := "dockerCompose3"

	team := "myTeam"
	pipeline := "myPipeline"

	man.Team = team
	man.Pipeline = pipeline
	man.Tasks = []manifest.Task{
		manifest.Run{
			Name:                   run1,
			Script:                 "./build.sh",
			SaveArtifacts:          []string{"build/lib"},
			SaveArtifactsOnFailure: []string{"test-reports"},
		},
		manifest.Run{
			Name:                   run2,
			Script:                 "./build.sh",
			SaveArtifactsOnFailure: []string{"test-reports"},
		},
		manifest.Run{
			Name:                   run3,
			Script:                 "./build.sh",
			RestoreArtifacts:       true,
			SaveArtifacts:          []string{"build/lib"},
			SaveArtifactsOnFailure: []string{"test-reports"},
		},
		manifest.DockerCompose{
			Name:                   dockerCompose1,
			SaveArtifacts:          []string{"build/lib"},
			SaveArtifactsOnFailure: []string{"test-reports"},
		},
		manifest.DockerCompose{
			Name:                   dockerCompose2,
			SaveArtifactsOnFailure: []string{"test-reports"},
		},
		manifest.DockerCompose{
			Name:                   dockerCompose3,
			RestoreArtifacts:       true,
			SaveArtifacts:          []string{"build/lib"},
			SaveArtifactsOnFailure: []string{"test-reports"},
		},
	}

	containsPut := func(putName string, config atc.JobConfig) bool {
		for _, c := range config.Plan {
			if c.Put == putName {
				return true
			}
		}
		return false
	}

	renderedPipeline := testPipeline().Render(man)
	config1, _ := renderedPipeline.Jobs.Lookup(run1)
	assert.Contains(t, config1.Plan[1].TaskConfig.Outputs, atc.TaskOutputConfig{Name: artifactsOutDir})
	assert.Contains(t, config1.Plan[1].TaskConfig.Outputs, atc.TaskOutputConfig{Name: artifactsOutDirOnFailure})
	assert.True(t, containsPut(artifactsName, config1))

	config2, _ := renderedPipeline.Jobs.Lookup(run2)
	assert.NotContains(t, config2.Plan[1].TaskConfig.Outputs, atc.TaskOutputConfig{Name: artifactsOutDir})
	assert.Contains(t, config2.Plan[1].TaskConfig.Outputs, atc.TaskOutputConfig{Name: artifactsOutDirOnFailure})
	assert.False(t, containsPut(artifactsName, config2))

	config3, _ := renderedPipeline.Jobs.Lookup(run3)
	assert.Contains(t, config3.Plan[2].TaskConfig.Outputs, atc.TaskOutputConfig{Name: artifactsOutDir})
	assert.Contains(t, config3.Plan[2].TaskConfig.Outputs, atc.TaskOutputConfig{Name: artifactsOutDirOnFailure})
	assert.True(t, containsPut(artifactsName, config3))

	config4, _ := renderedPipeline.Jobs.Lookup(dockerCompose1)
	assert.Contains(t, config4.Plan[1].TaskConfig.Outputs, atc.TaskOutputConfig{Name: artifactsOutDir})
	assert.Contains(t, config4.Plan[1].TaskConfig.Outputs, atc.TaskOutputConfig{Name: artifactsOutDirOnFailure})
	assert.True(t, containsPut(artifactsName, config4))

	config5, _ := renderedPipeline.Jobs.Lookup(dockerCompose2)
	assert.NotContains(t, config5.Plan[1].TaskConfig.Outputs, atc.TaskOutputConfig{Name: artifactsOutDir})
	assert.Contains(t, config5.Plan[1].TaskConfig.Outputs, atc.TaskOutputConfig{Name: artifactsOutDirOnFailure})
	assert.False(t, containsPut(artifactsName, config5))

	config6, _ := renderedPipeline.Jobs.Lookup(dockerCompose3)
	assert.Contains(t, config6.Plan[2].TaskConfig.Outputs, atc.TaskOutputConfig{Name: artifactsOutDir})
	assert.Contains(t, config6.Plan[2].TaskConfig.Outputs, atc.TaskOutputConfig{Name: artifactsOutDirOnFailure})
	assert.True(t, containsPut(artifactsName, config6))
}

func TestRendersPipelineFailureOutputIsCorrect(t *testing.T) {
	man := manifest.Manifest{}

	name := "name"

	team := "myTeam"
	pipeline := "myPipeline"

	man.Team = team
	man.Pipeline = pipeline
	man.Tasks = []manifest.Task{
		manifest.Run{
			Name:                   name,
			Script:                 "./build.sh",
			SaveArtifactsOnFailure: []string{"test-reports"},
		},
	}

	renderedPipeline := testPipeline().Render(man)
	config, _ := renderedPipeline.Jobs.Lookup(name)
	assert.Contains(t, config.Plan[1].TaskConfig.Outputs, atc.TaskOutputConfig{Name: artifactsOutDirOnFailure})

	failurePlan := (config.Failure.InParallel.Steps)[0]

	assert.Equal(t, artifactsOnFailureName, failurePlan.Put)
	assert.Equal(t, artifactsOutDirOnFailure, failurePlan.Params["folder"])
	assert.Equal(t, "git/.git/ref", failurePlan.Params["version_file"])
	assert.Equal(t, "failure", failurePlan.Params["postfix"])
}

func TestRendersPipelineFailureOutputHasResourceDef(t *testing.T) {
	man := manifest.Manifest{}

	name := "name"

	team := "myTeam"
	pipeline := "myPipeline"

	man.Team = team
	man.Pipeline = pipeline
	man.Tasks = []manifest.Task{
		manifest.Run{
			Name:                   name,
			Script:                 "./build.sh",
			SaveArtifactsOnFailure: []string{"test-reports"},
		},
	}

	renderedPipeline := testPipeline().Render(man)
	_, found := renderedPipeline.ResourceTypes.Lookup(artifactsResourceName)
	assert.True(t, found)
}

func TestRendersPipelineWithOutputFolderAndFileCopyIfSaveArtifactInMonoRepo(t *testing.T) {
	// Without any save artifact there should not be a copy and a output
	basePath := "apps/subapp1"
	runTask := manifest.Run{
		Script:        "./build.sh",
		SaveArtifacts: []string{"build/lib"},
	}
	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.Git{
				BasePath: basePath,
			},
		},
		Tasks: []manifest.Task{
			runTask,
		},
	}

	renderedPipeline := testPipeline().Render(man)
	assert.Len(t, renderedPipeline.Jobs[0].Plan[1].TaskConfig.Outputs, 1) // Plan[0] is always the git get, Plan[1] is the task
	expectedRunScript := runScriptArgs(runTask, man, true)
	assert.Equal(t, expectedRunScript, renderedPipeline.Jobs[0].Plan[1].TaskConfig.Run.Args)
}

func TestRendersPipelineWithCorrectResourceIfOverridingArtifactoryConfig(t *testing.T) {
	secondTaskName := "DoSomethingWithArtifact"
	man := manifest.Manifest{
		Team:     "team",
		Pipeline: "pipeline",
		ArtifactConfig: manifest.ArtifactConfig{
			Bucket:  "((override.Bucket))",
			JSONKey: "((override.JSONKey))",
		},
		Tasks: []manifest.Task{
			manifest.Run{
				Script:        "./build.sh",
				SaveArtifacts: []string{"build/lib/artifact.jar"},
			},
			manifest.Run{
				Name:             secondTaskName,
				Script:           "./something.sh",
				RestoreArtifacts: true,
			},
		},
	}

	renderedPipeline := testPipeline().Render(man)
	assert.Len(t, renderedPipeline.Jobs[0].Plan, 3)
	assert.Equal(t, artifactsName, renderedPipeline.Jobs[0].Plan[2].Put)
	assert.Equal(t, artifactsOutDir, renderedPipeline.Jobs[0].Plan[2].Params["folder"])
	assert.Equal(t, gitDir+"/.git/ref", renderedPipeline.Jobs[0].Plan[2].Params["version_file"])

	resourceType, _ := renderedPipeline.ResourceTypes.Lookup(artifactsResourceName)
	assert.NotNil(t, resourceType)
	assert.Equal(t, config.DockerRegistry+"gcp-resource", resourceType.Source["repository"])
	assert.NotEmpty(t, resourceType.Source["tag"])

	resource, _ := renderedPipeline.Resources.Lookup(artifactsName)
	assert.NotNil(t, resource)
	assert.Equal(t, man.ArtifactConfig.Bucket, resource.Source["bucket"])
	assert.Equal(t, man.ArtifactConfig.JSONKey, resource.Source["json_key"])

	config, found := renderedPipeline.Jobs.Lookup(secondTaskName)
	assert.True(t, found)
	assert.Equal(t, restoreArtifactTask(man), config.Plan[1])
	assert.Equal(t, man.ArtifactConfig.JSONKey, config.Plan[1].TaskConfig.Params["JSON_KEY"])
	assert.Equal(t, man.ArtifactConfig.Bucket, config.Plan[1].TaskConfig.Params["BUCKET"])
}

func TestRendersPipelineWithDeployArtifacts(t *testing.T) {
	basePath := "apps/subapp1"
	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.Git{
				BasePath: basePath,
			},
		},
		Tasks: []manifest.Task{
			manifest.Run{
				SaveArtifacts: []string{"path/to/artifact"},
			},
			manifest.DeployCF{
				DeployArtifact: "build/lib/artifact.jar",
			},
		},
	}

	renderedPipeline := testPipeline().Render(man)

	assert.Len(t, renderedPipeline.Jobs, 2)
	assert.Len(t, renderedPipeline.Jobs[1].Plan, 5)

	resourceType, _ := renderedPipeline.ResourceTypes.Lookup(artifactsResourceName)
	assert.NotNil(t, resourceType)
	assert.Equal(t, config.DockerRegistry+"gcp-resource", resourceType.Source["repository"])
	assert.NotEmpty(t, resourceType.Source["tag"])

	resource, _ := renderedPipeline.Resources.Lookup(artifactsName)
	assert.NotNil(t, resource)
	assert.Equal(t, config.ArtifactsBucket, resource.Source["bucket"])
	assert.Equal(t, path.Join(man.Team, man.Pipeline), resource.Source["folder"])
	assert.Equal(t, config.ArtifactsJSONKey, resource.Source["json_key"])
}

func TestRenderPipelineWithSaveAndDeploy(t *testing.T) {
	repoName := "yolo"
	gitURI := fmt.Sprintf("git@github.com:springernature/%s.git", repoName)
	basePath := "apps/subapp1"
	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.Git{
				URI:      gitURI,
				BasePath: basePath,
			},
		},
	}

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

	assert.Len(t, renderedPipeline.Jobs, 2)
	assert.Len(t, renderedPipeline.Jobs[0].Plan, 3)
	assert.Len(t, renderedPipeline.Jobs[1].Plan, 5)

	// order of the plans is important
	assert.Equal(t, restoreArtifactTask(man), renderedPipeline.Jobs[1].Plan[1])
	assert.Equal(t, "cf halfpipe-push", renderedPipeline.Jobs[1].Plan[2].Put)

	expectedAppPath := fmt.Sprintf("%s/%s/%s", artifactsInDir, basePath, deployArtifactPath)
	assert.Equal(t, expectedAppPath, renderedPipeline.Jobs[1].Plan[2].Params["appPath"])
}

func TestRenderPipelineWithSaveAndDeployInSingleAppRepo(t *testing.T) {
	man := manifest.Manifest{}
	man.Team = "team"
	man.Pipeline = "pipeline"
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
	assert.Equal(t, restoreArtifactTask(man), renderedPipeline.Jobs[1].Plan[1])
	assert.Equal(t, "cf halfpipe-push", renderedPipeline.Jobs[1].Plan[2].Put)
	assert.Equal(t, path.Join(artifactsInDir, "/build/lib/artifact.jar"), renderedPipeline.Jobs[1].Plan[2].Params["appPath"])
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

	assert.Equal(t, restoreArtifactTask(man), config.Jobs[0].Plan[1])
	assert.Equal(t, "artifacts-out", config.Jobs[0].Plan[2].TaskConfig.Outputs[0].Name)
}

func TestRenderGetHasTheSameConfigOptionsInTheRestoreAsInTheResourceConfig(t *testing.T) {
	restoreArtifactTaskName := "Restore Artifact"

	man := manifest.Manifest{
		Team:     "Proteus_PWG_performance_metrics",
		Pipeline: "proteus-services",

		Tasks: []manifest.Task{
			manifest.Run{
				Name: "SaveArtifactName",
				SaveArtifacts: []string{
					".",
				},
			},
			manifest.Run{
				Name:             restoreArtifactTaskName,
				RestoreArtifacts: true,
			},
		},
	}

	config := testPipeline().Render(man)

	artifactConfig, foundResourceConfig := config.Resources.Lookup(artifactsName)
	assert.True(t, foundResourceConfig)

	restoreJob, foundJobConfig := config.Jobs.Lookup(restoreArtifactTaskName)
	assert.True(t, foundJobConfig)

	assert.Equal(t, restoreArtifactTask(man), restoreJob.Plan[1])
	restoreArtifactTask := restoreJob.Plan[1]

	assert.Equal(t, artifactConfig.Source["bucket"], restoreArtifactTask.TaskConfig.Params["BUCKET"])
	assert.Equal(t, artifactConfig.Source["json_key"], restoreArtifactTask.TaskConfig.Params["JSON_KEY"])
	assert.Equal(t, artifactConfig.Source["folder"], restoreArtifactTask.TaskConfig.Params["FOLDER"])
}

func TestRenderRunWithSaveArtifactsAndSaveArtifactsOnFailure(t *testing.T) {
	jarOutputFolder := "build/jars"
	testReportsFolder := "build/test-reports"

	team := "team"
	pipeline := "pipeline"
	man := manifest.Manifest{
		Team:     team,
		Pipeline: pipeline,
		Tasks: []manifest.Task{
			manifest.Run{
				Script: "\\make ; ls -al",
				SaveArtifacts: []string{
					jarOutputFolder,
				},
				SaveArtifactsOnFailure: []string{
					testReportsFolder,
				},
			},
		},
	}

	config := testPipeline().Render(man)

	_, found := config.Resources.Lookup(artifactsName)
	assert.True(t, found)

	_, foundOnFailure := config.Resources.Lookup(artifactsOnFailureName)
	assert.True(t, foundOnFailure)

	assert.Equal(t, artifactsOutDir, config.Jobs[0].Plan[1].TaskConfig.Outputs[0].Name)
	assert.Equal(t, artifactsOutDirOnFailure, config.Jobs[0].Plan[1].TaskConfig.Outputs[1].Name)

	assert.Equal(t, atc.PlanConfig{
		Put: artifactsName,
		Params: atc.Params{
			"folder":       artifactsOutDir,
			"version_file": "git/.git/ref",
		},
	}, config.Jobs[0].Plan[2])

	failureInParallel := config.Jobs[0].Failure.InParallel.Steps
	assert.Equal(t, saveArtifactOnFailurePlan(), (failureInParallel)[0])

	assert.Contains(t, strings.Join(config.Jobs[0].Plan[1].TaskConfig.Run.Args, "\n"), fmt.Sprintf("copyArtifact %s", jarOutputFolder))
	assert.Contains(t, strings.Join(config.Jobs[0].Plan[1].TaskConfig.Run.Args, "\n"), fmt.Sprintf("copyArtifact %s", testReportsFolder))
}

func TestRenderRunWithCorrectResources(t *testing.T) {

	t.Run("It has no artifacts", func(t *testing.T) {
		team := "team"
		pipeline := "pipeline"
		man := manifest.Manifest{
			Team:     team,
			Pipeline: pipeline,
			Tasks: []manifest.Task{
				manifest.Run{
					Script: "\\make ; ls -al",
				},
			},
		}

		config := testPipeline().Render(man)

		_, found := config.Resources.Lookup(artifactsName)
		assert.False(t, found)

		_, foundOnFailure := config.Resources.Lookup(artifactsOnFailureName)
		assert.False(t, foundOnFailure)

	})

	t.Run("It has restore artifacts", func(t *testing.T) {
		team := "team"
		pipeline := "pipeline"
		man := manifest.Manifest{
			Team:     team,
			Pipeline: pipeline,
			Tasks: []manifest.Task{
				manifest.Run{
					SaveArtifacts: []string{"path/to/artifact"},
				},
				manifest.Run{
					RestoreArtifacts: true,
					Script:           "\\make ; ls -al",
				},
			},
		}

		config := testPipeline().Render(man)

		_, found := config.Resources.Lookup(artifactsName)
		assert.True(t, found)

		_, foundOnFailure := config.Resources.Lookup(artifactsOnFailureName)
		assert.False(t, foundOnFailure)
	})

	t.Run("It has safe artifacts", func(t *testing.T) {
		team := "team"
		pipeline := "pipeline"
		man := manifest.Manifest{
			Team:     team,
			Pipeline: pipeline,
			Tasks: []manifest.Task{
				manifest.Run{
					Script: "\\make ; ls -al",
					SaveArtifacts: []string{
						"a",
					},
				},
			},
		}

		config := testPipeline().Render(man)

		_, found := config.Resources.Lookup(artifactsName)
		assert.True(t, found)

		_, foundOnFailure := config.Resources.Lookup(artifactsOnFailureName)
		assert.False(t, foundOnFailure)
	})

	t.Run("It has safe artifacts on failure", func(t *testing.T) {
		team := "team"
		pipeline := "pipeline"
		man := manifest.Manifest{
			Team:     team,
			Pipeline: pipeline,
			Tasks: []manifest.Task{
				manifest.Run{
					Script: "\\make ; ls -al",
					SaveArtifactsOnFailure: []string{
						"a",
					},
				},
			},
		}

		config := testPipeline().Render(man)

		_, found := config.Resources.Lookup(artifactsName)
		assert.False(t, found)

		_, foundOnFailure := config.Resources.Lookup(artifactsOnFailureName)
		assert.True(t, foundOnFailure)
	})

	t.Run("It has save artifacts on failure and normal artifact save", func(t *testing.T) {
		team := "team"
		pipeline := "pipeline"
		man := manifest.Manifest{
			Team:     team,
			Pipeline: pipeline,
			Tasks: []manifest.Task{
				manifest.Run{
					Script: "\\make ; ls -al",
					SaveArtifactsOnFailure: []string{
						"a",
					},
					SaveArtifacts: []string{
						"b",
					},
				},
			},
		}

		config := testPipeline().Render(man)

		_, found := config.Resources.Lookup(artifactsName)
		assert.True(t, found)

		_, foundOnFailure := config.Resources.Lookup(artifactsOnFailureName)
		assert.True(t, foundOnFailure)
	})

	t.Run("It has save artifacts, save artifacts on failure and versioned resources", func(t *testing.T) {
		team := "team"
		pipeline := "pipeline"
		man := manifest.Manifest{
			Team:     team,
			Pipeline: pipeline,
			FeatureToggles: []string{
				manifest.FeatureUpdatePipeline,
			},
			Tasks: []manifest.Task{
				manifest.Run{
					Script: "\\make ; ls -al",
					SaveArtifactsOnFailure: []string{
						"a",
					},
					SaveArtifacts: []string{
						"b",
					},
				},
			},
		}
		config := testPipeline().Render(man)

		_, found := config.Resources.Lookup(artifactsName)
		assert.True(t, found)

		_, foundOnFailure := config.Resources.Lookup(artifactsOnFailureName)
		assert.True(t, foundOnFailure)

		_, foundVersion := config.Resources.Lookup(versionName)
		assert.True(t, foundVersion)
	})

}

func TestFullPathToArtifactDir(t *testing.T) {
	t.Run("when in root", func(t *testing.T) {
		t.Run("when saving artifact in current dir", func(t *testing.T) {
			assert.Equal(t, "../artifacts-out", fullPathToArtifactsDir(gitDir, "", artifactsOutDir, "file"))
			assert.Equal(t, "../artifacts-out/subFolder", fullPathToArtifactsDir(gitDir, "", artifactsOutDir, "subFolder/file"))
			assert.Equal(t, "../artifacts-out/subFolder1/subFolder2", fullPathToArtifactsDir(gitDir, "", artifactsOutDir, "subFolder1/subFolder2/file"))
		})
	})

	t.Run("when in a sub app", func(t *testing.T) {
		t.Run("when saving artifact in current dir", func(t *testing.T) {
			assert.Equal(t, "../../artifacts-out/subApp", fullPathToArtifactsDir(gitDir, "subApp", artifactsOutDir, "file"))
			assert.Equal(t, "../../artifacts-out/subApp/subFolder", fullPathToArtifactsDir(gitDir, "subApp", artifactsOutDir, "subFolder/file"))
			assert.Equal(t, "../../artifacts-out/subApp/subFolder1/subFolder2", fullPathToArtifactsDir(gitDir, "subApp", artifactsOutDir, "subFolder1/subFolder2/file"))
		})

		t.Run("when saving artifact from parent dir", func(t *testing.T) {
			assert.Equal(t, "../../artifacts-out", fullPathToArtifactsDir(gitDir, "subApp", artifactsOutDir, "../file"))
			assert.Equal(t, "../../artifacts-out/parentFolder", fullPathToArtifactsDir(gitDir, "subApp", artifactsOutDir, "../parentFolder/file"))
			assert.Equal(t, "../../artifacts-out/parentFolder/subFolder1/subFolder2", fullPathToArtifactsDir(gitDir, "subApp", artifactsOutDir, "../parentFolder/subFolder1/subFolder2/file"))
			assert.Equal(t, "../../artifacts-out/parentFolder/subFolder1/subFolder2", fullPathToArtifactsDir(gitDir, "subApp", artifactsOutDir, "../parentFolder/subFolder1/subFolder2/subFolder3/../file"))
		})
	})

}
