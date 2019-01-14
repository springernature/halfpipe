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
	"github.com/concourse/atc"
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

const dockerPushResourceName = "Docker Registry"
const dockerBuildTmpDir = "docker_build"

const versionName = "version"

const cronName = "cron"
const timerName = "timer"

const updateJobName = "update"
const updatePipelineName = "halfpipe update"

func (p pipeline) addSlackResourceTypeAndResource(cfg *atc.Config) {
	slackResourceType := p.slackResourceType()
	cfg.ResourceTypes = append(cfg.ResourceTypes, slackResourceType)

	slackResource := p.slackResource()
	cfg.Resources = append(cfg.Resources, slackResource)
}

func restoreArtifactTask(man manifest.Manifest) atc.PlanConfig {
	// This function is used in pipeline.artifactResource for some reason to lowercase
	// and remove chars that are not part of the regex in the folder in the config..
	// So we must reuse it.
	filter := func(str string) string {
		reg := regexp.MustCompile(`[^a-z0-9\-]+`)
		return reg.ReplaceAllString(strings.ToLower(str), "")
	}

	JSON_KEY := "((gcr.private_key))"
	if man.ArtifactConfig.JsonKey != "" {
		JSON_KEY = man.ArtifactConfig.JsonKey
	}

	BUCKET := "halfpipe-io-artifacts"
	if man.ArtifactConfig.Bucket != "" {
		BUCKET = man.ArtifactConfig.Bucket
	}

	config := atc.TaskConfig{
		Platform:  "linux",
		RootfsURI: "",
		ImageResource: &atc.ImageResource{
			Type: "docker-image",
			Source: atc.Source{
				"repository": "platformengineering/gcp-resource",
				"tag":        "stable",
			},
		},
		Params: map[string]string{
			"BUCKET":       BUCKET,
			"FOLDER":       path.Join(filter(man.Team), filter(man.Pipeline)),
			"JSON_KEY":     JSON_KEY,
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
	}
}

func (p pipeline) initialPlan(cfg *atc.Config, man manifest.Manifest, includeVersion bool, task manifest.Task) []atc.PlanConfig {
	gitClone := atc.PlanConfig{Get: gitName}
	if man.Repo.Shallow {
		gitClone.Params = map[string]interface{}{
			"depth": 1,
		}
	}

	initialPlan := []atc.PlanConfig{gitClone}

	if includeVersion {
		initialPlan = append(initialPlan, atc.PlanConfig{Get: versionName})
		if task != nil && task.ReadsFromArtifacts() {
			initialPlan = append(initialPlan, restoreArtifactTask(man))
		}
	} else {
		if man.CronTrigger != "" {
			initialPlan = append(initialPlan, atc.PlanConfig{Get: cronName})
		}
		if man.TriggerInterval != "" {
			initialPlan = append(initialPlan, atc.PlanConfig{Get: timerName})
		}
		if task != nil && task.ReadsFromArtifacts() {
			initialPlan = append(initialPlan, restoreArtifactTask(man))
		}
	}

	return initialPlan
}

func (p pipeline) addCfResourceType(cfg *atc.Config) {
	resTypeName := "cf-resource"
	if _, exists := cfg.ResourceTypes.Lookup(resTypeName); !exists {
		cfg.ResourceTypes = append(cfg.ResourceTypes, halfpipeCfDeployResourceType(resTypeName))
	}
}

func (p pipeline) Render(man manifest.Manifest) (cfg atc.Config) {
	cfg.Resources = append(cfg.Resources, p.gitResource(man.Repo))

	if man.NotifiesOnFailure() || man.Tasks.NotifiesOnSuccess() {
		p.addSlackResourceTypeAndResource(&cfg)
	}

	p.addArtifactResource(&cfg, man)

	if man.CronTrigger != "" {
		p.addCronResource(&cfg, man)
	}

	if man.TriggerInterval != "" {
		p.addTriggerResource(&cfg, man)
	}

	var parallelTasks []string
	var taskBeforeParallelTasks string
	if len(cfg.Jobs) > 0 {
		taskBeforeParallelTasks = cfg.Jobs[len(cfg.Jobs)-1].Name
	}

	if man.FeatureToggles.Versioned() {
		cfg.Resources = append(cfg.Resources, p.versionResource(man))
		job := p.updateJob(man)
		if man.NotifiesOnFailure() || man.Tasks.NotifiesOnSuccess() {
			failurePlan := slackOnFailurePlan(man.SlackChannel)
			job.Failure = &failurePlan
		}

		job.Plan = append(p.initialPlan(&cfg, man, false, nil), job.Plan...)
		job.Plan = aggregateGets(&job)

		aggregate := *job.Plan[0].Aggregate
		for i := range aggregate {
			aggregate[i].Trigger = true
		}

		cfg.Jobs = append(cfg.Jobs, job)
	}

	for _, t := range man.Tasks {
		initialPlan := p.initialPlan(&cfg, man, man.FeatureToggles.Versioned(), t)

		var job *atc.JobConfig
		var parallel bool
		switch task := t.(type) {
		case manifest.Run:
			task.Name = uniqueName(&cfg, task.Name, fmt.Sprintf("run %s", strings.Replace(task.Script, "./", "", 1)))
			job = p.runJob(task, man, false)
			parallel = task.Parallel

		case manifest.DockerCompose:
			task.Name = uniqueName(&cfg, task.Name, "docker-compose")
			job = p.dockerComposeJob(task, man)
			parallel = task.Parallel

		case manifest.DeployCF:
			p.addCfResourceType(&cfg)
			resourceName := uniqueName(&cfg, deployCFResourceName(task), "")
			task.Name = uniqueName(&cfg, task.Name, "deploy-cf")
			cfg.Resources = append(cfg.Resources, p.deployCFResource(task, resourceName))
			job = p.deployCFJob(task, resourceName, man, &cfg)
			parallel = task.Parallel

		case manifest.DockerPush:
			resourceName := uniqueName(&cfg, dockerPushResourceName, "")
			task.Name = uniqueName(&cfg, task.Name, "docker-push")
			cfg.Resources = append(cfg.Resources, p.dockerPushResource(task, resourceName))
			job = p.dockerPushJob(task, resourceName, man)
			parallel = task.Parallel

		case manifest.ConsumerIntegrationTest:
			task.Name = uniqueName(&cfg, task.Name, "consumer-integration-test")
			job = p.consumerIntegrationTestJob(task, man)
			parallel = task.Parallel

		case manifest.DeployMLZip:
			task.Name = uniqueName(&cfg, task.Name, "deploy-ml-zip")
			runTask := ConvertDeployMLZipToRunTask(task, man)
			job = p.runJob(runTask, man, false)
			parallel = task.Parallel

		case manifest.DeployMLModules:
			task.Name = uniqueName(&cfg, task.Name, "deploy-ml-modules")
			runTask := ConvertDeployMLModulesToRunTask(task, man)
			job = p.runJob(runTask, man, false)
			parallel = task.Parallel
		}

		if t.SavesArtifactsOnFailure() || man.NotifiesOnFailure() {
			sequence := atc.PlanSequence{}

			if t.SavesArtifactsOnFailure() {
				sequence = append(sequence, saveArtifactOnFailurePlan(man.Team, man.Pipeline))
			}
			if man.NotifiesOnFailure() {
				sequence = append(sequence, slackOnFailurePlan(man.SlackChannel))
			}

			job.Failure = &atc.PlanConfig{
				Aggregate: &sequence,
			}
		}

		if t.NotifiesOnSuccess() {
			sequence := atc.PlanSequence{
				slackOnSuccessPlan(man.SlackChannel),
			}
			job.Success = &atc.PlanConfig{
				Aggregate: &sequence,
			}
		}

		job.Plan = append(initialPlan, job.Plan...)
		job.Plan = aggregateGets(job)

		var passedJobNames []string
		if parallel {
			parallelTasks = append(parallelTasks, job.Name)
			if taskBeforeParallelTasks != "" {
				passedJobNames = []string{taskBeforeParallelTasks}
			}
		} else {
			taskBeforeParallelTasks = job.Name
			if len(parallelTasks) > 0 {
				passedJobNames = parallelTasks
				parallelTasks = []string{}
			} else {
				if len(cfg.Jobs) > 0 {
					passedJobNames = append(passedJobNames, cfg.Jobs[len(cfg.Jobs)-1].Name)
				}
			}
		}

		addPassedJobsToGets(job, passedJobNames)
		configureTriggerOnGets(job, t.IsManualTrigger(), man.FeatureToggles.Versioned())

		cfg.Jobs = append(cfg.Jobs, *job)
	}

	return
}

func addPassedJobsToGets(job *atc.JobConfig, passedJobs []string) {
	aggregate := *job.Plan[0].Aggregate
	for i, get := range aggregate {
		if get.Name() == gitName ||
			get.Name() == versionName ||
			get.Name() == timerName ||
			get.Name() == cronName {
			aggregate[i].Passed = passedJobs
		}
	}
}

func configureTriggerOnGets(job *atc.JobConfig, manualTrigger bool, versioningEnabled bool) {
	aggregate := *job.Plan[0].Aggregate
	for i, get := range aggregate {
		if get.Get == versionName && !manualTrigger {
			aggregate[i].Trigger = true
		} else if get.Get != artifactsName {
			aggregate[i].Trigger = !manualTrigger && !versioningEnabled
		}
	}
}

func aggregateGets(job *atc.JobConfig) atc.PlanSequence {
	var numberOfGets int
	for i, plan := range job.Plan {
		if plan.Get == "" {
			numberOfGets = i
			break
		}
	}

	sequence := job.Plan[:numberOfGets]
	aggregatePlan := atc.PlanSequence{atc.PlanConfig{Aggregate: &sequence}}
	job.Plan = append(aggregatePlan, job.Plan[numberOfGets:]...)

	return job.Plan
}

func (p pipeline) runJob(task manifest.Run, man manifest.Manifest, isDockerCompose bool) *atc.JobConfig {
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
		Privileged: isDockerCompose,
		TaskConfig: &atc.TaskConfig{
			Platform:      "linux",
			Params:        task.Vars,
			ImageResource: p.imageResource(task.Docker),
			Run: atc.TaskRunConfig{
				Path: taskPath,
				Dir:  path.Join(gitDir, man.Repo.BasePath),
				Args: runScriptArgs(task, man, !isDockerCompose),
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
			Put:      artifactsName,
			Resource: GenerateArtifactsResourceName(man.Team, man.Pipeline),
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

func (p pipeline) deployCFJob(task manifest.DeployCF, resourceName string, man manifest.Manifest, cfg *atc.Config) *atc.JobConfig {
	manifestPath := path.Join(gitDir, man.Repo.BasePath, task.Manifest)

	if strings.HasPrefix(task.Manifest, fmt.Sprintf("../%s/", artifactsInDir)) {
		manifestPath = strings.TrimPrefix(task.Manifest, "../")
	}

	vars := convertVars(task.Vars)

	appPath := path.Join(gitDir, man.Repo.BasePath)
	if len(task.DeployArtifact) > 0 {
		appPath = path.Join(artifactsInDir, task.DeployArtifact)
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
	job.Plan = append(job.Plan, push)

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
			ppTask.Name = uniqueName(cfg, ppTask.Name, fmt.Sprintf("run %s", strings.Replace(ppTask.Script, "./", "", 1)))
			ppJob = p.runJob(ppTask, man, false)
			restoreArtifactInPP = saveArtifactInPP && ppTask.RestoreArtifacts
			saveArtifactInPP = saveArtifactInPP || len(ppTask.SaveArtifacts) > 0

		case manifest.DockerCompose:
			if len(ppTask.Vars) == 0 {
				ppTask.Vars = make(map[string]string)
			}
			ppTask.Vars["TEST_ROUTE"] = testRoute
			ppTask.Name = uniqueName(cfg, ppTask.Name, "docker-compose")
			ppJob = p.dockerComposeJob(ppTask, man)
			restoreArtifactInPP = saveArtifactInPP && ppTask.RestoreArtifacts
			saveArtifactInPP = saveArtifactInPP || len(ppTask.SaveArtifacts) > 0

		case manifest.ConsumerIntegrationTest:
			ppTask.Name = uniqueName(cfg, ppTask.Name, "consumer-integration-test")
			if ppTask.ProviderHost == "" {
				ppTask.ProviderHost = testRoute
			}
			ppJob = p.consumerIntegrationTestJob(ppTask, man)
		}
		planConfig := atc.PlanConfig{Do: &ppJob.Plan}
		prePromoteTasks = append(prePromoteTasks, planConfig)
	}

	if len(prePromoteTasks) > 0 && !restoreArtifactInPP {
		aggregateJob := atc.PlanConfig{Aggregate: &prePromoteTasks}
		job.Plan = append(job.Plan, aggregateJob)
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
	vars := task.Vars
	if vars == nil {
		vars = make(map[string]string)
	}

	// it is really just a special run job, so let's reuse that
	vars["GCR_PRIVATE_KEY"] = "((gcr.private_key))"
	vars["HALFPIPE_CACHE_TEAM"] = man.Team
	return manifest.Run{
		Retries: task.Retries,
		Name:    task.Name,
		Script:  dockerComposeScript(task.Service, vars, task.Command, man.FeatureToggles.Versioned()),
		Docker: manifest.Docker{
			Image:    config.DockerComposeImage,
			Username: "_json_key",
			Password: "((gcr.private_key))",
		},
		Vars:                   vars,
		SaveArtifacts:          task.SaveArtifacts,
		RestoreArtifacts:       task.RestoreArtifacts,
		SaveArtifactsOnFailure: task.SaveArtifactsOnFailure,
	}
}

func (p pipeline) dockerComposeJob(task manifest.DockerCompose, man manifest.Manifest) *atc.JobConfig {
	return p.runJob(dockerComposeToRunTask(task, man), man, true)
}

func dockerPushJobWithoutRestoreArtifacts(task manifest.DockerPush, resourceName string, man manifest.Manifest) *atc.JobConfig {
	job := atc.JobConfig{
		Name:   task.Name,
		Serial: true,
		Plan: atc.PlanSequence{
			atc.PlanConfig{
				Attempts: task.GetAttempts(),
				Put:      resourceName,
				Params: atc.Params{
					"build":         path.Join(gitDir, man.Repo.BasePath),
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

func dockerPushJobWithRestoreArtifacts(task manifest.DockerPush, resourceName string, man manifest.Manifest) *atc.JobConfig {
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
							fmt.Sprintf("cp -r %s/. %s", artifactsInDir, path.Join(dockerBuildTmpDir, man.Repo.BasePath)),
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
					"build":         path.Join(dockerBuildTmpDir, man.Repo.BasePath),
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

func (p pipeline) dockerPushJob(task manifest.DockerPush, resourceName string, man manifest.Manifest) *atc.JobConfig {
	if task.RestoreArtifacts {
		return dockerPushJobWithRestoreArtifacts(task, resourceName, man)
	}
	return dockerPushJobWithoutRestoreArtifacts(task, resourceName, man)
}

func pathToArtifactsDir(repoName string, basePath string, artifactsDir string) (artifactPath string) {
	fullPath := path.Join(repoName, basePath)
	numberOfParentsToConcourseRoot := len(strings.Split(fullPath, "/"))

	for i := 0; i < numberOfParentsToConcourseRoot; i++ {
		artifactPath += "../"
	}

	artifactPath += artifactsDir
	return
}

func pathToGitRef(repoName string, basePath string) (gitRefPath string) {
	p, _ := filepath.Rel(path.Join(repoName, basePath), path.Join(repoName, ".git", "ref"))
	gitRefPath = windowsToLinuxPath(p)
	return
}

func pathToVersionFile(basePath string) (gitRefPath string) {
	p, _ := filepath.Rel(basePath, path.Join("..", "version", "version"))
	gitRefPath = windowsToLinuxPath(p)
	return
}

func windowsToLinuxPath(path string) (unixPath string) {
	return strings.Replace(path, `\`, "/", -1)
}

func dockerComposeScript(service string, vars map[string]string, containerCommand string, versioningEnabled bool) string {
	envStrings := []string{"-e GIT_REVISION"}
	for key := range vars {
		if key == "GCR_PRIVATE_KEY" {
			continue
		}
		envStrings = append(envStrings, fmt.Sprintf("-e %s", key))
	}
	if versioningEnabled {
		envStrings = append(envStrings, "-e BUILD_VERSION")
	}
	sort.Strings(envStrings)

	composeCommand := fmt.Sprintf("docker-compose run %s %s", strings.Join(envStrings, " "), service)
	if containerCommand != "" {
		composeCommand = fmt.Sprintf("%s %s", composeCommand, containerCommand)
	}

	return fmt.Sprintf(`\docker login -u _json_key -p "$GCR_PRIVATE_KEY" https://eu.gcr.io
%s
`, composeCommand)
}

func (p pipeline) addArtifactResource(cfg *atc.Config, man manifest.Manifest) {
	var savesArtifacts bool
	var savesArtifactsOnFailure bool
	var restoresArtifacts bool

	for _, task := range man.Tasks {
		if task.SavesArtifacts() {
			savesArtifacts = true
		}
		if task.SavesArtifactsOnFailure() {
			savesArtifactsOnFailure = true
		}
		if task.ReadsFromArtifacts() {
			restoresArtifacts = true
		}
	}

	if savesArtifacts || restoresArtifacts || savesArtifactsOnFailure {
		cfg.ResourceTypes = append(cfg.ResourceTypes, p.gcpResourceType())
	}

	if savesArtifacts || restoresArtifacts {
		cfg.Resources = append(cfg.Resources, p.artifactResource(man.Team, man.Pipeline, man.ArtifactConfig))
	}

	if savesArtifactsOnFailure {
		cfg.Resources = append(cfg.Resources, p.artifactResourceOnFailure(man.Team, man.Pipeline, man.ArtifactConfig))
	}
}

func (p pipeline) addCronResource(cfg *atc.Config, man manifest.Manifest) {
	if man.CronTrigger != "" {
		cfg.ResourceTypes = append(cfg.ResourceTypes, cronResourceType())
		cfg.Resources = append(cfg.Resources, p.cronResource(man.CronTrigger))
	}
}

func (p pipeline) addTriggerResource(cfg *atc.Config, man manifest.Manifest) {
	if man.TriggerInterval != "" {
		cfg.Resources = append(cfg.Resources, p.timerResource(man.TriggerInterval))
	}
}

func runScriptArgs(task manifest.Run, man manifest.Manifest, checkForBash bool) []string {

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
  if [ -d $ARTIFACT ] ; then
    mkdir -p $ARTIFACT_OUT_PATH/$ARTIFACT
    cp -r $ARTIFACT/. $ARTIFACT_OUT_PATH/$ARTIFACT/
  elif [ -f $ARTIFACT ] ; then
    ARTIFACT_DIR=$(dirname $ARTIFACT)
    mkdir -p $ARTIFACT_OUT_PATH/$ARTIFACT_DIR
    cp $ARTIFACT $ARTIFACT_OUT_PATH/$ARTIFACT_DIR
  else
    echo "ERROR: Artifact '$ARTIFACT' not found. Try fly hijack to check the filesystem."
    exit 1
  fi
}
`)
	}

	if task.RestoreArtifacts {
		out = append(out, fmt.Sprintf("# Copying in artifacts from previous task"))
		out = append(out, fmt.Sprintf("cp -r %s/. .\n", pathToArtifactsDir(gitDir, man.Repo.BasePath, artifactsInDir)))
	}

	out = append(out,
		fmt.Sprintf("export GIT_REVISION=`cat %s`", pathToGitRef(gitDir, man.Repo.BasePath)),
	)

	if man.FeatureToggles.Versioned() {
		out = append(out,
			fmt.Sprintf("export BUILD_VERSION=`cat %s`", pathToVersionFile(man.Repo.BasePath)),
		)
	}

	scriptCall := fmt.Sprintf(`
%s
EXIT_STATUS=$?
if [ $EXIT_STATUS != 0 ] ; then
%s
fi
`, script, onErrorScript(task.SaveArtifactsOnFailure, pathToArtifactsDir(gitDir, man.Repo.BasePath, artifactsOutDirOnFailure)))
	out = append(out, scriptCall)

	if len(task.SaveArtifacts) != 0 {
		out = append(out, "# Artifacts to copy from task")
	}
	for _, artifactPath := range task.SaveArtifacts {
		out = append(out, fmt.Sprintf("copyArtifact %s %s", artifactPath, pathToArtifactsDir(gitDir, man.Repo.BasePath, artifactsOutDir)))
		//out = append(out, copyArtifactScript(artifactPath, artifactsOutPath))
	}
	return []string{"-c", strings.Join(out, "\n")}
}

func onErrorScript(artifactPaths []string, saveArtifactsOnFailurePath string) string {
	var returnScript []string
	if len(artifactPaths) != 0 {
		returnScript = append(returnScript, "  # Artifacts to copy in case of failure")
	}
	for _, artifactPath := range artifactPaths {
		returnScript = append(returnScript, fmt.Sprintf("  copyArtifact %s %s", artifactPath, saveArtifactsOnFailurePath))
	}
	returnScript = append(returnScript, "  exit 1")
	return strings.Join(returnScript, "\n")
}
