package pipeline

import (
	"fmt"
	"strings"

	"text/template"

	"bytes"

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
const artifactsDir = "artifacts"

const artifactsOutDirOnFailure = "artifacts-out-failure"
const artifactsOnFailureName = "artifacts-on-failure"
const artifactsOnFailureDir = "artifacts-on-failure-out"

const gitDir = "git"

const dockerPushResourceName = "Docker Registry"
const dockerBuildTmpDir = "docker_build"

const versionName = "version"

const cronName = "cron"
const timerName = "timer"

func (p pipeline) addSlackResourceTypeAndResource(cfg *atc.Config) {
	slackResourceType := p.slackResourceType()
	cfg.ResourceTypes = append(cfg.ResourceTypes, slackResourceType)

	slackResource := p.slackResource()
	cfg.Resources = append(cfg.Resources, slackResource)
}

func (p pipeline) initialPlan(cfg *atc.Config, man manifest.Manifest, includeVersion bool) []atc.PlanConfig {
	initialPlan := []atc.PlanConfig{{Get: gitDir}}

	if includeVersion {
		initialPlan = append(initialPlan, atc.PlanConfig{Get: versionName})
	} else {
		if man.CronTrigger != "" {
			initialPlan = append(initialPlan, atc.PlanConfig{Get: cronName})
		}
		if man.TriggerInterval != "" {
			initialPlan = append(initialPlan, atc.PlanConfig{Get: timerName})
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

	if man.SlackChannel != "" {
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
		job := p.versionUpdateJob(man)
		if man.SlackChannel != "" {
			failurePlan := slackOnFailurePlan(man.SlackChannel)
			job.Failure = &failurePlan
		}

		job.Plan = append(p.initialPlan(&cfg, man, false), job.Plan...)
		job.Plan = aggregateGets(&job)

		aggregate := *job.Plan[0].Aggregate
		for i := range aggregate {
			aggregate[i].Trigger = true
		}

		cfg.Jobs = append(cfg.Jobs, job)
	}

	for _, t := range man.Tasks {
		initialPlan := p.initialPlan(&cfg, man, man.FeatureToggles.Versioned())

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

		if t.SavesArtifactsOnFailure() || man.SlackChannel != "" {
			sequence := atc.PlanSequence{}

			if t.SavesArtifactsOnFailure() {
				sequence = append(sequence, saveArtifactOnFailurePlan(man.Team, man.Pipeline))
			}
			if man.SlackChannel != "" {
				sequence = append(sequence, slackOnFailurePlan(man.SlackChannel))
			}

			job.Failure = &atc.PlanConfig{
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
		if get.Name() == gitDir ||
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

	if task.RestoreArtifacts {
		jobConfig.Plan = append(jobConfig.Plan, atc.PlanConfig{
			Get:      artifactsName,
			Resource: GenerateArtifactsResourceName(man.Team, man.Pipeline),
		})
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
				Args: runScriptArgs(task.Script, !isDockerCompose, pathToArtifactsDir(gitDir, man.Repo.BasePath), pathToArtifactsOutDir(gitDir, man.Repo.BasePath), task.RestoreArtifacts, task.SaveArtifacts, pathToGitRef(gitDir, man.Repo.BasePath), man.FeatureToggles.Versioned(), pathToVersionFile(man.Repo.BasePath)),
			},
			Inputs: []atc.TaskInputConfig{
				{Name: gitDir},
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

	runTaskIndex := 0
	if task.RestoreArtifacts {
		// If we restore an artifact prior to saving the
		// get of the artifact will be the first task in the plan.
		runTaskIndex = 1
	}

	if len(task.SaveArtifacts) > 0 {
		jobConfig.Plan[runTaskIndex].TaskConfig.Outputs = append(jobConfig.Plan[runTaskIndex].TaskConfig.Outputs, atc.TaskOutputConfig{Name: artifactsOutDir})

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
		jobConfig.Plan[runTaskIndex].TaskConfig.Outputs = append(jobConfig.Plan[runTaskIndex].TaskConfig.Outputs, atc.TaskOutputConfig{Name: artifactsOutDirOnFailure})
	}

	return &jobConfig
}

func (p pipeline) deployCFJob(task manifest.DeployCF, resourceName string, man manifest.Manifest, cfg *atc.Config) *atc.JobConfig {
	manifestPath := path.Join(gitDir, man.Repo.BasePath, task.Manifest)

	if strings.HasPrefix(task.Manifest, fmt.Sprintf("../%s/", artifactsDir)) {
		manifestPath = strings.TrimPrefix(task.Manifest, "../")
	}

	vars := convertVars(task.Vars)

	appPath := path.Join(gitDir, man.Repo.BasePath)
	if len(task.DeployArtifact) > 0 {
		appPath = path.Join(artifactsDir, task.DeployArtifact)
	}

	job := atc.JobConfig{
		Name:   task.Name,
		Serial: true,
	}

	if len(task.DeployArtifact) > 0 || strings.HasPrefix(task.Manifest, fmt.Sprintf("../%s/", artifactsDir)) {
		artifactGet := atc.PlanConfig{
			Get:      artifactsName,
			Resource: GenerateArtifactsResourceName(man.Team, man.Pipeline),
			Params: atc.Params{
				"folder":       artifactsOnFailureDir,
				"version_file": path.Join(gitDir, ".git", "ref"),
			},
		}
		job.Plan = append(job.Plan, artifactGet)
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

func (p pipeline) dockerComposeJob(task manifest.DockerCompose, man manifest.Manifest) *atc.JobConfig {
	vars := task.Vars
	if vars == nil {
		vars = make(map[string]string)
	}

	// it is really just a special run job, so let's reuse that
	vars["GCR_PRIVATE_KEY"] = "((gcr.private_key))"
	runTask := manifest.Run{
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
	job := p.runJob(runTask, man, true)
	return job
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
				Get:      artifactsName,
				Resource: GenerateArtifactsResourceName(man.Team, man.Pipeline),
			},
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
							fmt.Sprintf("cp -r %s/. %s", artifactsDir, path.Join(dockerBuildTmpDir, man.Repo.BasePath)),
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
	if len(task.Vars) > 0 {
		job.Plan[2].Params["build_args"] = convertVars(task.Vars)
	}
	if man.FeatureToggles.Versioned() {
		job.Plan[2].Params["tag_file"] = "version/number"
	}
	return &job
}

func (p pipeline) dockerPushJob(task manifest.DockerPush, resourceName string, man manifest.Manifest) *atc.JobConfig {
	if task.RestoreArtifacts {
		return dockerPushJobWithRestoreArtifacts(task, resourceName, man)
	}
	return dockerPushJobWithoutRestoreArtifacts(task, resourceName, man)
}

func pathToArtifactsDir(repoName string, basePath string) (artifactPath string) {
	fullPath := path.Join(repoName, basePath)
	numberOfParentsToConcourseRoot := len(strings.Split(fullPath, "/"))

	for i := 0; i < numberOfParentsToConcourseRoot; i++ {
		artifactPath += "../"
	}

	artifactPath += artifactsDir
	return
}

func pathToArtifactsOutDir(repoName string, basePath string) (artifactPath string) {
	fullPath := path.Join(repoName, basePath)
	numberOfParentsToConcourseRoot := len(strings.Split(fullPath, "/"))

	for i := 0; i < numberOfParentsToConcourseRoot; i++ {
		artifactPath += "../"
	}

	artifactPath += artifactsOutDir
	return
}

func pathToGitRef(repoName string, basePath string) (gitRefPath string) {
	gitRefPath, _ = filepath.Rel(path.Join(repoName, basePath), path.Join(repoName, ".git", "ref"))
	return
}

func pathToVersionFile(basePath string) (gitRefPath string) {
	gitRefPath, _ = filepath.Rel(basePath, path.Join("..", "version", "version"))
	return
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

	for _, t := range man.Tasks {
		switch task := t.(type) {
		case manifest.Run:
			if len(task.SaveArtifacts) > 0 {
				savesArtifacts = true
			}
			if len(task.SaveArtifactsOnFailure) > 0 {
				savesArtifactsOnFailure = true
			}
		case manifest.DockerCompose:
			if len(task.SaveArtifacts) > 0 {
				savesArtifacts = true
			}
			if len(task.SaveArtifactsOnFailure) > 0 {
				savesArtifactsOnFailure = true
			}
		case manifest.DeployCF:
			if task.DeployArtifact != "" {
				restoresArtifacts = true
			}
		}

	}

	if savesArtifacts || savesArtifactsOnFailure || restoresArtifacts {
		cfg.ResourceTypes = append(cfg.ResourceTypes, p.gcpResourceType())
		cfg.Resources = append(cfg.Resources, p.gcpResource(man.Team, man.Pipeline, man.ArtifactConfig))
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

func runScriptArgs(script string, checkForBash bool, artifactsPath string, artifactsOutPath string, restoreArtifacts bool, saveArtifacts []string, pathToGitRef string, versioningEnabled bool, pathToVersionFile string) []string {
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
fi`)
	}

	if restoreArtifacts {
		out = append(out, fmt.Sprintf("cp -r %s/. .", artifactsPath))
	}

	out = append(out,
		"set -e",
		fmt.Sprintf("export GIT_REVISION=`cat %s`", pathToGitRef),
	)

	if versioningEnabled {
		out = append(out,
			fmt.Sprintf("export BUILD_VERSION=`cat %s`", pathToVersionFile),
		)
	}

	out = append(out,
		script,
	)

	for _, artifact := range saveArtifacts {
		out = append(out, copyArtifactScript(artifactsOutPath, artifact))
	}
	return []string{"-c", strings.Join(out, "\n")}
}

func copyArtifactScript(artifactsPath string, artifactOutputFolder string) string {
	tmpl, err := template.New("runScript").Parse(`
if [ -d {{.ArtifactsPath}} ]
then
  mkdir -p {{.ArtifactOutputFolder}}/{{.ArtifactsPath}}
  cp -r {{.ArtifactsPath}}/. {{.ArtifactOutputFolder}}/{{.ArtifactsPath}}/
elif [ -f {{.ArtifactsPath}} ]
then
  artifactDir=$(dirname {{.ArtifactsPath}})
  mkdir -p {{.ArtifactOutputFolder}}/$artifactDir
  cp {{.ArtifactsPath}} {{.ArtifactOutputFolder}}/$artifactDir
else
  echo "ERROR: Artifact '{{.ArtifactsPath}}' not found. Try fly hijack to check the filesystem."
  exit 1
fi
`)

	if err != nil {
		panic(err)
	}

	byteBuffer := new(bytes.Buffer)
	err = tmpl.Execute(byteBuffer, struct {
		ArtifactsPath  string
		ArtifactOutputFolder string
	}{
		ArtifactsPath:        artifactsPath,
		ArtifactOutputFolder: artifactOutputFolder,
	})

	if err != nil {
		panic(err)
	}

	return byteBuffer.String()
}
