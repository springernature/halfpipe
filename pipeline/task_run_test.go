package pipeline

import (
	"fmt"
	"testing"

	"github.com/concourse/concourse/atc"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestRenderRunTask(t *testing.T) {
	runTask := manifest.Run{
		Retries: 2,
		Name:    "run yolo.sh",
		Script:  "./yolo.sh",
		Docker: manifest.Docker{
			Image:    "imagename:TAG",
			Username: "",
			Password: "",
		},
		Privileged: true,
		Vars: manifest.Vars{
			"VAR1": "Value1",
			"VAR2": "Value2",
		},
	}

	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.GitTrigger{},
		},
		Tasks: []manifest.Task{
			runTask,
		},
	}

	expected := atc.JobConfig{
		Name: "run yolo.sh",
		BuildLogRetention: &(atc.BuildLogRetention{
			Builds:                 0,
			MinimumSucceededBuilds: 1,
		}),
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{InParallel: &atc.InParallelConfig{FailFast: true, Steps: atc.PlanSequence{atc.PlanConfig{Get: gitName, Trigger: true, Attempts: gitGetAttempts}}}},
			atc.PlanConfig{
				Attempts:   3,
				Task:       "run yolo.sh",
				Privileged: true,
				TaskConfig: &atc.TaskConfig{
					Platform: "linux",
					Params: map[string]string{
						"VAR1": "Value1",
						"VAR2": "Value2",
					},
					ImageResource: &atc.ImageResource{
						Type: "registry-image",
						Source: atc.Source{
							"repository": "imagename",
							"tag":        "TAG",
						},
					},
					Run: atc.TaskRunConfig{
						Path: "/bin/sh",
						Dir:  gitDir,
						Args: runScriptArgs(runTask, manifest.Manifest{}, true, ""),
					},
					Inputs: []atc.TaskInputConfig{
						{Name: gitName},
					},
					Caches: config.CacheDirs,
				}},
		}}

	assert.Equal(t, expected, testPipeline().Render(man).Jobs[0])
}

func TestRenderRunTaskWithPrivateRepo(t *testing.T) {
	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.GitTrigger{},
		},
	}
	runTask := manifest.Run{
		Name:   "run yolo.sh",
		Script: "./yolo.sh",
		Docker: manifest.Docker{
			Image:    "imagename:TAG",
			Username: "user",
			Password: "pass",
		},
		Vars: map[string]string{
			"VAR1": "Value1",
			"VAR2": "Value2",
		},
	}
	man.Tasks = []manifest.Task{
		runTask,
	}

	expected := atc.JobConfig{
		Name: "run yolo.sh",
		BuildLogRetention: &(atc.BuildLogRetention{
			Builds:                 0,
			MinimumSucceededBuilds: 1,
		}),
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{InParallel: &atc.InParallelConfig{FailFast: true, Steps: atc.PlanSequence{atc.PlanConfig{Get: gitName, Trigger: true, Attempts: gitGetAttempts}}}},
			atc.PlanConfig{
				Attempts:   1,
				Task:       "run yolo.sh",
				Privileged: false,
				TaskConfig: &atc.TaskConfig{
					Platform: "linux",
					Params: map[string]string{
						"VAR1": "Value1",
						"VAR2": "Value2",
					},
					ImageResource: &atc.ImageResource{
						Type: "registry-image",
						Source: atc.Source{
							"repository": "imagename",
							"tag":        "TAG",
							"username":   "user",
							"password":   "pass",
						},
					},
					Run: atc.TaskRunConfig{
						Path: "/bin/sh",
						Dir:  gitDir,
						Args: runScriptArgs(runTask, manifest.Manifest{}, true, ""),
					},
					Inputs: []atc.TaskInputConfig{
						{Name: gitName},
					},
					Caches: config.CacheDirs,
				}},
		}}

	assert.Equal(t, expected, testPipeline().Render(man).Jobs[0])
}

