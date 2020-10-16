package pipeline

import (
	"fmt"
	"github.com/springernature/halfpipe/defaults"
	"regexp"
	"strings"

	"path/filepath"

	"sort"

	"path"

	cfManifest "code.cloudfoundry.org/cli/util/manifest"
	boshTemplate "github.com/cloudfoundry/bosh-cli/director/template"
	"github.com/concourse/concourse/atc"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
)

type Renderer interface {
	Render(manifest manifest.Manifest) atc.Config
}

type CfManifestReader func(pathToManifest string, pathsToVarsFiles []string, vars []boshTemplate.VarKV) ([]cfManifest.Application, error)

type pipeline struct {
	fs             afero.Afero
	readCfManifest CfManifestReader
}

func NewPipeline(cfManifestReader CfManifestReader, fs afero.Afero) pipeline {
	return pipeline{readCfManifest: cfManifestReader, fs: fs}
}

const artifactsResourceName = "gcp-resource"
const artifactsName = "artifacts"
const artifactsOutDir = "artifacts-out"
const artifactsInDir = "artifacts"
const artifactsOnFailureName = "artifacts-on-failure"
const artifactsOutDirOnFailure = "artifacts-out-failure"

const gitDir = "git"
const gitGetAttempts = 2

const dockerBuildTmpDir = "docker_build"

const versionName = "version"
const versionGetAttempts = 2

func restoreArtifactTask(man manifest.Manifest) atc.Step {
	// This function is used in pipeline.artifactResource for some reason to lowercase
	// and remove chars that are not part of the regex in the folder in the config..
	// So we must reuse it.
	filter := func(str string) string {
		reg := regexp.MustCompile(`[^a-z0-9\-]+`)
		return reg.ReplaceAllString(strings.ToLower(str), "")
	}

	jsonKey := config.ArtifactsJSONKey
	if man.ArtifactConfig.JSONKey != "" {
		jsonKey = man.ArtifactConfig.JSONKey
	}

	BUCKET := config.ArtifactsBucket
	if man.ArtifactConfig.Bucket != "" {
		BUCKET = man.ArtifactConfig.Bucket
	}

	config := atc.TaskConfig{
		Platform:  "linux",
		RootfsURI: "",
		ImageResource: &atc.ImageResource{
			Type: "registry-image",
			Source: atc.Source{
				"repository": config.DockerRegistry + "gcp-resource",
				"tag":        "stable",
				"password":   "((halfpipe-gcr.private_key))",
				"username":   "_json_key",
			},
		},
		Params: map[string]string{
			"BUCKET":       BUCKET,
			"FOLDER":       path.Join(filter(man.Team), filter(man.PipelineName())),
			"JSON_KEY":     jsonKey,
			"VERSION_FILE": "git/.git/ref",
		},
		Run: atc.TaskRunConfig{
			Path: "/opt/resource/download",
			Dir:  artifactsInDir,
			Args: []string{"."},
		},
		Inputs: []atc.TaskInputConfig{
			{
				Name: manifest.GitTrigger{}.GetTriggerName(),
			},
		},
		Outputs: []atc.TaskOutputConfig{
			{
				Name: artifactsInDir,
			},
		},
	}

	return atc.Step{
		Config: &atc.RetryStep{
			Step: &atc.TimeoutStep{
				Step: &atc.TaskStep{
					Name:   "get-artifact",
					Config: &config,
				},
				Duration: "1h",
			},
			Attempts: 2,
		},
	}
}

func (p pipeline) initialPlan(man manifest.Manifest, task manifest.Task) []atc.Step {
	_, isUpdateTask := task.(manifest.Update)
	versioningEnabled := man.FeatureToggles.Versioned()

	var gets []atc.GetStep
	for _, trigger := range man.Triggers {
		switch trigger := trigger.(type) {
		case manifest.GitTrigger:
			gitClone := atc.GetStep{
				Name: trigger.GetTriggerName(),
			}
			if trigger.Shallow {
				gitClone.Params = map[string]interface{}{
					"depth": 1,
				}
			}
			gets = append(gets, gitClone)

		case manifest.TimerTrigger:
			if isUpdateTask || !versioningEnabled {
				gets = append(gets, atc.GetStep{Name: trigger.GetTriggerName()})
			}
		case manifest.DockerTrigger:
			if isUpdateTask || !versioningEnabled {
				dockerTrigger := atc.GetStep{
					Name: trigger.GetTriggerName(),
					Params: map[string]interface{}{
						"skip_download": true,
					},
				}

				gets = append(gets, dockerTrigger)

			}
		case manifest.PipelineTrigger:
			if isUpdateTask || !versioningEnabled {
				pipelineTrigger := atc.GetStep{
					Name: trigger.GetTriggerName(),
				}

				gets = append(gets, pipelineTrigger)
			}

		}
	}

	if !isUpdateTask && man.FeatureToggles.Versioned() {
		gets = append(gets, atc.GetStep{
			Name: versionName,
		})
	}

	var attemptsGet []atc.Step
	for i, _ := range gets {
		if gets[i].Name == "version" {
			attemptsGet = append(attemptsGet, atc.Step{
				Config: &atc.TimeoutStep{
					Step: &atc.RetryStep{
						Step:     &gets[i],
						Attempts: 2,
					},
					Duration: "1m",
				},
				UnknownFields: nil,
			})
		} else {
			attemptsGet = append(attemptsGet, atc.Step{
				Config: &atc.RetryStep{
					Step:     &gets[i],
					Attempts: 2,
				},
				UnknownFields: nil,
			})
		}
	}

	timeoutStep := atc.Step{
		Config: &atc.TimeoutStep{
			Step: &atc.InParallelStep{
				Config: atc.InParallelConfig{
					Steps:    attemptsGet,
					FailFast: true,
				},
			},
			Duration: task.GetTimeout(),
		},
	}
	steps := []atc.Step{
		timeoutStep,
	}

	if task.ReadsFromArtifacts() {
		steps = append(steps, restoreArtifactTask(man))
	}

	return steps
}

