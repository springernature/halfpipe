package linters

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"github.com/pkg/errors"
)

var validDockerCompose = `
version: 3
services:
  app:
    image: appropriate/curl`

func testTaskLinter() taskLinter {
	return taskLinter{
		Fs: afero.Afero{Fs: afero.NewMemMapFs()},
	}
}

func TestAtLeastOneTaskExists(t *testing.T) {
	man := manifest.Manifest{}
	taskLinter := testTaskLinter()

	result := taskLinter.Lint(man)
	assert.Len(t, result.Errors, 1)
	assertMissingField(t, "tasks", result.Errors[0])
}

func TestCallsOutToTheLintersCorrectly(t *testing.T) {
	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.Run{},
			manifest.DeployCF{
				PrePromote: []manifest.Task{
					manifest.Run{},
					manifest.DeployCF{
						PrePromote: []manifest.Task{
							manifest.Run{},
						},
					},
					manifest.DockerPush{},
					manifest.DockerCompose{},
					manifest.ConsumerIntegrationTest{},
					manifest.DeployMLZip{},
					manifest.DeployMLModules{},
				},
			},
			manifest.DockerPush{},
			manifest.DockerCompose{},
			manifest.ConsumerIntegrationTest{},
			manifest.DeployMLZip{},
			manifest.DeployMLModules{},
		},
	}

	calledLintRunTask := false
	calledLintRunTaskNum := 0
	calledLintDeployCFTask := false
	calledLintDeployCFTaskNum := 0
	calledLintDockerPushTask := false
	calledLintDockerPushTaskNum := 0
	calledLintDockerComposeTask := false
	calledLintDockerComposeTaskNum := 0
	calledLintConsumerIntegrationTestTask := false
	calledLintConsumerIntegrationTestTaskNum := 0
	calledLintDeployMLZipTask := false
	calledLintDeployMLZipTaskNum := 0
	calledLintDeployMLModulesTask := false
	calledLintDeployMLModulesTaskNum := 0

	taskLinter := taskLinter{
		Fs: afero.Afero{
			Fs: nil,
		},
		lintRunTask: func(task manifest.Run, taskID string, fs afero.Afero) (errs []error, warnings []error) {
			calledLintRunTask = true
			calledLintRunTaskNum += 1
			return
		},
		lintDeployCFTask: func(task manifest.DeployCF, taskID string, fs afero.Afero) (errs []error, warnings []error) {
			calledLintDeployCFTask = true
			calledLintDeployCFTaskNum += 1
			return
		},
		lintDockerPushTask: func(task manifest.DockerPush, taskID string, fs afero.Afero) (errs []error, warnings []error) {
			calledLintDockerPushTask = true
			calledLintDockerPushTaskNum += 1
			return
		},
		lintDockerComposeTask: func(task manifest.DockerCompose, taskID string, fs afero.Afero) (errs []error, warnings []error) {
			calledLintDockerComposeTask = true
			calledLintDockerComposeTaskNum += 1
			return
		},
		lintConsumerIntegrationTestTask: func(cit manifest.ConsumerIntegrationTest, taskID string, providerHostRequired bool) (errs []error, warnings []error) {
			calledLintConsumerIntegrationTestTask = true
			calledLintConsumerIntegrationTestTaskNum += 1
			return
		},
		lintDeployMLZipTask: func(task manifest.DeployMLZip, taskID string) (errs []error, warnings []error) {
			calledLintDeployMLZipTask = true
			calledLintDeployMLZipTaskNum += 1
			return
		},
		lintDeployMLModulesTask: func(task manifest.DeployMLModules, taskID string) (errs []error, warnings []error) {
			calledLintDeployMLModulesTask = true
			calledLintDeployMLModulesTaskNum += 1
			return
		},
	}

	taskLinter.Lint(man)

	assert.True(t, calledLintRunTask)
	assert.Equal(t, 3, calledLintRunTaskNum)

	assert.True(t, calledLintDeployCFTask)
	assert.Equal(t, 2, calledLintDeployCFTaskNum)

	assert.True(t, calledLintDockerPushTask)
	assert.Equal(t, 2, calledLintDockerPushTaskNum)

	assert.True(t, calledLintDockerComposeTask)
	assert.Equal(t, 2, calledLintDockerComposeTaskNum)

	assert.True(t, calledLintConsumerIntegrationTestTask)
	assert.Equal(t, 2, calledLintConsumerIntegrationTestTaskNum)

	assert.True(t, calledLintDeployMLZipTask)
	assert.Equal(t, 2, calledLintDeployMLZipTaskNum)

	assert.True(t, calledLintDeployMLModulesTask)
	assert.Equal(t, 2, calledLintDeployMLModulesTaskNum)
}