func TestRenderRunTaskFromHalfpipeNotInRoot(t *testing.T) {
	basePath := "subapp"

	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.GitTrigger{
				BasePath: basePath,
			},
		},
	}

	runTask := manifest.Run{
		Name:   "run yolo.sh",
		Script: "./yolo.sh",
		Docker: manifest.Docker{
			Image: "imagename:TAG",
		},
		Vars: map[string]string{
			"VAR1": "Value1",
			"VAR2": "Value2",
		},
	}
	man.Tasks = []manifest.Task{
		runTask,
	}

	expected := atc.JobConfig{
		Name:   "run yolo.sh",
		Serial: true,
		BuildLogRetention: &(atc.BuildLogRetention{
			Builds:                 0,
			MinimumSucceededBuilds: 1,
		}),
		Plan: atc.PlanSequence{
			atc.PlanConfig{InParallel: &atc.InParallelConfig{FailFast: true, Steps: atc.PlanSequence{atc.PlanConfig{Get: gitName, Trigger: true, Attempts: gitGetAttempts}}}},
			atc.PlanConfig{
				Attempts:   1,
				Task:       "run yolo.sh",
				Privileged: false,
				TaskConfig: &atc.TaskConfig{
					Platform: "linux",
					Params: map[string]string{
						"VAR1": "Value1",
						"VAR2": "Value2",
					},
					ImageResource: &atc.ImageResource{
						Type: "registry-image",
						Source: atc.Source{
							"repository": "imagename",
							"tag":        "TAG",
						},
					},
					Run: atc.TaskRunConfig{
						Path: "/bin/sh",
						Dir:  gitDir + "/" + basePath,
						Args: runScriptArgs(runTask, man, true, man.Triggers.GetGitTrigger().BasePath),
					},
					Inputs: []atc.TaskInputConfig{
						{Name: gitName},
					},
					Caches: config.CacheDirs,
				}},
		}}

	assert.Equal(t, expected, testPipeline().Render(man).Jobs[0])
}

func TestRunScriptArgs(t *testing.T) {
	withNoArtifacts := runScriptArgs(manifest.Run{Script: "./build.sh"}, manifest.Manifest{}, true, "")
	expected := []string{"-c", fmt.Sprintf("%s\n%s\n%s", warningMissingBash, warningAlpineImage, "export GIT_REVISION=`cat .git/ref`\n\n./build.sh\nEXIT_STATUS=$?\nif [ $EXIT_STATUS != 0 ] ; then\n  exit 1\nfi\n")}

	assert.Equal(t, expected, withNoArtifacts)
}

func TestRunScriptArgsWhenInMonoRepo(t *testing.T) {
	withNoArtifacts := runScriptArgs(manifest.Run{Script: "./build.sh"}, manifest.Manifest{}, true, "")
	expected := []string{"-c", fmt.Sprintf("%s\n%s\n%s", warningMissingBash, warningAlpineImage, "export GIT_REVISION=`cat .git/ref`\n\n./build.sh\nEXIT_STATUS=$?\nif [ $EXIT_STATUS != 0 ] ; then\n  exit 1\nfi\n")}

	assert.Equal(t, expected, withNoArtifacts)
}

func TestRunScriptPath(t *testing.T) {
	tests := map[string]string{
		"./build.sh":          "./build.sh",
		"/build.sh":           "/build.sh",
		"build.sh":            "./build.sh",
		"../build.sh":         "./../build.sh",
		"./build.sh -v --p=1": "./build.sh -v --p=1",
		`\source foo.sh`:      `\source foo.sh`,
	}

	for initial, updated := range tests {
		args := runScriptArgs(manifest.Run{Script: initial}, manifest.Manifest{}, true, "")
		expected := []string{"-c", fmt.Sprintf("%s\n%s\nexport GIT_REVISION=`cat .git/ref`\n\n%s\nEXIT_STATUS=$?\nif [ $EXIT_STATUS != 0 ] ; then\n  exit 1\nfi\n", warningMissingBash, warningAlpineImage, updated)}

		assert.Equal(t, expected, args, initial)
	}
}