func (p pipeline) dockerPushResources(tasks manifest.TaskList) (resourceConfigs atc.ResourceConfigs) {
	for _, task := range tasks {
		switch task := task.(type) {
		case manifest.DockerPush:
			resourceConfigs = append(resourceConfigs, p.dockerPushResource(task))
		case manifest.Parallel:
			resourceConfigs = append(resourceConfigs, p.dockerPushResources(task.Tasks)...)
		case manifest.Sequence:
			resourceConfigs = append(resourceConfigs, p.dockerPushResources(task.Tasks)...)
		}
	}

	return resourceConfigs
}
func (p pipeline) pipelineResources(triggers manifest.TriggerList) (resourceTypes atc.ResourceTypes, resourceConfigs atc.ResourceConfigs) {

	for _, trigger := range triggers {
		switch trigger := trigger.(type) {
		case manifest.PipelineTrigger:
			resourceConfigs = append(resourceConfigs, p.pipelineTriggerResource(trigger))
		}
	}

	if len(resourceConfigs) > 0 {
		resourceTypes = append(resourceTypes, halfpipePipelineTriggerResourceType())
	}

	return resourceTypes, resourceConfigs
}

func (p pipeline) cfPushResourceConfig(man manifest.Manifest) atc.ResourceType {
	if man.FeatureToggles.OldDeployResource() {
		return p.halfpipeCfDeployResourceType(true)
	}

	return p.halfpipeCfDeployResourceType(false)
}

func (p pipeline) cfPushResourcesv2(man manifest.Manifest) (resourceTypes atc.ResourceTypes, resourceConfigs atc.ResourceConfigs) {

	for _, task := range man.Tasks.Flatten() {
		switch task := task.(type) {
		case manifest.DeployCF:
			resourceConfig := p.deployCFResource(task, deployCFResourceName(task))
			if _, found := resourceConfigs.Lookup(resourceConfig.Name); !found {
				resourceConfigs = append(resourceConfigs, resourceConfig)
			}
		}
	}

	if len(resourceConfigs) > 0 {
		resourceTypes = append(resourceTypes, p.cfPushResourceConfig(man))
	}

	return
}

func (p pipeline) resourceConfigs(man manifest.Manifest) (resourceTypes atc.ResourceTypes, resourceConfigs atc.ResourceConfigs) {
	for _, trigger := range man.Triggers {
		switch trigger := trigger.(type) {
		case manifest.GitTrigger:
			resourceConfigs = append(resourceConfigs, p.gitResource(trigger))
		case manifest.TimerTrigger:
			resourceTypes = append(resourceTypes, cronResourceType())
			resourceConfigs = append(resourceConfigs, p.cronResource(trigger))
		case manifest.DockerTrigger:
			resourceConfigs = append(resourceConfigs, p.dockerTriggerResource(trigger))
		}
	}

	if man.Tasks.UsesNotifications() {
		resourceTypes = append(resourceTypes, p.slackResourceType())
		resourceConfigs = append(resourceConfigs, p.slackResource())
	}

	if man.Tasks.SavesArtifacts() || man.Tasks.SavesArtifactsOnFailure() {
		resourceTypes = append(resourceTypes, p.gcpResourceType())

		if man.Tasks.SavesArtifacts() {
			resourceConfigs = append(resourceConfigs, p.artifactResource(man))
		}
		if man.Tasks.SavesArtifactsOnFailure() {
			resourceConfigs = append(resourceConfigs, p.artifactResourceOnFailure(man))
		}
	}

	if man.FeatureToggles.Versioned() {
		resourceConfigs = append(resourceConfigs, p.versionResource(man))
	}

	resourceConfigs = append(resourceConfigs, p.dockerPushResources(man.Tasks)...)

	cfResourceTypes, cfResources := p.cfPushResourcesv2(man)
	resourceTypes = append(resourceTypes, cfResourceTypes...)
	resourceConfigs = append(resourceConfigs, cfResources...)

	pipelineResourceTypes, pipelineResources := p.pipelineResources(man.Triggers)
	resourceTypes = append(resourceTypes, pipelineResourceTypes...)
	resourceConfigs = append(resourceConfigs, pipelineResources...)

	return resourceTypes, resourceConfigs
}

