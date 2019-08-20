package pipeline

import (
	"fmt"
	"path"
	"strings"
	"testing"

	"github.com/concourse/concourse/atc"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestRenderDockerPushTask(t *testing.T) {
	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.GitTrigger{},
		},
	}

	username := "halfpipe"
	password := "secret"
	repo := "halfpipe/halfpipe-cli"
	man.Tasks = []manifest.Task{
		manifest.DockerPush{
			Name:     "docker-push",
			Username: username,
			Password: password,
			Image:    repo,
			Vars: manifest.Vars{
				"A": "a",
				"B": "b",
			},
			DockerfilePath: "Dockerfile",
		},
	}

	expectedResource := atc.ResourceConfig{
		Name: "halfpipe-cli",
		Type: "docker-image",
		Source: atc.Source{
			"username":   username,
			"password":   password,
			"repository": repo,
		},
		CheckEvery: longResourceCheckInterval,
	}

	expectedJobConfig := atc.JobConfig{
		Name:   "docker-push",
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{InParallel: &atc.InParallelConfig{Steps: atc.PlanSequence{atc.PlanConfig{Get: gitName, Trigger: true, Attempts: gitGetAttempts}}}},
			atc.PlanConfig{
				Attempts: 1,
				Put:      "halfpipe-cli",
				Params: atc.Params{
					"build":      gitDir,
					"dockerfile": path.Join(gitDir, "Dockerfile"),
					"build_args": map[string]interface{}{
						"A": "a",
						"B": "b",
					},
					"tag_as_latest": true,
				},
			},
		},
	}

	// First resource will always be the git resource.
	assert.Equal(t, expectedResource, testPipeline().Render(man).Resources[1])
	assert.Equal(t, expectedJobConfig, testPipeline().Render(man).Jobs[0])
}

func TestRenderDockerPushTaskNotInRoot(t *testing.T) {
	basePath := "subapp/sub2"

	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.GitTrigger{
				BasePath: basePath,
			},
		},
	}

	username := "halfpipe"
	password := "secret"
	repo := "halfpipe/halfpipe-cli"
	man.Tasks = []manifest.Task{
		manifest.DockerPush{
			Name:           "docker-push",
			Username:       username,
			Password:       password,
			Image:          repo,
			DockerfilePath: "dockerfile/Dockerfile",
		},
	}

	expectedResource := atc.ResourceConfig{
		Name: "halfpipe-cli",
		Type: "docker-image",
		Source: atc.Source{
			"username":   username,
			"password":   password,
			"repository": repo,
		},
		CheckEvery: longResourceCheckInterval,
	}

	expectedJobConfig := atc.JobConfig{
		Name:   "docker-push",
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{InParallel: &atc.InParallelConfig{Steps: atc.PlanSequence{atc.PlanConfig{Get: gitName, Trigger: true, Attempts: gitGetAttempts}}}},
			atc.PlanConfig{
				Attempts: 1,
				Put:      "halfpipe-cli",
				Params: atc.Params{
					"build":         gitDir + "/" + basePath,
					"dockerfile":    path.Join(gitDir, basePath, man.Tasks[0].(manifest.DockerPush).DockerfilePath),
					"tag_as_latest": true,
				}},
		},
	}

	// First resource will always be the git resource.
	assert.Equal(t, expectedResource, testPipeline().Render(man).Resources[1])
	assert.Equal(t, expectedJobConfig, testPipeline().Render(man).Jobs[0])
}

func TestRenderDockerPushWithVersioning(t *testing.T) {
	basePath := "subapp/sub2"
	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.GitTrigger{
				URI:      "git@github.com:/springernature/foo.git",
				BasePath: basePath,
			},
		},
		FeatureToggles: manifest.FeatureToggles{
			manifest.FeatureUpdatePipeline,
		},
	}

	username := "halfpipe"
	password := "secret"
	repo := "halfpipe/halfpipe-cli"
	man.Tasks = []manifest.Task{
		manifest.Update{},
		manifest.DockerPush{
			Name:           "docker-push",
			Username:       username,
			Password:       password,
			Image:          repo,
			DockerfilePath: "Dockerfile",
		},
	}

	expectedResource := atc.ResourceConfig{
		Name: "halfpipe-cli",
		Type: "docker-image",
		Source: atc.Source{
			"username":   username,
			"password":   password,
			"repository": repo,
		},
		CheckEvery: longResourceCheckInterval,
	}

	expectedJobConfig := atc.JobConfig{
		Name:   "docker-push",
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{InParallel: &atc.InParallelConfig{Steps: atc.PlanSequence{
				atc.PlanConfig{Get: gitName, Passed: []string{updateJobName}, Attempts: gitGetAttempts},
				atc.PlanConfig{Get: versionName, Passed: []string{updateJobName}, Trigger: true, Attempts: versionGetAttempts}},
			}},
			atc.PlanConfig{
				Attempts: 1,
				Put:      "halfpipe-cli",
				Params: atc.Params{
					"tag_file":      "version/number",
					"build":         gitDir + "/" + basePath,
					"dockerfile":    path.Join(gitDir, basePath, man.Tasks[1].(manifest.DockerPush).DockerfilePath),
					"tag_as_latest": true,
				}},
		},
	}

	// First resource will always be the git resource.
	assert.Equal(t, expectedResource, testPipeline().Render(man).Resources[2])
	assert.Equal(t, expectedJobConfig, testPipeline().Render(man).Jobs[1])
}

