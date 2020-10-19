package pipeline

import (
	"fmt"
	"regexp"
	"strings"

	"path/filepath"

	"path"

	cfManifest "code.cloudfoundry.org/cli/util/manifest"
	boshTemplate "github.com/cloudfoundry/bosh-cli/director/template"
	"github.com/concourse/concourse/atc"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"

	"sigs.k8s.io/yaml"
)

type Renderer interface {
	Render(manifest manifest.Manifest) atc.Config
}

type CfManifestReader func(pathToManifest string, pathsToVarsFiles []string, vars []boshTemplate.VarKV) ([]cfManifest.Application, error)

type pipeline struct {
	readCfManifest CfManifestReader
}

func NewPipeline(cfManifestReader CfManifestReader) Renderer {
	return pipeline{readCfManifest: cfManifestReader}
}

const artifactsResourceName = "gcp-resource"
const artifactsName = "artifacts"
const artifactsOutDir = "artifacts-out"
const artifactsInDir = "artifacts"
const artifactsOnFailureName = "artifacts-on-failure"
const artifactsOutDirOnFailure = "artifacts-out-failure"

const gitDir = "git"

const dockerBuildTmpDir = "docker_build"

const versionName = "version"

func parallelizeSteps(steps []atc.Step) atc.Step {
	if len(steps) == 1 {
		return steps[0]
	}

	return atc.Step{
		Config: &atc.InParallelStep{
			Config: atc.InParallelConfig{
				Steps:    steps,
				Limit:    0,
				FailFast: true,
			},
		},
	}
}

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

	taskConfig := atc.TaskConfig{
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
		Config: &atc.TaskStep{
			Name:   "get-artifact",
			Config: &taskConfig,
		},
	}
}

func (p pipeline) initialPlan(man manifest.Manifest, task manifest.Task, previousTaskNames []string) []atc.Step {
	_, isUpdateTask := task.(manifest.Update)
	versioningEnabled := man.FeatureToggles.Versioned()

	var getSteps []atc.Step
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
			getSteps = append(getSteps, atc.Step{Config: &gitClone})

		case manifest.TimerTrigger:
			if isUpdateTask || !versioningEnabled {
				getSteps = append(getSteps, atc.Step{Config: &atc.GetStep{Name: trigger.GetTriggerName()}})
			}
		case manifest.DockerTrigger:
			if isUpdateTask || !versioningEnabled {
				dockerTrigger := atc.GetStep{
					Name: trigger.GetTriggerName(),
					Params: map[string]interface{}{
						"skip_download": true,
					},
				}

				getSteps = append(getSteps, atc.Step{Config: &dockerTrigger})

			}
		case manifest.PipelineTrigger:
			if isUpdateTask || !versioningEnabled {
				pipelineTrigger := atc.GetStep{
					Name: trigger.GetTriggerName(),
				}

				getSteps = append(getSteps, atc.Step{Config: &pipelineTrigger})
			}

		}
	}

	if !isUpdateTask && man.FeatureToggles.Versioned() {
		getSteps = append(getSteps, atc.Step{Config: &atc.GetStep{Name: versionName}})
	}

	getStep := p.configureTriggerOnGets(p.addPassedJobsToGets(parallelizeSteps(getSteps), previousTaskNames), task, man)

	steps := []atc.Step{getStep}

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
func (p pipeline) onFailure(task manifest.Task) *atc.Step {
	onFailureChannels := task.GetNotifications().OnFailure
	if task.SavesArtifactsOnFailure() || len(onFailureChannels) > 0 {
		var sequence []atc.Step

		if task.SavesArtifactsOnFailure() {
			s := saveArtifactOnFailurePlan()
			sequence = append(sequence, atc.Step{Config: &s})
		}

		for _, onFailureChannel := range onFailureChannels {
			sequence = append(sequence, slackOnFailurePlan(onFailureChannel, task.GetNotifications().OnFailureMessage))
		}

		onFailure := parallelizeSteps(sequence)
		return &onFailure
	}
	return nil
}

func (p pipeline) onSuccess(task manifest.Task) *atc.Step {
	onSuccessChannels := task.GetNotifications().OnSuccess
	if len(onSuccessChannels) > 0 {
		var sequence []atc.Step

		for _, onSuccessChannel := range onSuccessChannels {
			sequence = append(sequence, slackOnSuccessPlan(onSuccessChannel, task.GetNotifications().OnSuccessMessage))
		}

		onSuccess := parallelizeSteps(sequence)
		return &onSuccess
	}

	return nil
}

func (p pipeline) taskToJobs(task manifest.Task, man manifest.Manifest, previousTaskNames []string) (job atc.JobConfig) {
	initialPlan := p.initialPlan(man, task, previousTaskNames)
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

	job.OnFailure = p.onFailure(task)
	job.OnSuccess = p.onSuccess(task)
	job.BuildLogRetention = p.buildLogRetention(task)
	job.PlanSequence = append(initialPlan, job.PlanSequence...)

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

func (p pipeline) addPassedJobsToGets(task atc.Step, passedJobs []string) atc.Step {
	_ = task.Config.Visit(atc.StepRecursor{
		OnGet: func(step *atc.GetStep) error {
			step.Passed = passedJobs
			return nil
		},
	})
	return task
}

func (p pipeline) buildLogRetention(task manifest.Task) *atc.BuildLogRetention {
	retention := atc.BuildLogRetention{
		MinimumSucceededBuilds: 1,
	}
	if task.GetBuildHistory() != 0 {
		retention.Builds = task.GetBuildHistory()
	}

	return &retention
}

func (p pipeline) configureTriggerOnGets(step atc.Step, task manifest.Task, man manifest.Manifest) atc.Step {
	if task.IsManualTrigger() {
		return step
	}

	versioningEnabled := man.FeatureToggles.Versioned()
	manualGitTrigger := man.Triggers.GetGitTrigger().ManualTrigger

	_ = step.Config.Visit(atc.StepRecursor{
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

	return step
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

func convertVars(vars manifest.Vars) map[string]interface{} {
	out := make(map[string]interface{})
	for k, v := range vars {
		out[k] = v
	}
	return out
}

// convert string to uppercase and replace non A-Z 0-9 with underscores
func toEnvironmentKey(s string) string {
	return regexp.MustCompile(`[^A-Z0-9]`).ReplaceAllString(strings.ToUpper(s), "_")
}

func ToString(pipeline atc.Config) (string, error) {
	renderedPipeline, err := yaml.Marshal(pipeline)
	if err != nil {
		return "", err
	}

	versionComment := fmt.Sprintf("# Generated using halfpipe cli version %s", config.Version)
	return fmt.Sprintf("%s\n%s", versionComment, renderedPipeline), nil
}

func saveArtifactOnFailurePlan() atc.PutStep {
	return atc.PutStep{
		Name: artifactsOnFailureName,
		Params: atc.Params{
			"folder":       artifactsOutDirOnFailure,
			"version_file": path.Join(gitDir, ".git", "ref"),
			"postfix":      "failure",
		},
	}
}
