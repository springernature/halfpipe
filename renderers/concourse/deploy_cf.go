package concourse

import (
	"fmt"
	"github.com/springernature/halfpipe/renderers/shared"
	"path"
	"strings"

	"github.com/concourse/concourse/atc"
	"github.com/springernature/halfpipe/defaults"
	"github.com/springernature/halfpipe/manifest"
)

func (c Concourse) deployCFJob(task manifest.DeployCF, man manifest.Manifest, basePath string) atc.JobConfig {
	deploy := deployCF{}
	deploy.task = task
	deploy.resourceName = deployCFResourceName(task)
	deploy.halfpipeManifest = man
	deploy.basePath = basePath
	deploy.vars = convertVars(task.Vars)
	deploy.team = man.Team

	deploy.manifestPath = path.Join(gitDir, basePath, task.Manifest)
	if strings.HasPrefix(task.Manifest, fmt.Sprintf("../%s/", artifactsInDir)) {
		deploy.manifestPath = strings.TrimPrefix(task.Manifest, "../")
	}

	deploy.appPath = path.Join(gitDir, basePath)
	if len(task.DeployArtifact) > 0 {
		deploy.appPath = path.Join(artifactsInDir, basePath, task.DeployArtifact)
	}

	job := atc.JobConfig{
		Name:   task.GetName(),
		Serial: true,
	}

	var steps []atc.Step

	if task.SSORoute != "" {
		steps = append(steps, deploy.configureSSO())
	}

	if len(task.PrePromote) == 0 {
		steps = append(steps, deploy.pushApp())
	} else if task.Rolling {
		steps = append(steps, deploy.logsOnFailure(deploy.pushCandidateApp()))
		steps = append(steps, c.prePromoteTasks(deploy)...)
		steps = append(steps, deploy.pushApp())
		steps = append(steps, deploy.removeTestApp())
	} else {
		steps = append(steps, deploy.logsOnFailure(deploy.pushCandidateApp()))
		steps = append(steps, deploy.checkApp())
		steps = append(steps, c.prePromoteTasks(deploy)...)
		steps = append(steps, deploy.promoteCandidateAppToLive())
		job.Ensure = deploy.cleanupOldApps()
	}

	job.PlanSequence = steps
	return job
}

type deployCF struct {
	task             manifest.DeployCF
	resourceName     string
	halfpipeManifest manifest.Manifest
	manifestPath     string
	appPath          string
	basePath         string
	vars             map[string]any
	team             string
}

func (d deployCF) cleanupOldApps() *atc.Step {
	cleanup := &atc.PutStep{
		Name:     "halfpipe-cleanup",
		Resource: d.resourceName,
		Params: atc.Params{
			"command":      "halfpipe-cleanup",
			"manifestPath": d.manifestPath,
			"cliVersion":   d.task.CliVersion,
		},
		NoGet: true,
	}
	if d.task.Timeout != "" {
		cleanup.Params["timeout"] = d.task.Timeout
	}

	step := stepWithAttemptsAndTimeout(cleanup, d.task.GetAttempts(), d.task.GetTimeout())
	return &step
}

func (d deployCF) promoteCandidateAppToLive() atc.Step {
	promote := atc.PutStep{
		Name:     "halfpipe-promote",
		Resource: d.resourceName,
		Params: atc.Params{
			"command":      "halfpipe-promote",
			"testDomain":   d.task.TestDomain,
			"manifestPath": d.manifestPath,
			"cliVersion":   d.task.CliVersion,
		},
		NoGet: true,
	}
	if d.task.Timeout != "" {
		promote.Params["timeout"] = d.task.Timeout
	}
	return stepWithAttemptsAndTimeout(&promote, d.task.GetAttempts(), d.task.GetTimeout())
}

func (d deployCF) checkApp() atc.Step {
	check := atc.PutStep{
		Name:     "halfpipe-check",
		Resource: d.resourceName,
		Params: atc.Params{
			"command":      "halfpipe-check",
			"manifestPath": d.manifestPath,
			"cliVersion":   d.task.CliVersion,
		},
		NoGet: true,
	}
	if d.task.Timeout != "" {
		check.Params["timeout"] = d.task.Timeout
	}
	return stepWithAttemptsAndTimeout(&check, d.task.GetAttempts(), d.task.GetTimeout())
}

func (d deployCF) pushCandidateApp() atc.Step {
	push := atc.PutStep{
		Name:     "halfpipe-push",
		Resource: d.resourceName,
		Params: atc.Params{
			"command":      "halfpipe-push",
			"testDomain":   d.task.TestDomain,
			"manifestPath": d.manifestPath,
			"gitRefPath":   path.Join(gitDir, ".git", "ref"),
			"gitUri":       d.halfpipeManifest.Triggers.GetGitTrigger().URI,
			"cliVersion":   d.task.CliVersion,
			"team":         d.team,
		},
		NoGet: true,
	}

	if d.task.IsDockerPush {
		push.Params["dockerUsername"] = defaults.Concourse.Docker.Username
		push.Params["dockerPassword"] = defaults.Concourse.Docker.Password
		if d.task.DockerTag != "" {
			if d.task.DockerTag == "version" {
				push.Params["dockerTag"] = path.Join(versionName, "version")
			} else if d.task.DockerTag == "gitref" {
				push.Params["dockerTag"] = path.Join(gitDir, ".git", "ref")
			}
		}
	} else {
		push.Params["appPath"] = d.appPath
	}

	if len(d.vars) > 0 {
		push.Params["vars"] = d.vars
	}
	if d.task.Timeout != "" {
		push.Params["timeout"] = d.task.Timeout
	}
	if len(d.task.PreStart) > 0 {
		push.Params["preStartCommand"] = strings.Join(d.task.PreStart, "; ")
	}
	if d.halfpipeManifest.FeatureToggles.UpdatePipeline() {
		push.Params["buildVersionPath"] = path.Join("version", "version")
	}

	if d.task.Rolling {
		push.Name = "deploy-test-app"
		push.Params["instances"] = 1
	}

	return stepWithAttemptsAndTimeout(&push, d.task.GetAttempts(), d.task.GetTimeout())
}