func (p pipeline) taskToJobs(task manifest.Task, man manifest.Manifest, previousTaskNames []string) (job atc.JobConfig) {
	initialPlan := p.initialPlan(man, task)
	basePath := man.Triggers.GetGitTrigger().BasePath

	switch task := task.(type) {
	case manifest.Run:
		job = p.runJob(task, man, false, basePath)

	case manifest.DockerCompose:
		job = p.dockerComposeJob(task, man, basePath)

	case manifest.DeployCF:
		job = p.deployCFJob(task, man, basePath)

	case manifest.DockerPush:
		job = p.dockerPushJob(task, basePath)
	case manifest.ConsumerIntegrationTest:
		job = p.consumerIntegrationTestJob(task, man, basePath)

	case manifest.DeployMLZip:
		runTask := ConvertDeployMLZipToRunTask(task, man)
		job = p.runJob(runTask, man, false, basePath)

	case manifest.DeployMLModules:
		runTask := ConvertDeployMLModulesToRunTask(task, man)
		job = p.runJob(runTask, man, false, basePath)
	case manifest.Update:
		job = p.updateJobConfig(task, man.PipelineName(), basePath)
	}

	onFailureChannels := task.GetNotifications().OnFailure
	if task.SavesArtifactsOnFailure() || len(onFailureChannels) > 0 {
		var sequence []atc.Step

		if task.SavesArtifactsOnFailure() {
			s := saveArtifactOnFailurePlan()
			a := atc.Step{
				Config: &atc.RetryStep{
					Step:     &s,
					Attempts: 2,
				},
			}
			sequence = append(sequence, a)
		}

		for _, onFailureChannel := range onFailureChannels {
			sequence = append(sequence, slackOnFailurePlan(onFailureChannel, task.GetNotifications().OnFailureMessage))
		}

		job.OnFailure = &atc.Step{
			Config: &atc.InParallelStep{
				Config: atc.InParallelConfig{
					Steps: sequence,
				},
			},
		}
	}

	onSuccessChannels := task.GetNotifications().OnSuccess
	if len(onSuccessChannels) > 0 {
		var sequence []atc.Step

		for _, onSuccessChannel := range onSuccessChannels {
			sequence = append(sequence, slackOnSuccessPlan(onSuccessChannel, task.GetNotifications().OnSuccessMessage))
		}

		job.OnSuccess = &atc.Step{
			Config: &atc.InParallelStep{
				Config: atc.InParallelConfig{
					Steps: sequence,
				},
			},
		}
	}
	job.PlanSequence = append(initialPlan, job.PlanSequence...)
	//job.Plan = inParallelGets(job)

	configureTriggerOnGets(job, task, man)
	//addTimeout(job, task.GetTimeout())
	addPassedJobsToGets(job, previousTaskNames)
	addBuildLogRetentionSettings(&job, task)

	return job
}

func (p pipeline) taskNamesFromTask(task manifest.Task) (taskNames []string) {
	switch task := task.(type) {
	case manifest.Parallel:
		for _, subTask := range task.Tasks {
			taskNames = append(taskNames, p.taskNamesFromTask(subTask)...)
		}
	case manifest.Sequence:
		taskNames = append(taskNames, task.Tasks[len(task.Tasks)-1].GetName())
	default:
		taskNames = append(taskNames, task.GetName())
	}

	return taskNames
}

func (p pipeline) previousTaskNames(currentIndex int, taskList manifest.TaskList) []string {
	if currentIndex == 0 {
		return []string{}
	}
	return p.taskNamesFromTask(taskList[currentIndex-1])
}

func (p pipeline) Render(man manifest.Manifest) (cfg atc.Config) {
	resourceTypes, resourceConfigs := p.resourceConfigs(man)
	cfg.ResourceTypes = append(cfg.ResourceTypes, resourceTypes...)
	cfg.Resources = append(cfg.Resources, resourceConfigs...)

	for i, task := range man.Tasks {
		switch task := task.(type) {
		case manifest.Parallel:
			for _, subTask := range task.Tasks {
				switch subTask := subTask.(type) {
				case manifest.Sequence:
					previousTasksName := p.previousTaskNames(i, man.Tasks)
					for _, subTask := range subTask.Tasks {
						cfg.Jobs = append(cfg.Jobs, p.taskToJobs(subTask, man, previousTasksName))
						previousTasksName = p.taskNamesFromTask(subTask)
					}
				default:
					cfg.Jobs = append(cfg.Jobs, p.taskToJobs(subTask, man, p.previousTaskNames(i, man.Tasks)))
				}
			}
		default:
			cfg.Jobs = append(cfg.Jobs, p.taskToJobs(task, man, p.previousTaskNames(i, man.Tasks)))
		}
	}

	return cfg
}

func addTimeout(job *atc.JobConfig, timeout string) {
	//for i := range job.Plan {
	//	job.Plan[i].Timeout = timeout
	//}
	//
	//if job.Ensure != nil {
	//	job.Ensure.Timeout = timeout
	//}
}

func addPassedJobsToGets(job atc.JobConfig, passedJobs []string) {
	job.PlanSequence[0].Config.Visit(atc.StepRecursor{
		OnGet: func(step *atc.GetStep) error {
			step.Passed = passedJobs
			return nil
		},
	})
}

func addBuildLogRetentionSettings(job *atc.JobConfig, task manifest.Task) {
	retention := atc.BuildLogRetention{
		MinimumSucceededBuilds: 1,
	}
	if task.GetBuildHistory() != 0 {
		retention.Builds = task.GetBuildHistory()
	}

	job.BuildLogRetention = &retention
}

