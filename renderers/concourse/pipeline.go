package concourse

import (
	"fmt"
	"github.com/springernature/halfpipe/renderers/shared"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/concourse/concourse/atc"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"

	"sigs.k8s.io/yaml"
)

type Concourse struct {
	halfpipeFilePath string
}

func NewPipeline(halfpipeFilePath string) Concourse {
	return Concourse{
		halfpipeFilePath: halfpipeFilePath,
	}
}

func (c Concourse) PlatformURL(man manifest.Manifest) string {
	return fmt.Sprintf("%s/teams/%s/pipelines/%s", config.ConcourseURL, man.Team, man.PipelineName())
}

const artifactsResourceName = "gcp-resource"
const artifactsName = "artifacts"
const artifactsOutDir = "artifacts-out"
const artifactsInDir = "artifacts"
const artifactsOnFailureName = "artifacts-on-failure"
const artifactsOutDirOnFailure = "artifacts-out-failure"

const defaultStepAttempts = 2
const defaultStepTimeout = "15m"

const gitDir = "git"

const dockerBuildTmpDir = "docker_build"

const versionName = "version"

func parallelizeSteps(steps []atc.Step) atc.Step {
	if len(steps) == 1 {
		return steps[0]
	}

	limit := 0
	if len(steps) > 5 {
		limit = 5
	}

	return atc.Step{
		Config: &atc.InParallelStep{
			Config: atc.InParallelConfig{
				Steps:    steps,
				Limit:    limit,
				FailFast: limit == 0,
			},
		},
	}
}

func restoreArtifactTask(man manifest.Manifest) atc.Step {
	// This function is used in concourse.artifactResource for some reason to lowercase
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

	taskStep := &atc.TaskStep{
		Name: "get-artifact",
		Config: &atc.TaskConfig{
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
				{Name: manifest.GitTrigger{}.GetTriggerName()},
			},
			Outputs: []atc.TaskOutputConfig{
				{Name: artifactsInDir},
			},
		},
	}

	return stepWithAttemptsAndTimeout(taskStep, defaultStepAttempts, defaultStepTimeout)
}

func (c Concourse) initialPlan(man manifest.Manifest, task manifest.Task, previousTaskNames []string) []atc.Step {
	_, isUpdateTask := task.(manifest.Update)
	versioningEnabled := man.FeatureToggles.UpdatePipeline()

	var getSteps []atc.Step
	for _, trigger := range man.Triggers {
		switch trigger := trigger.(type) {
		case manifest.GitTrigger:
			getGit := &atc.GetStep{
				Name: trigger.GetTriggerName(),
			}
			if trigger.Shallow {
				getGit.Params = map[string]interface{}{
					"depth": 1,
				}
			}
			getSteps = append(getSteps, atc.Step{Config: getGit})

		case manifest.TimerTrigger:
			if isUpdateTask || !versioningEnabled {
				getTimer := &atc.GetStep{Name: trigger.GetTriggerName()}
				getSteps = append(getSteps, atc.Step{Config: getTimer})
			}

		case manifest.DockerTrigger:
			if isUpdateTask || !versioningEnabled {
				getDocker := &atc.GetStep{
					Name: trigger.GetTriggerName(),
					Params: map[string]interface{}{
						"skip_download": true,
					},
				}
				getSteps = append(getSteps, atc.Step{Config: getDocker})
			}

		case manifest.PipelineTrigger:
			if isUpdateTask || !versioningEnabled {
				getPipeline := &atc.GetStep{
					Name: trigger.GetTriggerName(),
				}
				getSteps = append(getSteps, atc.Step{Config: getPipeline})
			}
		}
	}

	if !isUpdateTask && man.FeatureToggles.UpdatePipeline() {
		getVersion := &atc.GetStep{Name: versionName}
		getSteps = append(getSteps, atc.Step{Config: getVersion})
	}

	parallelSteps := stepWithAttemptsAndTimeout(parallelizeSteps(getSteps).Config, defaultStepAttempts, defaultStepTimeout)
	parallelGetStep := c.configureTriggerOnGets(c.addPassedJobsToGets(parallelSteps, previousTaskNames), task, man)

	steps := []atc.Step{parallelGetStep}

	if task.ReadsFromArtifacts() {
		steps = append(steps, restoreArtifactTask(man))
	}

	return steps
}