func TestMergesTheErrorsAndWarningsCorrectly(t *testing.T) {
	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.Run{},
			manifest.DeployCF{
				PrePromote: []manifest.Task{
					manifest.Run{},
					manifest.DockerPush{},
				},
			},
			manifest.DeployMLZip{},
			manifest.DeployMLModules{},
		},
	}

	runErr1 := errors.New("runErr1")
	runErr2 := errors.New("runErr2")
	runWarn1 := errors.New("runWarn1")

	deployErr := errors.New("deployErr")

	dockerPushErr := errors.New("dockerPushErr")
	dockerPushWarn := errors.New("dockerPushWarn")

	deployMlZipErr := errors.New("deployMlZipErr")

	deployMlModulesWarn := errors.New("deployMlModulesWarn")
	taskLinter := taskLinter{
		Fs: afero.Afero{
			Fs: nil,
		},
		lintRunTask: func(task manifest.Run, taskID string, fs afero.Afero) (errs []error, warnings []error) {
			return []error{runErr1, runErr2}, []error{runWarn1}
		},
		lintDeployCFTask: func(task manifest.DeployCF, taskID string, fs afero.Afero) (errs []error, warnings []error) {
			return []error{deployErr}, []error{}
		},
		lintDockerPushTask: func(task manifest.DockerPush, taskID string, fs afero.Afero) (errs []error, warnings []error) {
			return []error{dockerPushErr}, []error{dockerPushWarn}

		},
		lintDeployMLZipTask: func(task manifest.DeployMLZip, taskID string) (errs []error, warnings []error) {
			return []error{deployMlZipErr}, []error{}

		},
		lintDeployMLModulesTask: func(task manifest.DeployMLModules, taskID string) (errs []error, warnings []error) {
			return []error{}, []error{deployMlModulesWarn}

		},
	}

	result := taskLinter.Lint(man)

	assert.Equal(t, []error{runErr1, runErr2, deployErr, runErr1, runErr2, dockerPushErr, deployMlZipErr}, result.Errors)
	assert.Equal(t, []error{runWarn1, runWarn1, dockerPushWarn, deployMlModulesWarn}, result.Warnings)
}

func TestRunTaskWithScriptAndImage(t *testing.T) {
	taskLinter := testTaskLinter()
	man := manifest.Manifest{}
	man.Tasks = []manifest.Task{
		manifest.Run{
			Script: "./build.sh",
			Docker: manifest.Docker{
				Image: "alpine",
			},
		},
	}

	result := taskLinter.Lint(man)
	if assert.Len(t, result.Warnings, 1) {
		assertFileError(t, "./build.sh", result.Warnings[0])
	}
}

func TestRunTaskWithScriptAndImageWithPasswordAndUsername(t *testing.T) {
	taskLinter := testTaskLinter()
	taskLinter.Fs.WriteFile("build.sh", []byte("foo"), 0777)
	man := manifest.Manifest{}
	man.Tasks = []manifest.Task{
		manifest.Run{
			Script: "./build.sh",
			Docker: manifest.Docker{
				Image:    "alpine",
				Password: "secret",
				Username: "Michiel",
			},
		},
	}

	result := taskLinter.Lint(man)
	assert.Len(t, result.Errors, 0)
}

