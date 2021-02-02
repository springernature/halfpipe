package concourse

import (
	"fmt"
	"path"
	"strings"

	"github.com/concourse/concourse/atc"
	"github.com/springernature/halfpipe/defaults"
	"github.com/springernature/halfpipe/manifest"
)

func (c Concourse) deployCFJob(task manifest.DeployCF, man manifest.Manifest, basePath string) atc.JobConfig {
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

	var steps []atc.Step
	if !task.Rolling {
		steps = append(steps, c.pushCandidateApp(task, resourceName, manifestPath, appPath, vars, man))
		steps = append(steps, c.checkApp(task, resourceName, manifestPath))
		steps = append(steps, c.prePromoteTasks(task, man, basePath)...)
		steps = append(steps, c.promoteCandidateAppToLive(task, resourceName, manifestPath))
		job.Ensure = c.cleanupOldApps(task, resourceName, manifestPath)
	} else {
		if len(task.PrePromote) == 0 {
			steps = append(steps, c.pushAppRolling(task, resourceName, manifestPath, appPath, vars, man))
		} else {
			steps = append(steps, c.pushCandidateApp(task, resourceName, manifestPath, appPath, vars, man))
			steps = append(steps, c.prePromoteTasks(task, man, basePath)...)
			steps = append(steps, c.pushAppRolling(task, resourceName, manifestPath, appPath, vars, man))
			steps = append(steps, c.removeTestApp(task, resourceName, manifestPath))
		}
	}

	job.PlanSequence = steps
	return job
}

func (c Concourse) cleanupOldApps(task manifest.DeployCF, resourceName string, manifestPath string) *atc.Step {
	cleanup := &atc.PutStep{
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

	step := stepWithAttemptsAndTimeout(cleanup, task.GetAttempts(), task.GetTimeout())
	return &step
}

func (c Concourse) promoteCandidateAppToLive(task manifest.DeployCF, resourceName string, manifestPath string) atc.Step {
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
	return stepWithAttemptsAndTimeout(&promote, task.GetAttempts(), task.GetTimeout())
}

func (c Concourse) checkApp(task manifest.DeployCF, resourceName string, manifestPath string) atc.Step {
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
	return stepWithAttemptsAndTimeout(&check, task.GetAttempts(), task.GetTimeout())
}

func (c Concourse) pushCandidateApp(task manifest.DeployCF, resourceName string, manifestPath string, appPath string, vars map[string]interface{}, man manifest.Manifest) atc.Step {
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
		push.Params["dockerUsername"] = defaults.Concourse.Docker.Username
		push.Params["dockerPassword"] = defaults.Concourse.Docker.Password
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
	if man.FeatureToggles.UpdatePipeline() {
		push.Params["buildVersionPath"] = path.Join("version", "version")
	}

	if task.Rolling {
		push.Name = "deploy-test-app"
		push.Params["instances"] = 1
		push.Params["instances"] = 1
	}

	return stepWithAttemptsAndTimeout(&push, task.GetAttempts(), task.GetTimeout())
}

func (c Concourse) removeTestApp(task manifest.DeployCF, resourceName string, manifestPath string) atc.Step {
	remove := atc.PutStep{
		Name:     "remove-test-app",
		Resource: resourceName,
		Params: atc.Params{
			"command":      "halfpipe-delete-test",
			"manifestPath": manifestPath,
		},
	}
	return stepWithAttemptsAndTimeout(&remove, task.GetAttempts(), task.GetTimeout())
}

func (c Concourse) pushAppRolling(task manifest.DeployCF, resourceName string, manifestPath string, appPath string, vars map[string]interface{}, man manifest.Manifest) atc.Step {
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
		deploy.Params["dockerUsername"] = defaults.Concourse.Docker.Username
		deploy.Params["dockerPassword"] = defaults.Concourse.Docker.Password
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
	if man.FeatureToggles.UpdatePipeline() {
		deploy.Params["buildVersionPath"] = path.Join("version", "version")
	}

	return stepWithAttemptsAndTimeout(&deploy, task.GetAttempts(), task.GetTimeout())
}

func (c Concourse) prePromoteTasks(task manifest.DeployCF, man manifest.Manifest, basePath string) []atc.Step {
	// saveArtifacts and restoreArtifacts are needed to make sure we don't run pre-promote
	// tasks in parallel when the first task saves an artifact and the second restores it.
	if len(task.PrePromote) == 0 {
		return []atc.Step{}
	}

	var prePromoteTasks []atc.Step
	for _, t := range task.PrePromote {
		testRoute := BuildTestRoute(task.CfApplication.Name, task.Space, task.TestDomain)
		var ppJob atc.JobConfig
		switch ppTask := t.(type) {
		case manifest.Run:
			if len(ppTask.Vars) == 0 {
				ppTask.Vars = make(map[string]string)
			}
			ppTask.Vars["TEST_ROUTE"] = testRoute
			ppJob = c.runJob(ppTask, man, false, basePath)
		case manifest.DockerCompose:
			if len(ppTask.Vars) == 0 {
				ppTask.Vars = make(map[string]string)
			}
			ppTask.Vars["TEST_ROUTE"] = testRoute
			runTask := convertDockerComposeToRunTask(ppTask, man)
			ppJob = c.runJob(runTask, man, true, basePath)

		case manifest.ConsumerIntegrationTest:
			if ppTask.ProviderHost == "" {
				ppTask.ProviderHost = testRoute
			}
			runTask := convertConsumerIntegrationTestToRunTask(ppTask, man)
			ppJob = c.runJob(runTask, man, true, basePath)
		}
		prePromoteTasks = append(prePromoteTasks, ppJob.PlanSequence...)
	}

	return []atc.Step{parallelizeSteps(prePromoteTasks)}
}

func BuildTestRoute(appName, space, testDomain string) string {
	return fmt.Sprintf("%s-%s-CANDIDATE.%s", strings.Replace(appName, "_", "-", -1), strings.Replace(space, "_", "-", -1), testDomain)
}

func deployCFResourceName(task manifest.DeployCF) (name string) {
	// if url remove the scheme
	api := strings.Replace(task.API, "https://", "", -1)
	api = strings.Replace(api, "http://", "", -1)
	api = strings.Replace(api, "((cloudfoundry.api-", "", -1)
	api = strings.Replace(api, "))", "", -1)
	api = strings.ToLower(api)

	name = fmt.Sprintf("cf-%s", api)
	if task.Rolling {
		name = fmt.Sprintf("rolling-cf-%s", api)

	}

	if org := strings.Replace(task.Org, "((cloudfoundry.org-snpaas))", "", -1); org != "" {
		name = fmt.Sprintf("%s-%s", name, strings.ToLower(org))
	}

	name = fmt.Sprintf("%s-%s", name, strings.ToLower(task.Space))
	name = strings.TrimSpace(name)
	return
}
