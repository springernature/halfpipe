package actions

import (
	"fmt"
	"path"
	"strings"

	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/renderers/shared"
)

func (a *Actions) deployCFSteps(task manifest.DeployCF, man manifest.Manifest) (steps Steps) {
	prefix := fmt.Sprintf("../artifacts/%s", man.Triggers.GetGitTrigger().BasePath)
	if strings.HasPrefix(task.Manifest, prefix) {
		//Stupid unused feature?
		task.Manifest = strings.Split(task.Manifest, prefix)[1]
	}

	manifestPath := path.Join(a.workingDir, task.Manifest)
	appPath := a.workingDir
	if len(task.DeployArtifact) > 0 {
		appPath = path.Join(appPath, task.DeployArtifact)
	}

	addCommonParams := func(params With) With {
		commonMap := With{
			"api":          task.API,
			"org":          task.Org,
			"space":        task.Space,
			"username":     task.Username,
			"password":     task.Password,
			"cli_version":  task.CliVersion,
			"manifestPath": manifestPath,
			"testDomain":   task.TestDomain,
			"appPath":      appPath,
		}

		for k, v := range params {
			commonMap[k] = v
		}

		return commonMap
	}

	uses := "docker://eu.gcr.io/halfpipe-io/cf-resource-v2:stable"

	envVars := map[string]string{}
	for k, v := range task.Vars {
		envVars[fmt.Sprintf("CF_ENV_VAR_%s", k)] = v
	}
	envVars["CF_ENV_VAR_BUILD_URL"] = "https://github.com/${{github.repository}}/actions/runs/${{github.run_id}}"

	deploySteps := Steps{}

	if task.SSORoute != "" {
		deploySteps = append(deploySteps, configureSSOStep(task, uses))
	}

	push := Step{
		Name: "Push",
		Uses: uses,
		With: addCommonParams(With{
			"command": "halfpipe-push",
		}),
		Env: envVars,
	}
	if task.CfApplication.Docker != nil {
		push.With["dockerUsername"] = "_json_key"
		push.With["dockerPassword"] = "((halfpipe-gcr.private_key_base64))"
	}
	if task.DockerTag == "gitref" {
		push.With["dockerTag"] = "${{ env.GIT_REVISION }}"
	} else if task.DockerTag == "version" {
		push.With["dockerTag"] = "${{ env.BUILD_VERSION }}"
	}
	deploySteps = append(deploySteps, push)

	deploySteps = append(deploySteps, Step{
		Name: "cf logs --recent",
		If:   "failure()",
		Uses: uses,
		With: addCommonParams(With{
			"command": "halfpipe-logs",
		}),
	})

	deploySteps = append(deploySteps, Step{
		Name: "Check",
		Uses: uses,
		With: addCommonParams(With{
			"command": "halfpipe-check",
		}),
	})

	testRoute := shared.BuildTestRoute(task)
	for _, ppTask := range task.PrePromote {
		switch ppTask := ppTask.(type) {
		case manifest.Run:
			if ppTask.Vars == nil {
				ppTask.Vars = make(map[string]string)
			}
			ppTask.Vars["TEST_ROUTE"] = testRoute
			deploySteps = append(deploySteps, a.runSteps(ppTask)...)
		case manifest.DockerCompose:
			if ppTask.Vars == nil {
				ppTask.Vars = make(map[string]string)
			}
			ppTask.Vars["TEST_ROUTE"] = testRoute
			deploySteps = append(deploySteps, a.dockerComposeSteps(ppTask, "")...)
		case manifest.ConsumerIntegrationTest:
			if ppTask.Vars == nil {
				ppTask.Vars = make(map[string]string)
			}
			ppTask.Vars["TEST_ROUTE"] = testRoute
			if ppTask.ProviderHost == "" {
				ppTask.ProviderHost = testRoute
			}
			deploySteps = append(deploySteps, a.consumerIntegrationTestSteps(ppTask, man)...)
		}
	}

	deploySteps = append(deploySteps, Step{
		Name: "Promote",
		Uses: uses,
		With: addCommonParams(With{
			"command": "halfpipe-promote",
		}),
	})

	sRun := []string{}
	sRun = append(sRun, `echo ":rocket: **Deployment Successful**" >> $GITHUB_STEP_SUMMARY`)
	sRun = append(sRun, `echo "" >> $GITHUB_STEP_SUMMARY`)
	sRun = append(sRun, `echo "[SNPaaS Mission Control](https://mission-control.snpaas.eu/)" >> $GITHUB_STEP_SUMMARY`)
	deploySteps = append(deploySteps, Step{
		Name: "Summary",
		Run:  strings.Join(sRun, "\n"),
	})

	deploySteps = append(deploySteps, Step{
		Name: "Cleanup",
		If:   "${{ !cancelled() }}",
		Uses: uses,
		With: addCommonParams(With{
			"command": "halfpipe-cleanup",
		}),
	})

	steps = append(steps, deploySteps...)
	return steps
}

func configureSSOStep(task manifest.DeployCF, uses string) Step {
	args := `-c "
cf8 login -a $CF_API -u $CF_USERNAME -p $CF_PASSWORD -o $CF_ORG -s $CF_SPACE;
cf8 service sso || cf8 create-user-provided-service sso -r https://ee-sso.public.springernature.app;
cf8 route public.springernature.app -n $SSO_HOST || cf8 create-route public.springernature.app -n $SSO_HOST;
cf8 bind-route-service public.springernature.app -n $SSO_HOST sso;
"`

	return Step{
		Name: "Configure SSO",
		Uses: uses,
		With: With{
			"entrypoint": "/bin/bash",
			"args":       args,
		},
		Env: Env{
			"CF_API":      task.API,
			"CF_ORG":      task.Org,
			"CF_SPACE":    task.Space,
			"CF_USERNAME": task.Username,
			"CF_PASSWORD": task.Password,
			"SSO_HOST":    strings.TrimSuffix(task.SSORoute, ".public.springernature.app"),
		},
	}
}