func TestRunTaskWithScriptAndImageAndOnlyPassword(t *testing.T) {
	taskLinter := testTaskLinter()
	taskLinter.Fs.WriteFile("build.sh", []byte("foo"), 0777)
	man := manifest.Manifest{}
	man.Tasks = []manifest.Task{
		manifest.Run{
			Script: "./build.sh",
			Docker: manifest.Docker{
				Image:    "alpine",
				Password: "secret",
			},
		},
	}

	result := taskLinter.Lint(man)
	assert.Len(t, result.Errors, 1)
	assertMissingField(t, "tasks[0] run.docker.username", result.Errors[0])
}
func TestRunTaskWithScriptAndImageAndOnlyUsername(t *testing.T) {
	taskLinter := testTaskLinter()
	taskLinter.Fs.WriteFile("build.sh", []byte("foo"), 0777)
	man := manifest.Manifest{}
	man.Tasks = []manifest.Task{
		manifest.Run{
			Script: "./build.sh",
			Docker: manifest.Docker{
				Image:    "alpine",
				Username: "Michiel",
			},
		},
	}

	result := taskLinter.Lint(man)
	assert.Len(t, result.Errors, 1)
	assertMissingField(t, "tasks[0] run.docker.password", result.Errors[0])
}

func TestRunTaskScriptFileExists(t *testing.T) {
	taskLinter := testTaskLinter()
	taskLinter.Fs.WriteFile("build.sh", []byte("foo"), 0777)

	man := manifest.Manifest{}
	man.Tasks = []manifest.Task{
		manifest.Run{
			Script: "./build.sh",
			Docker: manifest.Docker{
				Image: "alpine",
			},
		},
	}

	result := taskLinter.Lint(man)
	assert.Len(t, result.Errors, 0)
}

func TestRunTaskScriptAcceptsArguments(t *testing.T) {
	taskLinter := testTaskLinter()
	taskLinter.Fs.WriteFile("build.sh", []byte("foo"), 0777)

	for _, script := range []string{"./build.sh", "build.sh", "./build.sh --arg 1", "build.sh some args"} {
		man := manifest.Manifest{}
		man.Tasks = []manifest.Task{
			manifest.Run{
				Script: script,
				Docker: manifest.Docker{
					Image: "alpine",
				},
			},
		}

		result := taskLinter.Lint(man)
		assert.Len(t, result.Errors, 0)
	}
}

func TestCFDeployTaskWithEmptyTask(t *testing.T) {
	taskLinter := testTaskLinter()
	man := manifest.Manifest{}
	man.Tasks = []manifest.Task{
		manifest.DeployCF{Manifest: "manifest.yml"},
	}

	result := taskLinter.Lint(man)
	assert.Len(t, result.Errors, 4)
	assertMissingField(t, "tasks[0] deploy-cf.api", result.Errors[0])
	assertMissingField(t, "tasks[0] deploy-cf.space", result.Errors[1])
	assertMissingField(t, "tasks[0] deploy-cf.org", result.Errors[2])
	assertFileError(t, "manifest.yml", result.Errors[3])
}

func TestCFDeployTaskWithEmptyTestDomain(t *testing.T) {
	taskLinter := testTaskLinter()
	taskLinter.Fs.WriteFile("manifest.yml", []byte("foo"), 0777)
	allesOk := manifest.Manifest{}
	allesOk.Tasks = []manifest.Task{
		manifest.DeployCF{
			API:      "((cloudfoundry.api-dev))",
			Org:      "Something",
			Space:    "Something",
			Manifest: "manifest.yml"},
	}

	result := taskLinter.Lint(allesOk)
	assert.Len(t, result.Errors, 0)

	noAPIDefined := manifest.Manifest{}
	noAPIDefined.Tasks = []manifest.Task{
		manifest.DeployCF{
			API:      "",
			Org:      "Something",
			Space:    "Something",
			Manifest: "manifest.yml"},
	}

	result = taskLinter.Lint(noAPIDefined)
	assert.Len(t, result.Errors, 1)
	assertMissingField(t, "tasks[0] deploy-cf.api", result.Errors[0])

	randomAPIDefinedWithoutTestDomain := manifest.Manifest{}
	randomAPIDefinedWithoutTestDomain.Tasks = []manifest.Task{
		manifest.DeployCF{
			API:      "someRandomApi",
			Org:      "Something",
			Space:    "Something",
			Manifest: "manifest.yml"},
	}

	result = taskLinter.Lint(randomAPIDefinedWithoutTestDomain)
	assert.Len(t, result.Errors, 1)
	assertMissingField(t, "tasks[0] deploy-cf.testDomain", result.Errors[0])

}

