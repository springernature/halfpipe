package linters

import (
	"fmt"
	"regexp"

	"strings"

	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/defaults"
	"github.com/springernature/halfpipe/linters/errors"
	"github.com/springernature/halfpipe/linters/filechecker"
	"github.com/springernature/halfpipe/manifest"
	"gopkg.in/yaml.v2"
	"time"
)

type taskLinter struct {
	Fs afero.Afero
}

func NewTasksLinter(fs afero.Afero) taskLinter {
	return taskLinter{fs}
}

func (linter taskLinter) Lint(man manifest.Manifest) (result LintResult) {
	result.Linter = "Tasks"
	result.DocsURL = "https://docs.halfpipe.io/docs/manifest/#tasks"

	if len(man.Tasks) == 0 {
		result.AddError(errors.NewMissingField("tasks"))
		return
	}

	var lintTasks func(string, []manifest.Task)
	lintTasks = func(listName string, tasks []manifest.Task) {
		for i, t := range tasks {
			taskID := fmt.Sprintf("%s[%v]", listName, i)
			switch task := t.(type) {
			case manifest.Run:
				linter.lintRunTask(task, taskID, &result)
			case manifest.DeployCF:
				linter.lintDeployCFTask(task, taskID, &result)
				lintTasks(fmt.Sprintf("%s.pre_promote", taskID), task.PrePromote)
			case manifest.DockerPush:
				linter.lintDockerPushTask(task, taskID, &result)
			case manifest.DockerCompose:
				linter.lintDockerComposeTask(task, taskID, &result)
			case manifest.ConsumerIntegrationTest:
				if listName == "tasks" {
					linter.lintConsumerIntegrationTestTask(task, taskID, true, &result)
				} else {
					linter.lintConsumerIntegrationTestTask(task, taskID, false, &result)
				}
			case manifest.DeployMLZip:
				linter.lintDeployMLZipTask(task, taskID, &result)
			case manifest.DeployMLModules:
				linter.lintDeployMLModulesTask(task, taskID, &result)
			default:
				result.AddError(errors.NewInvalidField("task", fmt.Sprintf("%s is not a known task", taskID)))
			}
		}
	}

	lintTasks("tasks", man.Tasks)

	for i, t := range man.OnFailure {
		taskID := fmt.Sprintf("on_failure[%v] ", i)
		passedError := func(taskName string) errors.InvalidFieldError {
			return errors.NewInvalidField(taskID+taskName+".passed", "You are not allowed to set 'passed' inside an on_failure task")
		}
		manualTriggerError := func(taskName string) errors.InvalidFieldError {
			return errors.NewInvalidField(taskID+taskName+".manual_trigger", "You are not allowed to have a manual trigger inside an on_failure task")
		}
		switch task := t.(type) {
		case manifest.Run:
			if task.Parallel {
				result.AddError(passedError("run"))
			}
			if task.ManualTrigger == true {
				result.AddError(manualTriggerError("run"))
			}
		case manifest.DockerCompose:
			if task.Parallel {
				result.AddError(passedError("docker-compose"))
			}
			if task.ManualTrigger == true {
				result.AddError(manualTriggerError("docker-compose"))
			}
		case manifest.DockerPush:
			if task.Parallel {
				result.AddError(passedError("docker-push"))
			}
			if task.ManualTrigger == true {
				result.AddError(manualTriggerError("docker-push"))
			}
		case manifest.DeployCF:
			if task.Parallel {
				result.AddError(passedError("deploy-cf"))
			}
			if task.ManualTrigger == true {
				result.AddError(manualTriggerError("deploy-cf"))
			}
		}
	}

	lintTasks("onFailureTasks", man.OnFailure)

	return
}

func (linter taskLinter) lintDeployCFTask(cf manifest.DeployCF, taskID string, result *LintResult) {
	if cf.API == "" {
		result.AddError(errors.NewMissingField(taskID + " deploy-cf.api"))
	}
	if cf.Space == "" {
		result.AddError(errors.NewMissingField(taskID + " deploy-cf.space"))
	}
	if cf.Org == "" {
		result.AddError(errors.NewMissingField(taskID + " deploy-cf.org"))
	}
	if cf.TestDomain == "" {
		_, found := defaults.DefaultValues.CfTestDomains[cf.API]
		if cf.API != "" && !found {
			result.AddError(errors.NewMissingField(taskID + " deploy-cf.testDomain"))
		}
	}

	if cf.Timeout != "" {
		_, err := time.ParseDuration(cf.Timeout)
		if err != nil {
			result.AddError(errors.NewInvalidField(taskID + " deploy-cf.timeout", err.Error()))
		}
	}

	if err := filechecker.CheckFile(linter.Fs, cf.Manifest, false); err != nil {
		result.AddError(err)
	}

	for i, prePromote := range cf.PrePromote {
		ppTaskID := fmt.Sprintf("%s.pre_promote[%v]", taskID, i)
		switch task := prePromote.(type) {
		case manifest.Run:
			if task.ManualTrigger == true {
				result.AddError(errors.NewInvalidField(ppTaskID+" run.manual_trigger", "You are not allowed to have a manual trigger inside a pre promote task"))
			}
			if task.Parallel {
				result.AddError(errors.NewInvalidField(ppTaskID+" run.passed", "You are not allowed to set 'passed' inside a pre promote task"))
			}
		case manifest.DockerCompose:
			if task.ManualTrigger == true {
				result.AddError(errors.NewInvalidField(ppTaskID+" docker-compose.manual_trigger", "You are not allowed to have a manual trigger inside a pre promote task"))
			}
			if task.Parallel {
				result.AddError(errors.NewInvalidField(ppTaskID+" docker-compose.passed", "You are not allowed to set 'passed' inside a pre promote task"))
			}
		case manifest.DockerPush, manifest.DeployCF:
			result.AddError(errors.NewInvalidField(ppTaskID+" run.type", "You are not allowed to have a 'deploy-cf' or 'docker-push' task as a pre promote"))
		}

	}

	return
}