func TestRenderDockerPushWithVersioningAndRestoreArtifact(t *testing.T) {
	basePath := "subapp/sub2"
	buildPath := "build/path"

	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.GitTrigger{
				URI:      "git@github.com:/springernature/foo.git",
				BasePath: basePath,
			},
		},
		FeatureToggles: manifest.FeatureToggles{
			manifest.FeatureUpdatePipeline,
		},
	}

	username := "halfpipe"
	password := "secret"
	repo := "halfpipe/halfpipe-cli"
	dockerPush := manifest.DockerPush{
		Name:             "docker-push",
		Username:         username,
		Password:         password,
		Image:            repo,
		RestoreArtifacts: true,
		DockerfilePath:   "Dockerfile",
		BuildPath:        buildPath,
	}
	man.Tasks = []manifest.Task{
		manifest.Update{},
		dockerPush,
	}

	expectedResource := atc.ResourceConfig{
		Name: dockerPushResourceName(dockerPush),
		Type: "docker-image",
		Source: atc.Source{
			"username":   username,
			"password":   password,
			"repository": repo,
		},
		CheckEvery: longResourceCheckInterval,
	}

	jobName := "docker-push"
	expectedJobConfig := atc.JobConfig{
		Name:   jobName,
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{InParallel: &atc.InParallelConfig{Steps: atc.PlanSequence{
				atc.PlanConfig{Get: gitName, Passed: []string{updateJobName}, Attempts: gitGetAttempts},
				atc.PlanConfig{Get: versionName, Passed: []string{updateJobName}, Trigger: true, Attempts: versionGetAttempts}},
			}},
			restoreArtifactTask(man),
			atc.PlanConfig{
				Task: "Copying git repo and artifacts to a temporary build dir",
				TaskConfig: &atc.TaskConfig{
					Platform: "linux",
					ImageResource: &atc.ImageResource{
						Type: "docker-image",
						Source: atc.Source{
							"repository": "alpine",
						},
					},
					Run: atc.TaskRunConfig{
						Path: "/bin/sh",
						Args: []string{"-c", strings.Join([]string{
							fmt.Sprintf("cp -r %s/. %s", gitDir, dockerBuildTmpDir),
							fmt.Sprintf("cp -r %s/. %s", artifactsInDir, dockerBuildTmpDir),
						}, "\n")},
					},
					Inputs: []atc.TaskInputConfig{
						{Name: gitName},
						{Name: artifactsName},
					},
					Outputs: []atc.TaskOutputConfig{
						{Name: dockerBuildTmpDir},
					},
				},
			},
			atc.PlanConfig{
				Attempts: 1,
				Put:      dockerPushResourceName(dockerPush),
				Params: atc.Params{
					"tag_file":      "version/number",
					"build":         dockerBuildTmpDir + "/" + basePath + "/" + buildPath,
					"dockerfile":    path.Join(dockerBuildTmpDir, basePath, man.Tasks[1].(manifest.DockerPush).DockerfilePath),
					"tag_as_latest": true,
				}},
		},
	}

	// First resource will always be the git resource.
	dockerResource, found := testPipeline().Render(man).Resources.Lookup(dockerPushResourceName(dockerPush))
	assert.True(t, found)
	assert.Equal(t, expectedResource, dockerResource)

	config, foundJob := testPipeline().Render(man).Jobs.Lookup(jobName)
	assert.True(t, foundJob)
	assert.Equal(t, expectedJobConfig, config)
}
