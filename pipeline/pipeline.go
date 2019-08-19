package pipeline

import (
	"fmt"
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

const gitName = "git"
const gitDir = "git"

const dockerBuildTmpDir = "docker_build"

const versionName = "version"

const cronName = "cron"

const updateJobName = "update"
const updatePipelineName = "halfpipe update"
const updateTaskAttempts = 2

func restoreArtifactTask(man manifest.Manifest) atc.PlanConfig {
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
				Name: gitName,
			},
		},
		Outputs: []atc.TaskOutputConfig{
			{
				Name: artifactsInDir,
			},
		},
	}

	return atc.PlanConfig{
		Task:       "get artifact",
		TaskConfig: &config,
		Attempts:   2,
	}
}

func (p pipeline) initialPlan(man manifest.Manifest, task manifest.Task) []atc.PlanConfig {
	_, isUpdateTask := task.(manifest.Update)
	versioningEnabled := man.FeatureToggles.Versioned()

	var plan []atc.PlanConfig
	for _, trigger := range man.Triggers {
		switch trigger := trigger.(type) {
		case manifest.Git:
			gitClone := atc.PlanConfig{Get: trigger.GetTriggerName()}
			if trigger.Shallow {
				gitClone.Params = map[string]interface{}{
					"depth": 1,
				}
			}
			plan = append(plan, gitClone)
		case manifest.Cron:
			if isUpdateTask || !versioningEnabled {
				plan = append(plan, atc.PlanConfig{Get: trigger.GetTriggerName()})
			}
		}
	}

	if !isUpdateTask && man.FeatureToggles.Versioned() {
		plan = append(plan, atc.PlanConfig{Get: versionName})
	}

	if task.ReadsFromArtifacts() {
		plan = append(plan, restoreArtifactTask(man))
	}

	return plan
}

func (p pipeline) dockerPushResources(tasks manifest.TaskList) (resourceConfigs atc.ResourceConfigs) {
	for _, task := range tasks {
		switch task := task.(type) {
		case manifest.DockerPush:
			resourceConfigs = append(resourceConfigs, p.dockerPushResource(task))
		case manifest.Parallel:
			for _, subTask := range task.Tasks {
				switch subTask := subTask.(type) {
				case manifest.DockerPush:
					resourceConfigs = append(resourceConfigs, p.dockerPushResource(subTask))
				}
			}
		}
	}

	return
}

func (p pipeline) cfPushResources(tasks manifest.TaskList) (resourceType atc.ResourceTypes, resourceConfigs atc.ResourceConfigs) {
	for _, task := range tasks {
		switch task := task.(type) {
		case manifest.DeployCF:
			resourceName := deployCFResourceName(task)
			if _, found := resourceConfigs.Lookup(resourceName); !found {
				resourceConfigs = append(resourceConfigs, p.deployCFResource(task, resourceName))
			}
		case manifest.Parallel:
			for _, subTask := range task.Tasks {
				switch subTask := subTask.(type) {
				case manifest.DeployCF:
					resourceName := deployCFResourceName(subTask)
					if _, found := resourceConfigs.Lookup(resourceName); !found {
						resourceConfigs = append(resourceConfigs, p.deployCFResource(subTask, resourceName))
					}
				}
			}
		}
	}

	if len(resourceConfigs) > 0 {
		resourceType = append(resourceType, halfpipeCfDeployResourceType())
	}

	return
}