func (d deployCF) removeTestApp() atc.Step {
	remove := atc.PutStep{
		Name:     "remove-test-app",
		Resource: d.resourceName,
		Params: atc.Params{
			"command":      "halfpipe-delete-test",
			"manifestPath": d.manifestPath,
		},
		NoGet: true,
	}
	return stepWithAttemptsAndTimeout(&remove, d.task.GetAttempts(), d.task.GetTimeout())
}

func (d deployCF) pushApp() atc.Step {
	command := "halfpipe-all"
	cliVersion := d.task.CliVersion
	if d.task.Rolling {
		command = "halfpipe-rolling-deploy"
		cliVersion = "cf7"
	}
	push := atc.PutStep{
		Name:     command,
		Resource: d.resourceName,
		Params: atc.Params{
			"command":      command,
			"testDomain":   d.task.TestDomain,
			"manifestPath": d.manifestPath,
			"gitRefPath":   path.Join(gitDir, ".git", "ref"),
			"gitUri":       d.halfpipeManifest.Triggers.GetGitTrigger().URI,
			"cliVersion":   cliVersion,
			"team":         d.team,
		},
		NoGet: true,
	}

	if d.task.IsDockerPush {
		push.Params["dockerUsername"] = defaults.Concourse.Docker.Username
		push.Params["dockerPassword"] = defaults.Concourse.Docker.Password
		if d.task.DockerTag != "" {
			if d.task.DockerTag == "version" {
				push.Params["dockerTag"] = path.Join(versionName, "version")
			} else if d.task.DockerTag == "gitref" {
				push.Params["dockerTag"] = path.Join(gitDir, ".git", "ref")
			}
		}
	} else {
		push.Params["appPath"] = d.appPath
	}

	if len(d.vars) > 0 {
		push.Params["vars"] = d.vars
	}
	if d.task.Timeout != "" {
		push.Params["timeout"] = d.task.Timeout
	}
	if len(d.task.PreStart) > 0 {
		push.Params["preStartCommand"] = strings.Join(d.task.PreStart, "; ")
	}
	if d.halfpipeManifest.FeatureToggles.UpdatePipeline() {
		push.Params["buildVersionPath"] = path.Join("version", "version")
	}

	return d.logsOnFailure(stepWithAttemptsAndTimeout(&push, d.task.GetAttempts(), d.task.GetTimeout()))
}

func (d deployCF) logsOnFailure(stepConfig atc.Step) atc.Step {
	return atc.Step{
		Config: &atc.OnFailureStep{
			Step: stepConfig.Config,
			Hook: atc.Step{
				Config: &atc.PutStep{
					Name:     "cf-logs",
					Resource: d.resourceName,
					Params: atc.Params{
						"command":      "halfpipe-logs",
						"cliVersion":   d.task.CliVersion,
						"manifestPath": d.manifestPath,
					},
					NoGet: true,
				},
			},
		},
	}
}

func (d deployCF) configureSSO() atc.Step {
	configure := atc.PutStep{
		Name:     "configure-sso",
		Resource: d.resourceName,
		Params: atc.Params{
			"command":      "halfpipe-sso",
			"manifestPath": d.manifestPath,
			"ssoHost":      strings.TrimSuffix(d.task.SSORoute, ".public.springernature.app"),
			"cliVersion":   "cf8",
		},
		NoGet: true,
	}
	return stepWithAttemptsAndTimeout(&configure, d.task.GetAttempts(), d.task.GetTimeout())
}

func (c Concourse) prePromoteTasks(deploy deployCF) []atc.Step {
	// saveArtifacts and restoreArtifacts are needed to make sure we don't run pre-promote
	// tasks in parallel when the first task saves an artifact and the second restores it.
	if len(deploy.task.PrePromote) == 0 {
		return []atc.Step{}
	}

	testRoute := shared.BuildTestRoute(deploy.task)
	var prePromoteTasks []atc.Step
	for _, t := range deploy.task.PrePromote {
		var ppJob atc.JobConfig
		switch ppTask := t.(type) {
		case manifest.Run:
			if len(ppTask.Vars) == 0 {
				ppTask.Vars = make(map[string]string)
			}
			ppTask.Vars["TEST_ROUTE"] = testRoute
			ppJob = c.runJob(ppTask, deploy.halfpipeManifest, false, deploy.basePath)
		case manifest.DockerCompose:
			if len(ppTask.Vars) == 0 {
				ppTask.Vars = make(map[string]string)
			}
			ppTask.Vars["TEST_ROUTE"] = testRoute
			runTask := convertDockerComposeToRunTask(ppTask, deploy.halfpipeManifest)
			ppJob = c.runJob(runTask, deploy.halfpipeManifest, true, deploy.basePath)

		case manifest.ConsumerIntegrationTest:
			if ppTask.ProviderHost == "" {
				ppTask.ProviderHost = testRoute
			}
			runTask := convertConsumerIntegrationTestToRunTask(ppTask, deploy.halfpipeManifest)
			ppJob = c.runJob(runTask, deploy.halfpipeManifest, true, deploy.basePath)
		}
		prePromoteTasks = append(prePromoteTasks, ppJob.PlanSequence...)
	}

	return []atc.Step{parallelizeSteps(prePromoteTasks)}
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