func TestDockerPushTaskWithEmptyTask(t *testing.T) {
	taskLinter := testTaskLinter()
	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DockerPush{},
		},
	}

	result := taskLinter.Lint(man)
	assert.Len(t, result.Errors, 4)
	assertMissingField(t, "tasks[0] docker-push.username", result.Errors[0])
	assertMissingField(t, "tasks[0] docker-push.password", result.Errors[1])
	assertMissingField(t, "tasks[0] docker-push.image", result.Errors[2])
	assertFileError(t, "Dockerfile", result.Errors[3])

}

func TestDockerPushTaskWithBadRepo(t *testing.T) {
	taskLinter := testTaskLinter()
	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DockerPush{
				Username: "asd",
				Password: "asd",
				Image:    "asd",
			},
		},
	}

	result := taskLinter.Lint(man)
	assert.Len(t, result.Errors, 2)
	assertInvalidField(t, "tasks[0] docker-push.image", result.Errors[0])
	assertFileError(t, "Dockerfile", result.Errors[1])

}

func TestDockerPushTaskWhenDockerfileIsMissing(t *testing.T) {
	taskLinter := testTaskLinter()
	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DockerPush{
				Username: "asd",
				Password: "asd",
				Image:    "asd/asd",
			},
		},
	}

	result := taskLinter.Lint(man)
	assert.Len(t, result.Errors, 1)
	assertFileError(t, "Dockerfile", result.Errors[0])
}

func TestDockerPushTaskWithCorrectData(t *testing.T) {
	taskLinter := testTaskLinter()
	taskLinter.Fs.WriteFile("Dockerfile", []byte("FROM ubuntu"), 0777)

	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DockerPush{
				Username: "asd",
				Password: "asd",
				Image:    "asd/asd",
				Vars: map[string]string{
					"A": "a",
					"B": "b",
				},
			},
		},
	}

	result := taskLinter.Lint(man)
	assert.Len(t, result.Errors, 0)
}

func TestDockerCompose_Happy(t *testing.T) {
	taskLinter := testTaskLinter()
	taskLinter.Fs.WriteFile("docker-compose.yml", []byte(validDockerCompose), 0777)

	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DockerCompose{Service: "app"}, // empty is ok, everything is optional
			manifest.DockerCompose{
				Name:    "run docker compose",
				Service: "app",
				Vars: manifest.Vars{
					"A": "a",
					"B": "b",
				},
			},
		},
	}

	result := taskLinter.Lint(man)
	assert.Len(t, result.Errors, 0)
}

func TestDockerCompose_HappyWithoutServicesKey(t *testing.T) {
	var compose = `
app1:
  image: appropriate/curl

app2:
  image: appropriate/curl
`

	taskLinter := testTaskLinter()
	taskLinter.Fs.WriteFile("docker-compose.yml", []byte(compose), 0777)

	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DockerCompose{Service: "app2"},
			manifest.DockerCompose{
				Name:    "run docker compose",
				Service: "app1",
				Vars: manifest.Vars{
					"A": "a",
					"B": "b",
				},
			},
		},
	}

	result := taskLinter.Lint(man)
	assert.Len(t, result.Errors, 0)
}

