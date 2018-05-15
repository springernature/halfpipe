package pipeline

import (
	"testing"

	"github.com/concourse/atc"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestRenderRunTask(t *testing.T) {
	man := manifest.Manifest{}
	man.Repo.URI = "git@github.com:/springernature/foo.git"
	man.Tasks = []manifest.Task{
		manifest.Run{
			Script: "./yolo.sh",
			Docker: manifest.Docker{
				Image:    "imagename:TAG",
				Username: "",
				Password: "",
			},
			Vars: manifest.Vars{
				"VAR1": "Value1",
				"VAR2": "Value2",
			},
		},
	}

	expected := atc.JobConfig{
		Name:   "run yolo.sh",
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{Get: gitDir, Trigger: true},
			atc.PlanConfig{Task: "run", Privileged: false, TaskConfig: &atc.TaskConfig{
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
					Dir:  gitDir,
					Args: runScriptArgs("./yolo.sh", true, "", nil, ".git/ref"),
				},
				Inputs: []atc.TaskInputConfig{
					{Name: gitDir},
				},
			}},
		}}

	assert.Equal(t, expected, testPipeline().Render(man).Jobs[0])
}
func TestRenderRunTaskWithPrivateRepo(t *testing.T) {
	man := manifest.Manifest{}
	man.Repo.URI = "git@github.com:/springernature/foo.git"
	man.Tasks = []manifest.Task{
		manifest.Run{
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
		},
	}

	expected := atc.JobConfig{
		Name:   "run yolo.sh",
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{Get: gitDir, Trigger: true},
			atc.PlanConfig{Task: "run", Privileged: false, TaskConfig: &atc.TaskConfig{
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
						"username":   "user",
						"password":   "pass",
					},
				},
				Run: atc.TaskRunConfig{
					Path: "/bin/sh",
					Dir:  gitDir,
					Args: runScriptArgs("./yolo.sh", true, "", nil, ".git/ref"),
				},
				Inputs: []atc.TaskInputConfig{
					{Name: gitDir},
				},
			}},
		}}

	assert.Equal(t, expected, testPipeline().Render(man).Jobs[0])
}

func TestRenderRunTaskFromHalfpipeNotInRoot(t *testing.T) {
	man := manifest.Manifest{}
	basePath := "subapp"
	man.Repo.URI = "git@github.com:/springernature/foo.git"
	man.Repo.BasePath = basePath

	man.Tasks = []manifest.Task{
		manifest.Run{
			Script: "./yolo.sh",
			Docker: manifest.Docker{
				Image: "imagename:TAG",
			},
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
			atc.PlanConfig{Get: gitDir, Trigger: true},
			atc.PlanConfig{Task: "run", Privileged: false, TaskConfig: &atc.TaskConfig{
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
					Dir:  gitDir + "/" + basePath,
					Args: runScriptArgs("./yolo.sh", true, "", nil, "../.git/ref"),
				},
				Inputs: []atc.TaskInputConfig{
					{Name: gitDir},
				},
			}},
		}}

	assert.Equal(t, expected, testPipeline().Render(man).Jobs[0])
}

func TestRunScriptArgs(t *testing.T) {
	withNoArtifacts := runScriptArgs("./build.sh", true, "", nil, ".git/ref")
	expected := []string{"-c", "which bash > /dev/null\nif [ $? != 0 ]; then\n  echo \"Bash is not present in the docker image\"\n  echo \"If you script, or any of the script your script is calling depends on bash you will get a strange error message like:\"\n  echo \"sh: yourscript.sh: command not found\"\n  echo \"To fix, make sure your docker image contains bash!\"\n  echo \"\"\n  echo \"\"\nfi\nset -e\nexport GIT_REVISION=`cat .git/ref`\n./build.sh"}

	assert.Equal(t, expected, withNoArtifacts)
}

func TestRunScriptArgsWhenInMonoRepo(t *testing.T) {
	withNoArtifacts := runScriptArgs("./build.sh", true, "", nil, ".git/ref")
	expected := []string{"-c", "which bash > /dev/null\nif [ $? != 0 ]; then\n  echo \"Bash is not present in the docker image\"\n  echo \"If you script, or any of the script your script is calling depends on bash you will get a strange error message like:\"\n  echo \"sh: yourscript.sh: command not found\"\n  echo \"To fix, make sure your docker image contains bash!\"\n  echo \"\"\n  echo \"\"\nfi\nset -e\nexport GIT_REVISION=`cat .git/ref`\n./build.sh"}

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
		args := runScriptArgs(initial, true, "", nil, ".git/ref")
		expected := []string{"-c", "which bash > /dev/null\nif [ $? != 0 ]; then\n  echo \"Bash is not present in the docker image\"\n  echo \"If you script, or any of the script your script is calling depends on bash you will get a strange error message like:\"\n  echo \"sh: yourscript.sh: command not found\"\n  echo \"To fix, make sure your docker image contains bash!\"\n  echo \"\"\n  echo \"\"\nfi\nset -e\nexport GIT_REVISION=`cat .git/ref`\n" + updated}

		assert.Equal(t, expected, args, initial)
	}
}