func (p pipeline) resourceConfigs(man manifest.Manifest) (resourceTypes atc.ResourceTypes, resourceConfigs atc.ResourceConfigs) {
	for _, trigger := range man.Triggers {
		switch trigger := trigger.(type) {
		case manifest.Git:
			resourceConfigs = append(resourceConfigs, p.gitResource(trigger))
		case manifest.Cron:
			resourceTypes = append(resourceTypes, cronResourceType())
			resourceConfigs = append(resourceConfigs, p.cronResource(trigger))
		}
	}

	if man.NotifiesOnFailure() || man.Tasks.NotifiesOnSuccess() {
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

	cfResourceTypes, cfResources := p.cfPushResources(man.Tasks)
	resourceTypes = append(resourceTypes, cfResourceTypes...)
	resourceConfigs = append(resourceConfigs, cfResources...)

	return
}

func (p pipeline) taskToJobs(task manifest.Task, man manifest.Manifest, previousTaskNames []string) (job *atc.JobConfig) {
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
		job = p.dockerPushJob(task, man, basePath)

	case manifest.ConsumerIntegrationTest:
		job = p.consumerIntegrationTestJob(task, man, basePath)

	case manifest.DeployMLZip:
		runTask := ConvertDeployMLZipToRunTask(task, man)
		job = p.runJob(runTask, man, false, basePath)

	case manifest.DeployMLModules:
		runTask := ConvertDeployMLModulesToRunTask(task, man)
		job = p.runJob(runTask, man, false, basePath)
	case manifest.Update:
		job = p.updateJobConfig(man, basePath)
	}

	if task.SavesArtifactsOnFailure() || man.NotifiesOnFailure() {
		sequence := atc.PlanSequence{}

		if task.SavesArtifactsOnFailure() {
			sequence = append(sequence, saveArtifactOnFailurePlan())
		}
		if man.NotifiesOnFailure() {
			sequence = append(sequence, slackOnFailurePlan(man.SlackChannel))
		}

		job.Failure = &atc.PlanConfig{
			InParallel: &atc.InParallelConfig{
				Steps: sequence,
			},
		}
	}

	if task.NotifiesOnSuccess() {
		sequence := atc.PlanSequence{
			slackOnSuccessPlan(man.SlackChannel),
		}
		job.Success = &atc.PlanConfig{
			InParallel: &atc.InParallelConfig{
				Steps: sequence,
			},
		}
	}

	job.Plan = append(initialPlan, job.Plan...)
	job.Plan = inParallelGets(job)

	configureTriggerOnGets(job, task, man.FeatureToggles.Versioned())
	addTimeout(job, task.GetTimeout())
	addPassedJobsToGets(job, previousTaskNames)

	return
}

func (p pipeline) taskNamesFromTask(task manifest.Task) (taskNames []string) {
	switch task := task.(type) {
	case manifest.Parallel:
		for _, subTask := range task.Tasks {
			taskNames = append(taskNames, p.taskNamesFromTask(subTask)...)
		}
	default:
		taskNames = append(taskNames, task.GetName())
	}
	return
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
				cfg.Jobs = append(cfg.Jobs, *p.taskToJobs(subTask, man, p.previousTaskNames(i, man.Tasks)))
			}
		default:
			cfg.Jobs = append(cfg.Jobs, *p.taskToJobs(task, man, p.previousTaskNames(i, man.Tasks)))
		}
	}

	return
}

func addTimeout(job *atc.JobConfig, timeout string) {
	for i := range job.Plan {
		job.Plan[i].Timeout = timeout
	}

	if job.Ensure != nil {
		job.Ensure.Timeout = timeout
	}
}

func addPassedJobsToGets(job *atc.JobConfig, passedJobs []string) {
	if len(passedJobs) > 0 {
		inParallel := *job.Plan[0].InParallel
		for i, get := range inParallel.Steps {
			if get.Name() == gitName ||
				get.Name() == versionName ||
				get.Name() == cronName {
				inParallel.Steps[i].Passed = passedJobs
			}
		}
	}
}

func configureTriggerOnGets(job *atc.JobConfig, task manifest.Task, versioningEnabled bool) {
	inParallel := *job.Plan[0].InParallel
	switch task.(type) {
	case manifest.Update:
		for i := range inParallel.Steps {
			inParallel.Steps[i].Trigger = true
		}
	default:
		for i, get := range inParallel.Steps {
			if get.Get == versionName && !task.IsManualTrigger() {
				inParallel.Steps[i].Trigger = true
			} else {
				inParallel.Steps[i].Trigger = !task.IsManualTrigger() && !versioningEnabled
			}
		}
	}
}

func inParallelGets(job *atc.JobConfig) atc.PlanSequence {
	var numberOfGets int
	for i, plan := range job.Plan {
		if plan.Get == "" {
			numberOfGets = i
			break
		}
	}

	sequence := job.Plan[:numberOfGets]
	inParallelPlan := atc.PlanSequence{atc.PlanConfig{
		InParallel: &atc.InParallelConfig{
			Steps: sequence,
		},
	}}
	job.Plan = append(inParallelPlan, job.Plan[numberOfGets:]...)

	return job.Plan
}