func TestDockerCompose_MissingFile(t *testing.T) {
	taskLinter := testTaskLinter()
	man := manifest.Manifest{
		Tasks: []manifest.Task{manifest.DockerCompose{}},
	}

	result := taskLinter.Lint(man)
	assert.Len(t, result.Errors, 1)
	assertFileError(t, "docker-compose.yml", result.Errors[0])
}

func TestDockerCompose_UnknownService(t *testing.T) {
	taskLinter := testTaskLinter()
	taskLinter.Fs.WriteFile("docker-compose.yml", []byte(validDockerCompose), 0777)

	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DockerCompose{
				Service: "asdf",
			},
		},
	}

	result := taskLinter.Lint(man)
	assert.Len(t, result.Errors, 1)
	assertInvalidFieldInErrors(t, "service", result.Errors)
}

func TestDockerCompose_UnknownServiceWithoutServicesKey(t *testing.T) {

	var compose = `
app1:
  image: appropriate/curl
app2:
  image: appropriate/curl
`
	taskLinter := testTaskLinter()
	taskLinter.Fs.WriteFile("docker-compose.yml", []byte(compose), 0777)

	man := manifest.Manifest{
		Tasks: []manifest.Task{
			manifest.DockerCompose{
				Service: "asdf",
			},
		},
	}

	result := taskLinter.Lint(man)
	assert.Len(t, result.Errors, 1)
	assertInvalidFieldInErrors(t, "service", result.Errors)
}

func TestLintsSubTasksInDeployCF(t *testing.T) {

	man := manifest.Manifest{}
	taskLinter := testTaskLinter()

	man.Tasks = []manifest.Task{
		manifest.DeployCF{
			API:   "((cloudfoundry.api-dev))",
			Org:   "org",
			Space: "space",
			PrePromote: []manifest.Task{
				manifest.Run{
					ManualTrigger: true,
				},
				manifest.DockerCompose{},
				manifest.DeployCF{},
				manifest.DockerPush{},
			},
		},
	}

	result := taskLinter.Lint(man)
	assert.Len(t, result.Errors, 13)

	assertInvalidFieldInErrors(t, "tasks[0].pre_promote[0] run.manual_trigger", result.Errors)
	assertMissingFieldInErrors(t, "tasks[0].pre_promote[0] run.script", result.Errors)
	assertMissingFieldInErrors(t, "tasks[0].pre_promote[0] run.docker.image", result.Errors)
	assertInvalidFieldInErrors(t, "tasks[0].pre_promote[2] run.type", result.Errors)
	assertInvalidFieldInErrors(t, "tasks[0].pre_promote[3] run.type", result.Errors)
	assertMissingFieldInErrors(t, "tasks[0].pre_promote[0] run.docker.image", result.Errors)
	assertFileErrorInErrors(t, "docker-compose.yml", result.Errors)
}

func TestConsumerIntegrationTestTaskHasRequiredFieldsInPrePromote(t *testing.T) {
	man := manifest.Manifest{}
	taskLinter := testTaskLinter()

	man.Tasks = []manifest.Task{
		manifest.DeployCF{
			API:   "cf-api",
			Org:   "cf-org",
			Space: "cf-space",
			PrePromote: manifest.TaskList{
				manifest.ConsumerIntegrationTest{},
			},
			TestDomain: "some.domain.io",
		},
	}

	result := taskLinter.Lint(man)
	if assert.Len(t, result.Errors, 3) {
		assertMissingField(t, "tasks[0].pre_promote[0] consumer-integration-test.consumer", result.Errors[0])
		assertMissingField(t, "tasks[0].pre_promote[0] consumer-integration-test.consumer_host", result.Errors[1])
		assertMissingField(t, "tasks[0].pre_promote[0] consumer-integration-test.script", result.Errors[2])
	}
}