func (c Concourse) dockerPushResources(tasks manifest.TaskList, ociBuild bool) (resourceConfigs atc.ResourceConfigs) {
	for _, task := range tasks {
		switch task := task.(type) {
		case manifest.DockerPush:
			resourceConfigs = append(resourceConfigs, c.dockerPushResource(task, ociBuild))
		case manifest.Parallel:
			resourceConfigs = append(resourceConfigs, c.dockerPushResources(task.Tasks, ociBuild)...)
		case manifest.Sequence:
			resourceConfigs = append(resourceConfigs, c.dockerPushResources(task.Tasks, ociBuild)...)
		}
	}

	return resourceConfigs
}
func (c Concourse) pipelineResources(triggers manifest.TriggerList) (resourceTypes atc.ResourceTypes, resourceConfigs atc.ResourceConfigs) {

	for _, trigger := range triggers {
		switch trigger := trigger.(type) {
		case manifest.PipelineTrigger:
			resourceConfigs = append(resourceConfigs, c.pipelineTriggerResource(trigger))
		}
	}

	if len(resourceConfigs) > 0 {
		resourceTypes = append(resourceTypes, halfpipePipelineTriggerResourceType())
	}

	return resourceTypes, resourceConfigs
}

func (c Concourse) cfPushResources(man manifest.Manifest) (resourceTypes atc.ResourceTypes, resourceConfigs atc.ResourceConfigs) {

	for _, task := range man.Tasks.Flatten() {
		switch task := task.(type) {
		case manifest.DeployCF:
			resourceConfig := c.deployCFResource(task, deployCFResourceName(task))
			if _, found := resourceConfigs.Lookup(resourceConfig.Name); !found {
				resourceConfigs = append(resourceConfigs, resourceConfig)
			}
		}
	}

	if len(resourceConfigs) > 0 {
		resourceTypes = append(resourceTypes, c.halfpipeCfDeployResourceType())
	}

	return
}