func (p pipeline) runJob(task manifest.Run, man manifest.Manifest, isDockerCompose bool, basePath string) *atc.JobConfig {
	jobConfig := atc.JobConfig{
		Name:   task.Name,
		Serial: true,
		Plan:   atc.PlanSequence{},
	}

	taskPath := "/bin/sh"
	if isDockerCompose {
		taskPath = "docker.sh"
	}

	runPlan := atc.PlanConfig{
		Attempts:   task.GetAttempts(),
		Task:       task.Name,
		Privileged: task.Privileged,
		TaskConfig: &atc.TaskConfig{
			Platform:      "linux",
			Params:        task.Vars,
			ImageResource: p.imageResource(task.Docker),
			Run: atc.TaskRunConfig{
				Path: taskPath,
				Dir:  path.Join(gitDir, basePath),
				Args: runScriptArgs(task, man, !isDockerCompose, basePath),
			},
			Inputs: []atc.TaskInputConfig{
				{Name: gitName},
			},
			Caches: config.CacheDirs,
		}}

	if task.RestoreArtifacts {
		runPlan.TaskConfig.Inputs = append(runPlan.TaskConfig.Inputs, atc.TaskInputConfig{Name: artifactsName})
	}

	if man.FeatureToggles.Versioned() {
		runPlan.TaskConfig.Inputs = append(runPlan.TaskConfig.Inputs, atc.TaskInputConfig{Name: versionName})
	}

	jobConfig.Plan = append(jobConfig.Plan, runPlan)

	if len(task.SaveArtifacts) > 0 {
		jobConfig.Plan[0].TaskConfig.Outputs = append(jobConfig.Plan[0].TaskConfig.Outputs, atc.TaskOutputConfig{Name: artifactsOutDir})

		artifactPut := atc.PlanConfig{
			Put: artifactsName,
			Params: atc.Params{
				"folder":       artifactsOutDir,
				"version_file": path.Join(gitDir, ".git", "ref"),
			},
		}
		jobConfig.Plan = append(jobConfig.Plan, artifactPut)
	}

	if len(task.SaveArtifactsOnFailure) > 0 {
		jobConfig.Plan[0].TaskConfig.Outputs = append(jobConfig.Plan[0].TaskConfig.Outputs, atc.TaskOutputConfig{Name: artifactsOutDirOnFailure})
	}

	return &jobConfig
}