func TestConsumerIntegrationTestTaskHasRequiredFieldsOutsidePrePromote(t *testing.T) {
	man := manifest.Manifest{}
	taskLinter := testTaskLinter()

	man.Tasks = []manifest.Task{
		manifest.ConsumerIntegrationTest{},
	}

	result := taskLinter.Lint(man)
	if assert.Len(t, result.Errors, 4) {
		assertMissingField(t, "tasks[0] consumer-integration-test.consumer", result.Errors[0])
		assertMissingField(t, "tasks[0] consumer-integration-test.consumer_host", result.Errors[1])
		assertMissingField(t, "tasks[0] consumer-integration-test.provider_host", result.Errors[2])
		assertMissingField(t, "tasks[0] consumer-integration-test.script", result.Errors[3])
	}
}

func TestDeployMLZipTaskHasRequiredFields(t *testing.T) {
	man := manifest.Manifest{}
	taskLinter := testTaskLinter()

	man.Tasks = []manifest.Task{
		manifest.DeployMLZip{},
	}

	result := taskLinter.Lint(man)
	if assert.Len(t, result.Errors, 2) {
		assertMissingField(t, "tasks[0] deploy-ml.target", result.Errors[0])
		assertMissingField(t, "tasks[0] deploy-ml.deploy_zip", result.Errors[1])
	}
}

func TestDeployMLModulesTaskHasRequiredFields(t *testing.T) {
	man := manifest.Manifest{}
	taskLinter := testTaskLinter()

	man.Tasks = []manifest.Task{
		manifest.DeployMLModules{},
	}

	result := taskLinter.Lint(man)
	if assert.Len(t, result.Errors, 2) {
		assertMissingField(t, "tasks[0] deploy-ml.target", result.Errors[0])
		assertMissingField(t, "tasks[0] deploy-ml.ml_modules_version", result.Errors[1])
	}
}

func TestCannotSetPassedInPrePromoteTasks(t *testing.T) {

	man := manifest.Manifest{}
	taskLinter := testTaskLinter()
	taskLinter.Fs.WriteFile("docker-compose.yml", []byte(validDockerCompose), 0777)

	man.Tasks = []manifest.Task{
		manifest.DeployCF{
			API:        "api",
			Org:        "org",
			Space:      "space",
			TestDomain: "foo.com",
			PrePromote: []manifest.Task{
				manifest.Run{Script: "/foo", Docker: manifest.Docker{Image: "foo"}, Parallel: true},
				manifest.DockerCompose{Service: "app", Parallel: true},
			},
		},
	}

	result := taskLinter.Lint(man)
	if assert.Len(t, result.Errors, 2) {
		assertInvalidField(t, "tasks[0].pre_promote[0] run.passed", result.Errors[0])
		assertInvalidField(t, "tasks[0].pre_promote[1] docker-compose.passed", result.Errors[1])
	}
}

func TestLintsTheTimeoutInDeployTask(t *testing.T) {
	man := manifest.Manifest{}
	taskLinter := testTaskLinter()

	man.Tasks = []manifest.Task{
		manifest.DeployCF{
			API:        "asdf",
			Space:      "asdf",
			Org:        "asdf",
			TestDomain: "asdf",
			Timeout:    "notAValidDuration",
		},
	}

	result := taskLinter.Lint(man)
	assertInvalidField(t, "tasks[0] deploy-cf.timeout", result.Errors[0])
}

func TestCFDeployTaskWithManifestFromArtifacts(t *testing.T) {
	taskLinter := testTaskLinter()
	man := manifest.Manifest{}
	man.Tasks = []manifest.Task{
		manifest.DeployCF{
			Manifest:   "../artifacts/manifest.yml",
			API:        "api",
			Space:      "space",
			Org:        "org",
			TestDomain: "test.domain",
		},
	}

	result := taskLinter.Lint(man)
	assert.Len(t, result.Errors, 0)
	assert.Len(t, result.Warnings, 1)

	println(result.Warnings)
}