func (c Concourse) resourceConfigs(man manifest.Manifest) (resourceTypes atc.ResourceTypes, resourceConfigs atc.ResourceConfigs) {
	for _, trigger := range man.Triggers {
		switch trigger := trigger.(type) {
		case manifest.GitTrigger:
			resourceConfigs = append(resourceConfigs, c.gitResource(trigger))
		case manifest.TimerTrigger:
			resourceTypes = append(resourceTypes, cronResourceType())
			resourceConfigs = append(resourceConfigs, c.cronResource(trigger))
		case manifest.DockerTrigger:
			resourceConfigs = append(resourceConfigs, c.dockerTriggerResource(trigger))
		}
	}

	if man.Tasks.UsesNotifications() {
		resourceTypes = append(resourceTypes, c.slackResourceType())
		resourceConfigs = append(resourceConfigs, c.slackResource())
	}

	if man.Tasks.SavesArtifacts() || man.Tasks.SavesArtifactsOnFailure() {
		resourceTypes = append(resourceTypes, c.gcpResourceType())

		if man.Tasks.SavesArtifacts() {
			resourceConfigs = append(resourceConfigs, c.artifactResource(man))
		}
		if man.Tasks.SavesArtifactsOnFailure() {
			resourceConfigs = append(resourceConfigs, c.artifactResourceOnFailure(man))
		}
	}

	if man.FeatureToggles.UpdatePipeline() {
		resourceConfigs = append(resourceConfigs, c.versionResource(man))
	}

	resourceConfigs = append(resourceConfigs, c.dockerPushResources(man.Tasks, man.FeatureToggles.DockerOciBuild())...)

	cfResourceTypes, cfResources := c.cfPushResources(man)
	resourceTypes = append(resourceTypes, cfResourceTypes...)
	resourceConfigs = append(resourceConfigs, cfResources...)

	pipelineResourceTypes, pipelineResources := c.pipelineResources(man.Triggers)
	resourceTypes = append(resourceTypes, pipelineResourceTypes...)
	resourceConfigs = append(resourceConfigs, pipelineResources...)

	if man.FeatureToggles.GithubStatuses() {
		resourceTypes = append(resourceTypes, c.githubStatusesResourceType())
		resourceConfigs = append(resourceConfigs, c.githubStatusesResource(man))
	}

	return resourceTypes, resourceConfigs
}
func (c Concourse) onFailure(task manifest.Task, man manifest.Manifest) *atc.Step {
	onFailureChannels := task.GetNotifications().OnFailure
	if task.SavesArtifactsOnFailure() || len(onFailureChannels) > 0 || man.FeatureToggles.GithubStatuses() {
		var sequence []atc.Step

		if task.SavesArtifactsOnFailure() {
			sequence = append(sequence, saveArtifactOnFailurePlan())
		}

		for _, onFailureChannel := range onFailureChannels {
			sequence = append(sequence, slackOnFailurePlan(onFailureChannel, task.GetNotifications().OnFailureMessage))
		}

		if man.FeatureToggles.GithubStatuses() {
			sequence = append(sequence, statusesOnFailurePlan(man, task))
		}

		onFailure := stepWithAttemptsAndTimeout(parallelizeSteps(sequence).Config, defaultStepAttempts, defaultStepTimeout)
		return &onFailure
	}
	return nil
}

func (c Concourse) onSuccess(task manifest.Task, man manifest.Manifest) *atc.Step {
	onSuccessChannels := task.GetNotifications().OnSuccess
	if len(onSuccessChannels) > 0 || man.FeatureToggles.GithubStatuses() {
		var sequence []atc.Step

		for _, onSuccessChannel := range onSuccessChannels {
			sequence = append(sequence, slackOnSuccessPlan(onSuccessChannel, task.GetNotifications().OnSuccessMessage))
		}

		if man.FeatureToggles.GithubStatuses() {
			sequence = append(sequence, statusesOnSuccessPlan(man, task))
		}

		onSuccess := stepWithAttemptsAndTimeout(parallelizeSteps(sequence).Config, defaultStepAttempts, defaultStepTimeout)
		return &onSuccess
	}

	return nil
}

func (c Concourse) taskToJobs(task manifest.Task, man manifest.Manifest, previousTaskNames []string) (job atc.JobConfig) {
	initialPlan := c.initialPlan(man, task, previousTaskNames)
	basePath := man.Triggers.GetGitTrigger().BasePath

	switch task := task.(type) {
	case manifest.Run:
		job = c.runJob(task, man, false, basePath)

	case manifest.DockerCompose:
		runTask := convertDockerComposeToRunTask(task, man)
		job = c.runJob(runTask, man, true, basePath)

	case manifest.DeployCF:
		job = c.deployCFJob(task, man, basePath)

	case manifest.DeployKatee:
		job = c.deployKateeJob(task, man, basePath)

	case manifest.DockerPush:
		job = c.dockerPushJob(task, basePath, man)

	case manifest.ConsumerIntegrationTest:
		runTask := convertConsumerIntegrationTestToRunTask(task, man)
		job = c.runJob(runTask, man, true, basePath)

	case manifest.DeployMLZip:
		runTask := shared.ConvertDeployMLZip(task, man)
		job = c.runJob(runTask, man, false, basePath)

	case manifest.DeployMLModules:
		runTask := shared.ConvertDeployMLModules(task, man)
		job = c.runJob(runTask, man, false, basePath)

	case manifest.Update:
		job = c.updateJobConfig(task, man.PipelineName(), basePath)
	}

	job.OnFailure = c.onFailure(task, man)
	job.OnSuccess = c.onSuccess(task, man)
	job.BuildLogRetention = c.buildLogRetention(task)
	job.PlanSequence = append(initialPlan, job.PlanSequence...)

	return job
}