func configureTriggerOnGets(job atc.JobConfig, task manifest.Task, man manifest.Manifest) {
	if task.IsManualTrigger() {
		return
	}

	versioningEnabled := man.FeatureToggles.Versioned()
	manualGitTrigger := man.Triggers.GetGitTrigger().ManualTrigger

	job.PlanSequence[0].Config.Visit(atc.StepRecursor{
		OnGet: func(step *atc.GetStep) error {
			switch task.(type) {
			case manifest.Update:
				if step.Name == (manifest.GitTrigger{}.GetTriggerName()) {
					step.Trigger = !manualGitTrigger
				} else {
					step.Trigger = true
				}
			default:
				if step.Name == versionName {
					step.Trigger = true
				} else if step.Name == (manifest.GitTrigger{}.GetTriggerName()) {
					step.Trigger = !versioningEnabled && !manualGitTrigger
				} else {
					step.Trigger = !versioningEnabled
				}
			}
			return nil
		},
	})

	//switch task.(type) {
	//case manifest.Update:
	//	for i, step := range gets.Steps {
	//		if step.Get == (manifest.GitTrigger{}.GetTriggerName()) {
	//			gets.Steps[i].Trigger = !manualGitTrigger
	//		} else {
	//			gets.Steps[i].Trigger = true
	//		}
	//	}
	//default:
	//	manifest.GitTrigger{}.GetTriggerName()
	//	for i, step := range gets.Steps {
	//		if step.Get == versionName {
	//			gets.Steps[i].Trigger = true
	//		} else if step.Get == (manifest.GitTrigger{}.GetTriggerName()) {
	//			gets.Steps[i].Trigger = !versioningEnabled && !manualGitTrigger
	//		} else {
	//			gets.Steps[i].Trigger = !versioningEnabled
	//		}
	//	}
	//}
}

func inParallelGets(job *atc.JobConfig) []atc.Plan {
	//var numberOfGets int
	//for i, plan := range job.Plan {
	//	if plan.Get == "" {
	//		numberOfGets = i
	//		break
	//	}
	//}
	//
	//sequence := job.Plan[:numberOfGets]
	//inParallelPlan := atc.PlanSequence{atc.PlanConfig{
	//	InParallel: &atc.InParallelConfig{
	//		Steps:    sequence,
	//		FailFast: true,
	//	},
	//}}
	//job.Plan = append(inParallelPlan, job.Plan[numberOfGets:]...)
	//
	//return job.Plan
	return []atc.Plan{}
}

func (p pipeline) runJob(task manifest.Run, man manifest.Manifest, isDockerCompose bool, basePath string) atc.JobConfig {
	taskInputs := func() []atc.TaskInputConfig {
		inputs := []atc.TaskInputConfig{{Name: manifest.GitTrigger{}.GetTriggerName()}}
		if task.RestoreArtifacts {
			inputs = append(inputs, atc.TaskInputConfig{Name: artifactsName})
		}

		if man.FeatureToggles.Versioned() {
			inputs = append(inputs, atc.TaskInputConfig{Name: versionName})
		}
		return inputs
	}

	taskOutputs := func() []atc.TaskOutputConfig {
		var outputs []atc.TaskOutputConfig
		if len(task.SaveArtifacts) > 0 {
			outputs = append(outputs, atc.TaskOutputConfig{Name: artifactsOutDir})
		}

		if len(task.SaveArtifactsOnFailure) > 0 {
			outputs = append(outputs, atc.TaskOutputConfig{Name: artifactsOutDirOnFailure})
		}
		return outputs
	}

	jobConfig := atc.JobConfig{
		Name:   task.GetName(),
		Serial: true,
	}

	taskPath := "/bin/sh"
	if isDockerCompose {
		taskPath = "docker.sh"
	}

	taskEnv := make(atc.TaskEnv)
	for key, value := range task.Vars {
		taskEnv[key] = value
	}

	runStep := atc.Step{
		Config: &atc.RetryStep{
			Step: &atc.TimeoutStep{
				Step: &atc.TaskStep{
					Name:       restrictAllowedCharacterSet(task.GetName()),
					Privileged: task.Privileged,
					Config: &atc.TaskConfig{
						Platform:      "linux",
						Params:        taskEnv,
						ImageResource: p.imageResource(task.Docker),
						Run: atc.TaskRunConfig{
							Path: taskPath,
							Dir:  path.Join(gitDir, basePath),
							Args: runScriptArgs(task, man, !isDockerCompose, basePath),
						},
						Inputs:  taskInputs(),
						Outputs: taskOutputs(),
						Caches:  config.CacheDirs,
					},
				},
				Duration: task.GetTimeout(),
			},
			Attempts: task.GetAttempts(),
		},
	}

	jobConfig.PlanSequence = append(jobConfig.PlanSequence, runStep)

	if len(task.SaveArtifacts) > 0 {
		artifactPut := atc.Step{
			Config: &atc.RetryStep{
				Step: &atc.TimeoutStep{
					Step: &atc.PutStep{
						Name: artifactsName,
						Params: atc.Params{
							"folder":       artifactsOutDir,
							"version_file": path.Join(gitDir, ".git", "ref"),
						},
					},
					Duration: task.GetTimeout(),
				},
				Attempts: task.GetAttempts(),
			},
		}
		jobConfig.PlanSequence = append(jobConfig.PlanSequence, artifactPut)
	}

	return jobConfig
}