func (p pipeline) deployCFJob(task manifest.DeployCF, man manifest.Manifest, basePath string) *atc.JobConfig {
	resourceName := deployCFResourceName(task)
	manifestPath := path.Join(gitDir, basePath, task.Manifest)

	if strings.HasPrefix(task.Manifest, fmt.Sprintf("../%s/", artifactsInDir)) {
		manifestPath = strings.TrimPrefix(task.Manifest, "../")
	}

	vars := convertVars(task.Vars)

	appPath := path.Join(gitDir, basePath)
	if len(task.DeployArtifact) > 0 {
		appPath = path.Join(artifactsInDir, basePath, task.DeployArtifact)
	}

	job := atc.JobConfig{
		Name:   task.Name,
		Serial: true,
	}

	push := atc.PlanConfig{
		Put:      "cf halfpipe-push",
		Attempts: task.GetAttempts(),
		Resource: resourceName,
		Params: atc.Params{
			"command":      "halfpipe-push",
			"testDomain":   task.TestDomain,
			"manifestPath": manifestPath,
			"appPath":      appPath,
			"gitRefPath":   path.Join(gitDir, ".git", "ref"),
		},
	}
	if len(vars) > 0 {
		push.Params["vars"] = vars
	}
	if task.Timeout != "" {
		push.Params["timeout"] = task.Timeout
	}

	if man.FeatureToggles.Versioned() {
		push.Params["buildVersionPath"] = path.Join("version", "version")
	}

	job.Plan = append(job.Plan, push)

	check := atc.PlanConfig{
		Put:      "cf halfpipe-check",
		Attempts: task.GetAttempts(),
		Resource: resourceName,
		Params: atc.Params{
			"command":      "halfpipe-check",
			"manifestPath": manifestPath,
		},
	}
	if task.Timeout != "" {
		check.Params["timeout"] = task.Timeout
	}
	job.Plan = append(job.Plan, check)

	// saveArtifactInPP and restoreArtifactInPP are needed to make sure we don't run pre-promote tasks in parallel when the first task saves an artifact and the second restores it.
	var prePromoteTasks atc.PlanSequence
	var saveArtifactInPP bool
	var restoreArtifactInPP bool
	for _, t := range task.PrePromote {
		applications, e := p.readCfManifest(task.Manifest, nil, nil)
		if e != nil {
			panic(e)
		}
		testRoute := buildTestRoute(applications[0].Name, task.Space, task.TestDomain)
		var ppJob *atc.JobConfig
		switch ppTask := t.(type) {
		case manifest.Run:
			if len(ppTask.Vars) == 0 {
				ppTask.Vars = make(map[string]string)
			}
			ppTask.Vars["TEST_ROUTE"] = testRoute
			ppJob = p.runJob(ppTask, man, false, basePath)
			restoreArtifactInPP = saveArtifactInPP && ppTask.RestoreArtifacts
			saveArtifactInPP = saveArtifactInPP || len(ppTask.SaveArtifacts) > 0

		case manifest.DockerCompose:
			if len(ppTask.Vars) == 0 {
				ppTask.Vars = make(map[string]string)
			}
			ppTask.Vars["TEST_ROUTE"] = testRoute
			ppJob = p.dockerComposeJob(ppTask, man, basePath)
			restoreArtifactInPP = saveArtifactInPP && ppTask.RestoreArtifacts
			saveArtifactInPP = saveArtifactInPP || len(ppTask.SaveArtifacts) > 0

		case manifest.ConsumerIntegrationTest:
			if ppTask.ProviderHost == "" {
				ppTask.ProviderHost = testRoute
			}
			ppJob = p.consumerIntegrationTestJob(ppTask, man, basePath)
		}
		planConfig := atc.PlanConfig{Do: &ppJob.Plan}
		prePromoteTasks = append(prePromoteTasks, planConfig)
	}

	if len(prePromoteTasks) > 0 && !restoreArtifactInPP {
		inParallelJob := atc.PlanConfig{
			InParallel: &atc.InParallelConfig{
				Steps: prePromoteTasks,
			},
		}

		job.Plan = append(job.Plan, inParallelJob)
	} else if len(prePromoteTasks) > 0 {
		job.Plan = append(job.Plan, prePromoteTasks...)
	}

	promote := atc.PlanConfig{
		Put:      "cf halfpipe-promote",
		Attempts: task.GetAttempts(),
		Resource: resourceName,
		Params: atc.Params{
			"command":      "halfpipe-promote",
			"testDomain":   task.TestDomain,
			"manifestPath": manifestPath,
		},
	}
	if task.Timeout != "" {
		promote.Params["timeout"] = task.Timeout
	}
	job.Plan = append(job.Plan, promote)

	cleanup := atc.PlanConfig{
		Put:      "cf halfpipe-cleanup",
		Attempts: task.GetAttempts(),
		Resource: resourceName,
		Params: atc.Params{
			"command":      "halfpipe-cleanup",
			"manifestPath": manifestPath,
		},
	}
	if task.Timeout != "" {
		cleanup.Params["timeout"] = task.Timeout
	}

	job.Ensure = &cleanup
	return &job
}

func buildTestRoute(appName, space, testDomain string) string {
	return fmt.Sprintf("%s-%s-CANDIDATE.%s", appName, space, testDomain)
}

func dockerComposeToRunTask(task manifest.DockerCompose, man manifest.Manifest) manifest.Run {
	if task.Vars == nil {
		task.Vars = make(map[string]string)
	}
	task.Vars["GCR_PRIVATE_KEY"] = "((halfpipe-gcr.private_key))"
	task.Vars["HALFPIPE_CACHE_TEAM"] = man.Team

	return manifest.Run{
		Retries: task.Retries,
		Name:    task.Name,
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
	}
}

func (p pipeline) dockerComposeJob(task manifest.DockerCompose, man manifest.Manifest, basePath string) *atc.JobConfig {
	return p.runJob(dockerComposeToRunTask(task, man), man, true, basePath)
}

func dockerPushJobWithoutRestoreArtifacts(task manifest.DockerPush, resourceName string, man manifest.Manifest, basePath string) *atc.JobConfig {
	job := atc.JobConfig{
		Name:   task.Name,
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{
				Attempts: task.GetAttempts(),
				Put:      resourceName,
				Params: atc.Params{
					"build":         path.Join(gitDir, basePath, task.BuildPath),
					"dockerfile":    path.Join(gitDir, basePath, task.DockerfilePath),
					"tag_as_latest": true,
				}},
		},
	}
	if len(task.Vars) > 0 {
		job.Plan[0].Params["build_args"] = convertVars(task.Vars)
	}
	if man.FeatureToggles.Versioned() {
		job.Plan[0].Params["tag_file"] = "version/number"
	}
	return &job
}

