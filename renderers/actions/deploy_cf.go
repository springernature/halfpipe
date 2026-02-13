package actions

import (
	"fmt"
	"maps"
	"path"
	"strings"

	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/renderers/shared"
)

func (a *Actions) commonParamsWith(task manifest.DeployCF, manifestPath string, appPath string, extra With) With {
	commonMap := With{
		"api":          task.API,
		"org":          task.Org,
		"space":        task.Space,
		"username":     task.Username,
		"password":     task.Password,
		"cliVersion":   task.CliVersion,
		"manifestPath": manifestPath,
		"testDomain":   task.TestDomain,
		"appPath":      appPath,
	}

	maps.Copy(commonMap, extra)

	return commonMap
}

func (a *Actions) deployCFSteps(task manifest.DeployCF, man manifest.Manifest) (steps Steps) {
	manifestPath := path.Join(a.workingDir, task.Manifest)
	appPath := a.workingDir
	if len(task.DeployArtifact) > 0 {
		appPath = path.Join(appPath, task.DeployArtifact)
	}

	if task.SSORoute != "" {
		steps = append(steps, a.configureSSOStep(task, manifestPath, appPath))
	}

	if len(task.PrePromote) == 0 {
		steps = append(steps, a.allStep(task, manifestPath, appPath, man))
	} else {
		steps = append(steps, a.pushStep(task, manifestPath, appPath, man))
		steps = append(steps, a.logsStep(task, manifestPath, appPath))
		steps = append(steps, a.checkStep(task, manifestPath, appPath))
		steps = append(steps, a.prePromoteSteps(task, man)...)
		steps = append(steps, a.promoteStep(task, manifestPath, appPath))
		steps = append(steps, a.cleanupStep(task, manifestPath, appPath))
	}

	steps = append(steps, a.SummaryStep())

	return steps
}

func (a *Actions) configureSSOStep(task manifest.DeployCF, manifestPath string, appPath string) Step {
	return Step{
		Name: "Configure SSO",
		Uses: ExternalActions.DeployCF,
		With: a.commonParamsWith(task, manifestPath, appPath, With{
			"command":    "halfpipe-sso",
			"ssoHost":    strings.TrimSuffix(task.SSORoute, ".public.springernature.app"),
			"cliVersion": "cf8",
		}),
	}
}

func (a *Actions) pushStep(task manifest.DeployCF, manifestPath string, appPath string, man manifest.Manifest) Step {
	envVars := map[string]string{}
	for k, v := range task.Vars {
		envVars[fmt.Sprintf("CF_ENV_VAR_%s", k)] = v
	}
	envVars["CF_ENV_VAR_BUILD_URL"] = "https://github.com/${{github.repository}}/actions/runs/${{github.run_id}}"

	push := Step{
		Name: "Push",
		Uses: ExternalActions.DeployCF,
		With: a.commonParamsWith(task, manifestPath, appPath, With{
			"command": "halfpipe-push",
			"team":    man.Team,
			"gitUri":  man.Triggers.GetGitTrigger().URI,
		}),
		Env: envVars,
	}

	if task.CfApplication.Docker != nil {
		push.With["dockerUsername"] = "_json_key"
		push.With["dockerPassword"] = "((halfpipe-gcr.private_key_base64))"

		if task.DockerTag == "gitref" {
			push.With["dockerTag"] = "${{ env.GIT_REVISION }}"
		} else if task.DockerTag == "version" {
			push.With["dockerTag"] = "${{ env.BUILD_VERSION }}"
		}
	}

	return push
}

func (a *Actions) allStep(task manifest.DeployCF, manifestPath string, appPath string, man manifest.Manifest) Step {
	t := a.pushStep(task, manifestPath, appPath, man)
	t.Name = "Deploy"
	t.With["command"] = "halfpipe-all"
	return t
}

func (a *Actions) logsStep(task manifest.DeployCF, manifestPath string, appPath string) Step {
	return Step{
		Name: "cf logs --recent",
		If:   "failure()",
		Uses: ExternalActions.DeployCF,
		With: a.commonParamsWith(task, manifestPath, appPath, With{
			"command": "halfpipe-logs",
		}),
	}
}

func (a *Actions) checkStep(task manifest.DeployCF, manifestPath string, appPath string) Step {
	return Step{
		Name: "Check",
		Uses: ExternalActions.DeployCF,
		With: a.commonParamsWith(task, manifestPath, appPath, With{
			"command": "halfpipe-check",
		}),
	}
}

func (a *Actions) prePromoteSteps(task manifest.DeployCF, man manifest.Manifest) []Step {
	prePromotes := []Step{}
	testRoute := shared.BuildTestRoute(task)
	for _, ppTask := range task.PrePromote {
		switch ppTask := ppTask.(type) {
		case manifest.Run:
			if ppTask.Vars == nil {
				ppTask.Vars = make(map[string]string)
			}
			ppTask.Vars["TEST_ROUTE"] = testRoute
			prePromotes = append(prePromotes, a.runSteps(ppTask)...)
		case manifest.DockerCompose:
			if ppTask.Vars == nil {
				ppTask.Vars = make(map[string]string)
			}
			ppTask.Vars["TEST_ROUTE"] = testRoute
			prePromotes = append(prePromotes, a.dockerComposeSteps(ppTask, man.Team)...)
		case manifest.ConsumerIntegrationTest:
			if ppTask.Vars == nil {
				ppTask.Vars = make(map[string]string)
			}
			ppTask.Vars["TEST_ROUTE"] = testRoute
			if ppTask.ProviderHost == "" {
				ppTask.ProviderHost = testRoute
			}
			prePromotes = append(prePromotes, a.consumerIntegrationTestSteps(ppTask, man)...)
		}
	}

	return prePromotes
}

func (a *Actions) promoteStep(task manifest.DeployCF, manifestPath string, appPath string) Step {
	return Step{
		Name: "Promote",
		Uses: ExternalActions.DeployCF,
		With: a.commonParamsWith(task, manifestPath, appPath, With{
			"command": "halfpipe-promote",
		}),
	}
}

func (a *Actions) SummaryStep() Step {
	sRun := []string{}
	sRun = append(sRun, `echo ":rocket: **Deployment Successful**" >> $GITHUB_STEP_SUMMARY`)
	sRun = append(sRun, `echo "" >> $GITHUB_STEP_SUMMARY`)
	sRun = append(sRun, `echo "[SNPaaS Mission Control](https://mission-control.snpaas.eu/)" >> $GITHUB_STEP_SUMMARY`)
	return Step{
		Name: "Summary",
		Run:  strings.Join(sRun, "\n"),
	}
}

func (a *Actions) cleanupStep(task manifest.DeployCF, manifestPath string, appPath string) Step {
	return Step{
		Name: "Cleanup",
		If:   "${{ !cancelled() }}",
		Uses: ExternalActions.DeployCF,
		With: a.commonParamsWith(task, manifestPath, appPath, With{
			"command": "halfpipe-cleanup",
		}),
	}
}