func (p pipeline) deployCFJob(task manifest.DeployCF, man manifest.Manifest, basePath string) atc.JobConfig {
	resourceName := deployCFResourceName(task)
	manifestPath := path.Join(gitDir, basePath, task.Manifest)
	vars := convertVars(task.Vars)
	//
	if strings.HasPrefix(task.Manifest, fmt.Sprintf("../%s/", artifactsInDir)) {
		manifestPath = strings.TrimPrefix(task.Manifest, "../")
	}

	appPath := path.Join(gitDir, basePath)
	if len(task.DeployArtifact) > 0 {
		appPath = path.Join(artifactsInDir, basePath, task.DeployArtifact)
	}

	job := atc.JobConfig{
		Name:   task.GetName(),
		Serial: true,
	}

	addTimeoutAndRetry := func(step atc.Step) atc.Step {
		return atc.Step{
			Config: &atc.TimeoutStep{
				Step: &atc.RetryStep{
					Step:     step.Config,
					Attempts: task.GetAttempts(),
				},
				Duration: task.GetTimeout(),
			},
		}
	}

	var steps []atc.Step
	if !task.Rolling {
		steps = append(steps, addTimeoutAndRetry(p.pushCandidateApp(task, resourceName, manifestPath, appPath, vars, man)))
		steps = append(steps, addTimeoutAndRetry(p.checkApp(task, resourceName, manifestPath)))
		steps = append(steps, p.prePromoteTasks(task, man, basePath)...)
		steps = append(steps, addTimeoutAndRetry(p.promoteCandidateAppToLive(task, resourceName, manifestPath)))
		job.Ensure = p.cleanupOldApps(task, resourceName, manifestPath)
	} else {
		if len(task.PrePromote) == 0 {
			steps = append(steps, addTimeoutAndRetry(p.pushAppRolling(task, resourceName, manifestPath, appPath, vars, man)))
		} else {
			steps = append(steps, addTimeoutAndRetry(p.pushCandidateApp(task, resourceName, manifestPath, appPath, vars, man)))
			steps = append(steps, p.prePromoteTasks(task, man, basePath)...)
			steps = append(steps, addTimeoutAndRetry(p.pushAppRolling(task, resourceName, manifestPath, appPath, vars, man)))
			steps = append(steps, addTimeoutAndRetry(p.removeTestApp(task, resourceName, manifestPath)))
		}
	}

	job.PlanSequence = steps
	return job
}

func (p pipeline) cleanupOldApps(task manifest.DeployCF, resourceName string, manifestPath string) *atc.Step {
	cleanup := atc.PutStep{
		Name:     "halfpipe-cleanup",
		Resource: resourceName,
		Params: atc.Params{
			"command":      "halfpipe-cleanup",
			"manifestPath": manifestPath,
			"cliVersion":   task.CliVersion,
		},
	}
	if task.Timeout != "" {
		cleanup.Params["timeout"] = task.Timeout
	}

	return &atc.Step{
		Config: &atc.RetryStep{
			Step: &atc.TimeoutStep{
				Step:     &cleanup,
				Duration: task.GetTimeout(),
			},
			Attempts: task.GetAttempts(),
		},
	}
}

func (p pipeline) promoteCandidateAppToLive(task manifest.DeployCF, resourceName string, manifestPath string) atc.Step {
	promote := atc.PutStep{
		Name:     "halfpipe-promote",
		Resource: resourceName,
		Params: atc.Params{
			"command":      "halfpipe-promote",
			"testDomain":   task.TestDomain,
			"manifestPath": manifestPath,
			"cliVersion":   task.CliVersion,
		},
	}
	if task.Timeout != "" {
		promote.Params["timeout"] = task.Timeout
	}
	return atc.Step{
		Config: &promote,
	}
}

func (p pipeline) checkApp(task manifest.DeployCF, resourceName string, manifestPath string) atc.Step {
	check := atc.PutStep{
		Name:     "halfpipe-check",
		Resource: resourceName,
		Params: atc.Params{
			"command":      "halfpipe-check",
			"manifestPath": manifestPath,
			"cliVersion":   task.CliVersion,
		},
	}
	if task.Timeout != "" {
		check.Params["timeout"] = task.Timeout
	}
	return atc.Step{
		Config: &check,
	}
}

func (p pipeline) pushCandidateApp(task manifest.DeployCF, resourceName string, manifestPath string, appPath string, vars map[string]interface{}, man manifest.Manifest) atc.Step {
	push := atc.PutStep{
		Name:     "halfpipe-push",
		Resource: resourceName,
		Params: atc.Params{
			"command":      "halfpipe-push",
			"testDomain":   task.TestDomain,
			"manifestPath": manifestPath,
			"gitRefPath":   path.Join(gitDir, ".git", "ref"),
			"cliVersion":   task.CliVersion,
		},
	}

	if task.IsDockerPush {
		push.Params["dockerUsername"] = defaults.DefaultValues.DockerUsername
		push.Params["dockerPassword"] = defaults.DefaultValues.DockerPassword
		if task.DockerTag != "" {
			if task.DockerTag == "version" {
				push.Params["dockerTag"] = path.Join(versionName, "version")
			} else if task.DockerTag == "gitref" {
				push.Params["dockerTag"] = path.Join(gitDir, ".git", "ref")
			}
		}
	} else {
		push.Params["appPath"] = appPath
	}

	if len(vars) > 0 {
		push.Params["vars"] = vars
	}
	if task.Timeout != "" {
		push.Params["timeout"] = task.Timeout
	}
	if len(task.PreStart) > 0 {
		push.Params["preStartCommand"] = strings.Join(task.PreStart, "; ")
	}
	if man.FeatureToggles.Versioned() {
		push.Params["buildVersionPath"] = path.Join("version", "version")
	}

	if task.Rolling {
		push.Name = "deploy-test-app"
		push.Params["instances"] = 1
		push.Params["instances"] = 1
	}

	return atc.Step{
		Config: &push,
	}
}

