package linters

import (
	"fmt"
	"regexp"

	"strings"

	"time"

	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/defaults"
	"github.com/springernature/halfpipe/linters/errors"
	"github.com/springernature/halfpipe/linters/filechecker"
	"github.com/springernature/halfpipe/manifest"
	"gopkg.in/yaml.v2"
)

type taskLinter struct {
	Fs afero.Afero
}

func NewTasksLinter(fs afero.Afero) taskLinter {
	return taskLinter{fs}
}

func (linter taskLinter) Lint(man manifest.Manifest) (result LintResult) {
	result.Linter = "Tasks"
	result.DocsURL = "https://docs.halfpipe.io/manifest/#tasks"

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
			result.AddError(errors.NewInvalidField(taskID+" deploy-cf.timeout", err.Error()))
		}
	}

	if cf.Retries < 0 || cf.Retries > 5 {
		result.AddError(errors.NewInvalidField(taskID+" deploy-cf.retries", "must be between 0 and 5"))
	}

	if strings.HasPrefix(cf.Manifest, "../artifacts/") {
		result.AddWarning(errors.NewFileError(cf.Manifest, "this file must be saved as an artifact in a previous task"))
	} else if err := filechecker.CheckFile(linter.Fs, cf.Manifest, false); err != nil {
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

	if docker.Retries < 0 || docker.Retries > 5 {
		result.AddError(errors.NewInvalidField(taskID+" docker-push.retries", "must be between 0 and 5"))
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

	if run.Retries < 0 || run.Retries > 5 {
		result.AddError(errors.NewInvalidField(taskID+" run.retries", "must be between 0 and 5"))
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
	if dc.Retries < 0 || dc.Retries > 5 {
		result.AddError(errors.NewInvalidField(taskID+" docker-compose.retries", "must be between 0 and 5"))
	}

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

	if cit.Retries < 0 || cit.Retries > 5 {
		result.AddError(errors.NewInvalidField(taskID+" consumer-integration-test.retries", "must be between 0 and 5"))
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

	if mlTask.Retries < 0 || mlTask.Retries > 5 {
		result.AddError(errors.NewInvalidField(taskID+" deploy-ml-zip.retries", "must be between 0 and 5"))
	}

}

func (linter taskLinter) lintDeployMLModulesTask(mlTask manifest.DeployMLModules, taskID string, result *LintResult) {
	if len(mlTask.Targets) == 0 {
		result.AddError(errors.NewMissingField(taskID + " deploy-ml.target"))
	}
	if mlTask.MLModulesVersion == "" {
		result.AddError(errors.NewMissingField(taskID + " deploy-ml.ml_modules_version"))
	}

	if mlTask.Retries < 0 || mlTask.Retries > 5 {
		result.AddError(errors.NewInvalidField(taskID+" deploy-ml-modules.retries", "must be between 0 and 5"))
	}
}