func TestAttempts(t *testing.T) {
	taskLinter := testTaskLinter()
	man := manifest.Manifest{}
	man.Tasks = []manifest.Task{
		manifest.Run{Retries: -20},
		manifest.DockerPush{Retries: -1},
		manifest.DockerCompose{Retries: -324},
		manifest.DeployCF{
			Retries: -3,
			PrePromote: []manifest.Task{
				manifest.Run{Retries: -100},
			},
		},
		manifest.ConsumerIntegrationTest{Retries: -1},
		manifest.DeployMLZip{Retries: -200},
		manifest.DeployMLModules{Retries: -1337},
	}

	result := taskLinter.Lint(man)
	assertInvalidFieldInErrors(t, "run.retries", result.Errors)
	assertInvalidFieldInErrors(t, "deploy-cf.retries", result.Errors)
	assertInvalidFieldInErrors(t, "tasks[3].pre_promote[0] run.retries", result.Errors)
	assertInvalidFieldInErrors(t, "docker-push.retries", result.Errors)
	assertInvalidFieldInErrors(t, "docker-compose.retries", result.Errors)
	assertInvalidFieldInErrors(t, "consumer-integration-test.retries", result.Errors)
	assertInvalidFieldInErrors(t, "deploy-ml-zip.retries", result.Errors)
	assertInvalidFieldInErrors(t, "deploy-ml-modules.retries", result.Errors)

	man.Tasks = []manifest.Task{
		manifest.Run{Retries: 6},
		manifest.DockerPush{Retries: 6},
		manifest.DockerCompose{Retries: 6},
		manifest.DeployCF{Retries: 6, PrePromote: []manifest.Task{
			manifest.Run{Retries: 6},
		}},
		manifest.ConsumerIntegrationTest{Retries: 6},
		manifest.DeployMLZip{Retries: 6},
		manifest.DeployMLModules{Retries: 6},
	}

	result2 := taskLinter.Lint(man)
	assertInvalidFieldInErrors(t, "run.retries", result2.Errors)
	assertInvalidFieldInErrors(t, "deploy-cf.retries", result2.Errors)
	assertInvalidFieldInErrors(t, "tasks[3].pre_promote[0] run.retries", result2.Errors)
	assertInvalidFieldInErrors(t, "docker-push.retries", result2.Errors)
	assertInvalidFieldInErrors(t, "docker-compose.retries", result2.Errors)
	assertInvalidFieldInErrors(t, "consumer-integration-test.retries", result2.Errors)
	assertInvalidFieldInErrors(t, "deploy-ml-zip.retries", result2.Errors)
	assertInvalidFieldInErrors(t, "deploy-ml-modules.retries", result2.Errors)

	man.Tasks = []manifest.Task{
		manifest.Run{Retries: 0},
		manifest.DockerPush{Retries: 2},
		manifest.DockerCompose{Retries: 3},
		manifest.DeployCF{Retries: 4,
			PrePromote: []manifest.Task{
				manifest.Run{Retries: 0},
			},
		},
		manifest.ConsumerIntegrationTest{Retries: 5},
		manifest.DeployMLZip{Retries: 4},
		manifest.DeployMLModules{Retries: 3},
	}
	result3 := taskLinter.Lint(man)
	assertInvalidFieldShouldNotBeInErrors(t, "run.retries", result3.Errors)
	assertInvalidFieldShouldNotBeInErrors(t, "deploy-cf.retries", result3.Errors)
	assertInvalidFieldShouldNotBeInErrors(t, "tasks[3].pre_promote[0] run.retries", result3.Errors)
	assertInvalidFieldShouldNotBeInErrors(t, "docker-push.retries", result3.Errors)
	assertInvalidFieldShouldNotBeInErrors(t, "docker-compose.retries", result3.Errors)
	assertInvalidFieldShouldNotBeInErrors(t, "consumer-integration-test.retries", result3.Errors)
	assertInvalidFieldShouldNotBeInErrors(t, "deploy-ml-zip.retries", result3.Errors)
	assertInvalidFieldShouldNotBeInErrors(t, "deploy-ml-modules.retries", result3.Errors)
}