func (p pipeline) removeTestApp(task manifest.DeployCF, resourceName string, manifestPath string) atc.Step {
	return atc.Step{
		Config: &atc.PutStep{
			Name:     "remove-test-app",
			Resource: resourceName,
			Params: atc.Params{
				"command":      "halfpipe-delete-test",
				"manifestPath": manifestPath,
			},
		},
	}
}

func (p pipeline) pushAppRolling(task manifest.DeployCF, resourceName string, manifestPath string, appPath string, vars map[string]interface{}, man manifest.Manifest) atc.Step {
	deploy := atc.PutStep{
		Name:     "rolling-deploy",
		Resource: resourceName,
		Params: atc.Params{
			"command":      "halfpipe-rolling-deploy",
			"manifestPath": manifestPath,
			"gitRefPath":   path.Join(gitDir, ".git", "ref"),
			"cliVersion":   "cf7",
		},
	}

	if task.IsDockerPush {
		deploy.Params["dockerUsername"] = defaults.DefaultValues.DockerUsername
		deploy.Params["dockerPassword"] = defaults.DefaultValues.DockerPassword
		if task.DockerTag != "" {
			if task.DockerTag == "version" {
				deploy.Params["dockerTag"] = path.Join(versionName, "version")
			} else if task.DockerTag == "gitref" {
				deploy.Params["dockerTag"] = path.Join(gitDir, ".git", "ref")
			}
		}
	} else {
		deploy.Params["appPath"] = appPath
	}

	if len(vars) > 0 {
		deploy.Params["vars"] = vars
	}
	if task.Timeout != "" {
		deploy.Params["timeout"] = task.Timeout
	}
	if man.FeatureToggles.Versioned() {
		deploy.Params["buildVersionPath"] = path.Join("version", "version")
	}

	return atc.Step{
		Config: &deploy,
	}
}

func (p pipeline) prePromoteTasks(task manifest.DeployCF, man manifest.Manifest, basePath string) []atc.Step {
	// saveArtifacts and restoreArtifacts are needed to make sure we don't run pre-promote
	// tasks in parallel when the first task saves an artifact and the second restores it.

	var prePromoteTasks []atc.Step
	for _, t := range task.PrePromote {
		applications, e := p.readCfManifest(task.Manifest, nil, nil)
		if e != nil {
			panic(fmt.Sprintf("Failed to read manifest at path: %s\n\n%s", task.Manifest, e))
		}
		testRoute := buildTestRoute(applications[0].Name, task.Space, task.TestDomain)
		var ppJob atc.JobConfig
		switch ppTask := t.(type) {
		case manifest.Run:
			if len(ppTask.Vars) == 0 {
				ppTask.Vars = make(map[string]string)
			}
			ppTask.Vars["TEST_ROUTE"] = testRoute
			ppJob = p.runJob(ppTask, man, false, basePath)
		case manifest.DockerCompose:
			if len(ppTask.Vars) == 0 {
				ppTask.Vars = make(map[string]string)
			}
			ppTask.Vars["TEST_ROUTE"] = testRoute
			ppJob = p.dockerComposeJob(ppTask, man, basePath)

		case manifest.ConsumerIntegrationTest:
			if ppTask.ProviderHost == "" {
				ppTask.ProviderHost = testRoute
			}
			ppJob = p.consumerIntegrationTestJob(ppTask, man, basePath)
		}
		prePromoteTasks = append(prePromoteTasks, ppJob.PlanSequence...)
	}

	if len(prePromoteTasks) == 0 {
		return []atc.Step{}
	}

	var doSteps []atc.Step
	for _, ppTask := range prePromoteTasks {

		doSteps = append(doSteps, atc.Step{
			Config: &atc.DoStep{
				Steps: []atc.Step{ppTask},
			},
		})
	}
	if len(prePromoteTasks) > 1 {
		return []atc.Step{
			{
				Config: &atc.TimeoutStep{
					Step: &atc.InParallelStep{
						Config: atc.InParallelConfig{
							Steps:    doSteps,
							FailFast: true,
						},
					},
					Duration: task.GetTimeout(),
				},
			},
		}
	}

	return []atc.Step{
		{
			Config: &atc.TimeoutStep{
				Step: &atc.InParallelStep{
					Config: atc.InParallelConfig{
						Steps:    doSteps,
						FailFast: true,
					},
				},
				Duration: task.GetTimeout(),
			},
		},
	}
}

func buildTestRoute(appName, space, testDomain string) string {
	return fmt.Sprintf("%s-%s-CANDIDATE.%s",
		strings.Replace(appName, "_", "-", -1),
		strings.Replace(space, "_", "-", -1),
		testDomain)
}

func dockerComposeToRunTask(task manifest.DockerCompose, man manifest.Manifest) manifest.Run {
	if task.Vars == nil {
		task.Vars = make(map[string]string)
	}
	task.Vars["GCR_PRIVATE_KEY"] = "((halfpipe-gcr.private_key))"
	task.Vars["HALFPIPE_CACHE_TEAM"] = man.Team

	return manifest.Run{
		Retries: task.Retries,
		Name:    task.GetName(),
		Script:  dockerComposeScript(task, man.FeatureToggles.Versioned()),
		Docker: manifest.Docker{
			Image:    config.DockerRegistry + config.DockerComposeImage,
			Username: "_json_key",
			Password: "((halfpipe-gcr.private_key))",
		},
		Privileged:             true,
		Vars:                   task.Vars,
		SaveArtifacts:          task.SaveArtifacts,
		RestoreArtifacts:       task.RestoreArtifacts,
		SaveArtifactsOnFailure: task.SaveArtifactsOnFailure,
		Timeout:                task.GetTimeout(),
	}
}