func (linter taskLinter) lintDockerPushTask(docker manifest.DockerPush, taskID string, result *LintResult) {
	if docker.Username == "" {
		result.AddError(errors.NewMissingField(taskID + " docker-push.username"))
	}
	if docker.Password == "" {
		result.AddError(errors.NewMissingField(taskID + " docker-push.password"))
	}
	if docker.Image == "" {
		result.AddError(errors.NewMissingField(taskID + " docker-push.image"))
	} else {
		matched, _ := regexp.Match(`^(.*)/(.*)$`, []byte(docker.Image))
		if !matched {
			result.AddError(errors.NewInvalidField(taskID+" docker-push.image", "must be specified as 'user/image' or 'registry/user/image'"))
		}
	}

	if err := filechecker.CheckFile(linter.Fs, "Dockerfile", false); err != nil {
		result.AddError(err)
	}

	return
}

func (linter taskLinter) lintRunTask(run manifest.Run, taskID string, result *LintResult) {
	if run.Script == "" {
		result.AddError(errors.NewMissingField(taskID + " run.script"))
	} else {
		// Possible for script to have args,
		fields := strings.Fields(strings.TrimSpace(run.Script))
		command := fields[0]
		if err := filechecker.CheckFile(linter.Fs, command, true); err != nil {
			result.AddWarning(err)
		}
	}

	if run.Docker.Image == "" {
		result.AddError(errors.NewMissingField(taskID + " run.docker.image"))
	}

	if run.Docker.Username != "" && run.Docker.Password == "" {
		result.AddError(errors.NewMissingField(taskID + " run.docker.password"))
	}
	if run.Docker.Password != "" && run.Docker.Username == "" {
		result.AddError(errors.NewMissingField(taskID + " run.docker.username"))
	}

	return
}

func (linter taskLinter) lintDockerComposeTask(dc manifest.DockerCompose, taskID string, result *LintResult) {
	if err := filechecker.CheckFile(linter.Fs, "docker-compose.yml", false); err != nil {
		result.AddError(err)
		return
	}

	linter.lintDockerComposeService(dc.Service, result)
	return
}

func (linter taskLinter) lintDockerComposeService(service string, result *LintResult) {
	content, err := linter.Fs.ReadFile("docker-compose.yml")
	if err != nil {
		result.AddError(err)
		return
	}

	var compose struct {
		Services map[string]interface{} `yaml:"services"`
	}
	err = yaml.Unmarshal(content, &compose)
	if err != nil {
		result.AddError(err)
		return
	}

	if _, ok := compose.Services[service]; ok {
		return
	}

	var composeWithoutServices map[string]interface{}
	err = yaml.Unmarshal(content, &composeWithoutServices)
	if err != nil {
		result.AddError(err)
		return
	}

	if _, ok := composeWithoutServices[service]; ok {
		return
	}

	result.AddError(errors.NewInvalidField("service", fmt.Sprintf("Could not find service '%s' in docker-compose.yml", service)))
	return
}

func (linter taskLinter) lintConsumerIntegrationTestTask(cit manifest.ConsumerIntegrationTest, taskID string, providerHostRequired bool, result *LintResult) {
	if cit.Consumer == "" {
		result.AddError(errors.NewMissingField(taskID + " consumer-integration-test.consumer"))
	}
	if cit.ConsumerHost == "" {
		result.AddError(errors.NewMissingField(taskID + " consumer-integration-test.consumer_host"))
	}
	if providerHostRequired {
		if cit.ProviderHost == "" {
			result.AddError(errors.NewMissingField(taskID + " consumer-integration-test.provider_host"))
		}
	}
	if cit.Script == "" {
		result.AddError(errors.NewMissingField(taskID + " consumer-integration-test.script"))
	}
	return
}

func (linter taskLinter) lintDeployMLZipTask(mlTask manifest.DeployMLZip, taskID string, result *LintResult) {
	if len(mlTask.Targets) == 0 {
		result.AddError(errors.NewMissingField(taskID + " deploy-ml.target"))
	}

	if mlTask.DeployZip == "" {
		result.AddError(errors.NewMissingField(taskID + " deploy-ml.deploy_zip"))
	}
}

func (linter taskLinter) lintDeployMLModulesTask(mlTask manifest.DeployMLModules, taskID string, result *LintResult) {
	if len(mlTask.Targets) == 0 {
		result.AddError(errors.NewMissingField(taskID + " deploy-ml.target"))
	}
	if mlTask.MLModulesVersion == "" {
		result.AddError(errors.NewMissingField(taskID + " deploy-ml.ml_modules_version"))
	}
}