func dockerPushJobWithRestoreArtifacts(task manifest.DockerPush, resourceName string, man manifest.Manifest, basePath string) *atc.JobConfig {
	job := atc.JobConfig{
		Name:   task.Name,
		Serial: true,
		Plan: atc.PlanSequence{
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
						{Name: gitDir},
						{Name: artifactsName},
					},
					Outputs: []atc.TaskOutputConfig{
						{Name: dockerBuildTmpDir},
					},
				}},

			atc.PlanConfig{
				Attempts: task.GetAttempts(),
				Put:      resourceName,
				Params: atc.Params{
					"build":         path.Join(dockerBuildTmpDir, basePath, task.BuildPath),
					"dockerfile":    path.Join(dockerBuildTmpDir, basePath, task.DockerfilePath),
					"tag_as_latest": true,
				}},
		},
	}

	putIndex := 1
	if len(task.Vars) > 0 {
		job.Plan[putIndex].Params["build_args"] = convertVars(task.Vars)
	}
	if man.FeatureToggles.Versioned() {
		job.Plan[putIndex].Params["tag_file"] = "version/number"
	}
	return &job
}

func (p pipeline) dockerPushJob(task manifest.DockerPush, man manifest.Manifest, basePath string) *atc.JobConfig {
	resourceName := dockerPushResourceName(task)
	if task.RestoreArtifacts {
		return dockerPushJobWithRestoreArtifacts(task, resourceName, man, basePath)
	}
	return dockerPushJobWithoutRestoreArtifacts(task, resourceName, man, basePath)
}

func pathToArtifactsDir(repoName string, basePath string, artifactsDir string) (artifactPath string) {
	fullPath := path.Join(repoName, basePath)
	numberOfParentsToConcourseRoot := len(strings.Split(fullPath, "/"))

	for i := 0; i < numberOfParentsToConcourseRoot; i++ {
		artifactPath = path.Join(artifactPath, "../")
	}

	artifactPath = path.Join(artifactPath, artifactsDir)
	return
}

func fullPathToArtifactsDir(repoName string, basePath string, artifactsDir string, artifactPath string) (fullArtifactPath string) {
	fullArtifactPath = path.Join(pathToArtifactsDir(repoName, basePath, artifactsDir), basePath)

	if subfolderPath := path.Dir(artifactPath); subfolderPath != "." {
		fullArtifactPath = path.Join(fullArtifactPath, subfolderPath)
	}

	return
}

func relativePathToRepoRoot(repoName string, basePath string) (relativePath string) {
	relativePath, _ = filepath.Rel(path.Join(repoName, basePath), repoName)
	return
}

func pathToGitRef(repoName string, basePath string) (gitRefPath string) {
	p := path.Join(relativePathToRepoRoot(repoName, basePath), ".git", "ref")
	gitRefPath = windowsToLinuxPath(p)
	return
}

func pathToVersionFile(repoName string, basePath string) (gitRefPath string) {
	p := path.Join(relativePathToRepoRoot(repoName, basePath), path.Join("..", "version", "version"))
	gitRefPath = windowsToLinuxPath(p)
	return
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
	if task.ComposeFile != "" {
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

func (p pipeline) addCronResource(cfg *atc.Config, man manifest.Manifest) {
	if man.CronTrigger != "" {
	}
}

func runScriptArgs(task manifest.Run, man manifest.Manifest, checkForBash bool, basePath string) []string {

	script := task.Script
	if !strings.HasPrefix(script, "./") && !strings.HasPrefix(script, "/") && !strings.HasPrefix(script, `\`) {
		script = "./" + script
	}

	var out []string

	if checkForBash {
		out = append(out, `which bash > /dev/null
if [ $? != 0 ]; then
  echo "WARNING: Bash is not present in the docker image"
  echo "If your script depends on bash you will get a strange error message like:"
  echo "  sh: yourscript.sh: command not found"
  echo "To fix, make sure your docker image contains bash!"
  echo ""
  echo ""
fi
`)
	}

	out = append(out, `if [ -e /etc/alpine-release ]
then
  echo "WARNING: you are running your build in a Alpine image or one that is based on the Alpine"
  echo "There is a known issue where DNS resolving does not work as expected"
  echo "https://github.com/gliderlabs/docker-alpine/issues/255"
  echo "If you see any errors related to resolving hostnames the best course of action is to switch to another image"
  echo "we recommend debian:stretch-slim as an alternative"
  echo ""
  echo ""
fi
`)
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