func (p pipeline) dockerComposeJob(task manifest.DockerCompose, man manifest.Manifest, basePath string) atc.JobConfig {
	return p.runJob(dockerComposeToRunTask(task, man), man, true, basePath)
}

func dockerPushJobWithoutRestoreArtifacts(task manifest.DockerPush, resourceName string, basePath string) atc.JobConfig {
	put := atc.Step{
		Config: &atc.TimeoutStep{
			Step: &atc.RetryStep{
				Step: &atc.PutStep{
					Name: resourceName,
					Params: atc.Params{
						"build":         path.Join(gitDir, basePath, task.BuildPath),
						"dockerfile":    path.Join(gitDir, basePath, task.DockerfilePath),
						"tag_as_latest": true,
						"tag_file":      task.GetTagPath(),
						"build_args":    convertVars(task.Vars),
					},
				},
				Attempts: task.GetAttempts(),
			},
			Duration: task.GetTimeout(),
		},
	}

	return atc.JobConfig{
		Name:         task.GetName(),
		Serial:       true,
		PlanSequence: []atc.Step{put},
	}
}

func dockerPushJobWithRestoreArtifacts(task manifest.DockerPush, resourceName string, basePath string) atc.JobConfig {
	copyArtifact := atc.Step{
		Config: &atc.TimeoutStep{
			Step: &atc.TaskStep{
				Name: "copying-git-repo-and-artifacts-to-a-temporary-build-dir",
				Config: &atc.TaskConfig{
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
						{Name: gitDir},
						{Name: artifactsName},
					},
					Outputs: []atc.TaskOutputConfig{
						{Name: dockerBuildTmpDir},
					},
				},
			},
			Duration: task.GetTimeout(),
		},
	}

	put := atc.Step{
		Config: &atc.TimeoutStep{
			Step: &atc.RetryStep{
				Step: &atc.PutStep{
					Name: resourceName,
					Params: atc.Params{
						"build":         path.Join(dockerBuildTmpDir, basePath, task.BuildPath),
						"dockerfile":    path.Join(dockerBuildTmpDir, basePath, task.DockerfilePath),
						"tag_as_latest": true,
						"tag_file":      task.GetTagPath(),
						"build_args":    convertVars(task.Vars),
					},
				},
				Attempts: task.GetAttempts(),
			},
			Duration: task.GetTimeout(),
		},
	}

	return atc.JobConfig{
		Name:         task.GetName(),
		Serial:       true,
		PlanSequence: []atc.Step{copyArtifact, put},
	}
}

func (p pipeline) dockerPushJob(task manifest.DockerPush, basePath string) atc.JobConfig {
	resourceName := manifest.DockerTrigger{Image: task.Image}.GetTriggerName()
	if task.RestoreArtifacts {
		return dockerPushJobWithRestoreArtifacts(task, resourceName, basePath)
	}
	return dockerPushJobWithoutRestoreArtifacts(task, resourceName, basePath)
}

func pathToArtifactsDir(repoName string, basePath string, artifactsDir string) (artifactPath string) {
	fullPath := path.Join(repoName, basePath)
	numberOfParentsToConcourseRoot := len(strings.Split(fullPath, "/"))

	for i := 0; i < numberOfParentsToConcourseRoot; i++ {
		artifactPath = path.Join(artifactPath, "../")
	}

	return path.Join(artifactPath, artifactsDir)
}

func fullPathToArtifactsDir(repoName string, basePath string, artifactsDir string, artifactPath string) (fullArtifactPath string) {
	artifactPath = strings.TrimRight(artifactPath, "/")
	fullArtifactPath = path.Join(pathToArtifactsDir(repoName, basePath, artifactsDir), basePath)

	if subfolderPath := path.Dir(artifactPath); subfolderPath != "." {
		fullArtifactPath = path.Join(fullArtifactPath, subfolderPath)
	}

	return fullArtifactPath
}

func relativePathToRepoRoot(repoName string, basePath string) (relativePath string) {
	relativePath, _ = filepath.Rel(path.Join(repoName, basePath), repoName)
	return relativePath
}

func pathToGitRef(repoName string, basePath string) (gitRefPath string) {
	p := path.Join(relativePathToRepoRoot(repoName, basePath), ".git", "ref")
	return windowsToLinuxPath(p)

}

func pathToVersionFile(repoName string, basePath string) (gitRefPath string) {
	p := path.Join(relativePathToRepoRoot(repoName, basePath), path.Join("..", "version", "version"))
	return windowsToLinuxPath(p)
}

