package concourse

import (
	"fmt"
	"github.com/concourse/concourse/atc"
	"github.com/springernature/halfpipe/defaults"
	"github.com/springernature/halfpipe/manifest"
	"path"
	"strings"
)

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

	var steps []atc.Step
	if !task.Rolling {
		steps = append(steps, p.pushCandidateApp(task, resourceName, manifestPath, appPath, vars, man))
		steps = append(steps, p.checkApp(task, resourceName, manifestPath))
		steps = append(steps, p.prePromoteTasks(task, man, basePath)...)
		steps = append(steps, p.promoteCandidateAppToLive(task, resourceName, manifestPath))
		job.Ensure = p.cleanupOldApps(task, resourceName, manifestPath)
	} else {
		if len(task.PrePromote) == 0 {
			steps = append(steps, p.pushAppRolling(task, resourceName, manifestPath, appPath, vars, man))
		} else {
			steps = append(steps, p.pushCandidateApp(task, resourceName, manifestPath, appPath, vars, man))
			steps = append(steps, p.prePromoteTasks(task, man, basePath)...)
			steps = append(steps, p.pushAppRolling(task, resourceName, manifestPath, appPath, vars, man))
			steps = append(steps, p.removeTestApp(resourceName, manifestPath))
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

func (p pipeline) removeTestApp(resourceName string, manifestPath string) atc.Step {
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
	if len(task.PrePromote) == 0 {
		return []atc.Step{}
	}

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

	return []atc.Step{parallelizeSteps(prePromoteTasks)}
}

func buildTestRoute(appName, space, testDomain string) string {
	return fmt.Sprintf("%s-%s-CANDIDATE.%s",
		strings.Replace(appName, "_", "-", -1),
		strings.Replace(space, "_", "-", -1),
		testDomain)
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

	name = fmt.Sprintf(fmt.Sprintf("%s-%s", name, strings.ToLower(task.Space)))
	name = strings.TrimSpace(name)
	return
}