func (c Concourse) Render(man manifest.Manifest) (string, error) {
	atcConfig := c.RenderAtcConfig(man)
	pipelineYaml, err := yaml.Marshal(atcConfig)
	if err != nil {
		return "", err
	}
	return string(pipelineYaml), nil
}

func (c Concourse) RenderAtcConfig(man manifest.Manifest) (cfg atc.Config) {
	resourceTypes, resourceConfigs := c.resourceConfigs(man)
	cfg.ResourceTypes = append(cfg.ResourceTypes, resourceTypes...)
	cfg.Resources = append(cfg.Resources, resourceConfigs...)

	type parentTask struct {
		isParallel bool
		passed     []string
	}

	var jobs func(manifest.TaskList, *parentTask)
	jobs = func(tasks manifest.TaskList, parent *parentTask) {
		for i, task := range tasks {
			passed := tasks.PreviousTaskNames(i)
			if parent != nil {
				if parent.isParallel || i == 0 {
					passed = parent.passed
				}
			}
			switch task := task.(type) {
			case manifest.Parallel:
				jobs(task.Tasks, &parentTask{isParallel: true, passed: passed})
			case manifest.Sequence:
				jobs(task.Tasks, &parentTask{isParallel: false, passed: passed})
			default:
				cfg.Jobs = append(cfg.Jobs, c.taskToJobs(task, man, passed))
			}
		}
	}
	jobs(man.Tasks, nil)
	return cfg
}

func (c Concourse) addPassedJobsToGets(task atc.Step, passedJobs []string) atc.Step {
	_ = task.Config.Visit(atc.StepRecursor{
		OnGet: func(step *atc.GetStep) error {
			step.Passed = passedJobs
			return nil
		},
	})
	return task
}

func (c Concourse) buildLogRetention(task manifest.Task) *atc.BuildLogRetention {
	retention := atc.BuildLogRetention{
		MinimumSucceededBuilds: 1,
	}
	if task.GetBuildHistory() != 0 {
		retention.Builds = task.GetBuildHistory()
	}

	return &retention
}

func (c Concourse) configureTriggerOnGets(step atc.Step, task manifest.Task, man manifest.Manifest) atc.Step {
	if task.IsManualTrigger() {
		return step
	}

	versioningEnabled := man.FeatureToggles.UpdatePipeline()
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

func windowsToLinuxPath(path string) (unixPath string) {
	return strings.Replace(path, `\`, "/", -1)
}

func pathToVersionFile(repoName string, basePath string) (gitRefPath string) {
	p := path.Join(relativePathToRepoRoot(repoName, basePath), path.Join("..", "version", "version"))
	return windowsToLinuxPath(p)
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

func saveArtifactOnFailurePlan() atc.Step {
	return atc.Step{
		Config: &atc.PutStep{
			Name: artifactsOnFailureName,
			Params: atc.Params{
				"folder":       artifactsOutDirOnFailure,
				"version_file": path.Join(gitDir, ".git", "ref"),
				"postfix":      "failure",
			},
		},
	}
}

func stepWithAttemptsAndTimeout(stepConfig atc.StepConfig, attempts int, timeout string) atc.Step {
	timeoutStep := &atc.TimeoutStep{
		Step:     stepConfig,
		Duration: timeout,
	}

	if attempts == 1 {
		return atc.Step{Config: timeoutStep}
	}

	return atc.Step{
		Config: &atc.RetryStep{
			Step:     timeoutStep,
			Attempts: attempts,
		},
	}

}