func windowsToLinuxPath(path string) (unixPath string) {
	return strings.Replace(path, `\`, "/", -1)
}

func dockerComposeScript(task manifest.DockerCompose, versioningEnabled bool) string {
	envStrings := []string{"-e GIT_REVISION"}
	for key := range task.Vars {
		if key == "GCR_PRIVATE_KEY" {
			continue
		}
		envStrings = append(envStrings, fmt.Sprintf("-e %s", key))
	}
	if versioningEnabled {
		envStrings = append(envStrings, "-e BUILD_VERSION")
	}
	sort.Strings(envStrings)

	var cacheVolumeFlags []string
	for _, cacheVolume := range config.DockerComposeCacheDirs {
		cacheVolumeFlags = append(cacheVolumeFlags, fmt.Sprintf("-v %s:%s", cacheVolume, cacheVolume))
	}

	composeFileOption := ""
	if task.ComposeFile != "docker-compose.yml" {
		composeFileOption = "-f " + task.ComposeFile
	}
	envOption := strings.Join(envStrings, " ")
	volumeOption := strings.Join(cacheVolumeFlags, " ")

	composeCommand := fmt.Sprintf("docker-compose %s run %s %s %s",
		composeFileOption,
		envOption,
		volumeOption,
		task.Service,
	)

	if task.Command != "" {
		composeCommand = fmt.Sprintf("%s %s", composeCommand, task.Command)
	}

	return fmt.Sprintf(`\docker login -u _json_key -p "$GCR_PRIVATE_KEY" https://eu.gcr.io
%s
`, composeCommand)
}

var warningMissingBash = `if ! which bash > /dev/null && [ "$SUPPRESS_BASH_WARNING" != "true" ]; then
  echo "WARNING: Bash is not present in the docker image"
  echo "If your script depends on bash you will get a strange error message like:"
  echo "  sh: yourscript.sh: command not found"
  echo "To fix, make sure your docker image contains bash!"
  echo "Or if you are sure you don't need bash you can suppress this warning by setting the environment variable \"SUPPRESS_BASH_WARNING\" to \"true\"."
  echo ""
  echo ""
fi
`

var warningAlpineImage = `if [ -e /etc/alpine-release ]
then
  echo "WARNING: you are running your build in a Alpine image or one that is based on the Alpine"
  echo "There is a known issue where DNS resolving does not work as expected"
  echo "https://github.com/gliderlabs/docker-alpine/issues/255"
  echo "If you see any errors related to resolving hostnames the best course of action is to switch to another image"
  echo "we recommend debian:buster-slim as an alternative"
  echo ""
  echo ""
fi
`

func runScriptArgs(task manifest.Run, man manifest.Manifest, checkForBash bool, basePath string) []string {

	script := task.Script
	if !strings.HasPrefix(script, "./") && !strings.HasPrefix(script, "/") && !strings.HasPrefix(script, `\`) {
		script = "./" + script
	}

	var out []string

	if checkForBash {
		out = append(out, warningMissingBash)
	}

	out = append(out, warningAlpineImage)
	if len(task.SaveArtifacts) != 0 || len(task.SaveArtifactsOnFailure) != 0 {
		out = append(out, `copyArtifact() {
  ARTIFACT=$1
  ARTIFACT_OUT_PATH=$2

  if [ -e $ARTIFACT ] ; then
    mkdir -p $ARTIFACT_OUT_PATH
    cp -r $ARTIFACT $ARTIFACT_OUT_PATH
  else
    echo "ERROR: Artifact '$ARTIFACT' not found. Try fly hijack to check the filesystem."
    exit 1
  fi
}
`)
	}

	if task.RestoreArtifacts {
		out = append(out, fmt.Sprintf("# Copying in artifacts from previous task"))
		out = append(out, fmt.Sprintf("cp -r %s/. %s\n", pathToArtifactsDir(gitDir, basePath, artifactsInDir), relativePathToRepoRoot(gitDir, basePath)))
	}

	out = append(out,
		fmt.Sprintf("export GIT_REVISION=`cat %s`", pathToGitRef(gitDir, basePath)),
	)

	if man.FeatureToggles.Versioned() {
		out = append(out,
			fmt.Sprintf("export BUILD_VERSION=`cat %s`", pathToVersionFile(gitDir, basePath)),
		)
	}

	scriptCall := fmt.Sprintf(`
%s
EXIT_STATUS=$?
if [ $EXIT_STATUS != 0 ] ; then
%s
fi
`, script, onErrorScript(task.SaveArtifactsOnFailure, basePath))
	out = append(out, scriptCall)

	if len(task.SaveArtifacts) != 0 {
		out = append(out, "# Artifacts to copy from task")
	}
	for _, artifactPath := range task.SaveArtifacts {
		out = append(out, fmt.Sprintf("copyArtifact %s %s", artifactPath, fullPathToArtifactsDir(gitDir, basePath, artifactsOutDir, artifactPath)))
	}
	return []string{"-c", strings.Join(out, "\n")}
}

func onErrorScript(artifactPaths []string, basePath string) string {
	var returnScript []string
	if len(artifactPaths) != 0 {
		returnScript = append(returnScript, "  # Artifacts to copy in case of failure")
	}
	for _, artifactPath := range artifactPaths {
		returnScript = append(returnScript, fmt.Sprintf("  copyArtifact %s %s", artifactPath, fullPathToArtifactsDir(gitDir, basePath, artifactsOutDirOnFailure, artifactPath)))

	}
	returnScript = append(returnScript, "  exit 1")
	return strings.Join(returnScript, "\n")
}

func restrictAllowedCharacterSet(in string) string {
	// https://concourse-ci.org/config-basics.html#schema.identifier
	simplified := regexp.MustCompile("[^a-z0-9-.]+").ReplaceAllString(strings.ToLower(in), " ")
	return strings.Replace(strings.TrimSpace(simplified), " ", "-", -1)
}
