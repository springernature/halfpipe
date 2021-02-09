package actions

import (
	"fmt"
	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/renderers/shared"
	"path"
	"strings"
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
		return append(With{
			{"api", task.API},
			{"org", task.Org},
			{"space", task.Space},
			{"username", task.Username},
			{"password", task.Password},
			{"cli_version", task.CliVersion},
			{"manifestPath", manifestPath},
			{"testDomain", task.TestDomain},
			{"appPath", appPath},
		}, params...)
	}

	uses := "docker://eu.gcr.io/halfpipe-io/cf-resource-v2:stable"

	envVars := map[string]string{}
	for k, v := range task.Vars {
		envVars[fmt.Sprintf("CF_ENV_VAR_%s", k)] = v
	}
	envVars["CF_ENV_VAR_GITHUB_WORKFLOW_URL"] = "https://github.com/${{github.repository}}/actions/runs/${{github.run_id}}"

	deploySteps := Steps{}

	push := Step{
		Name: "Push",
		Uses: uses,
		With: addCommonParams(With{
			{"command", "halfpipe-push"},
		}),
		Env: envVars,
	}
	if task.CfApplication.DockerImage != "" {
		push.With = append(push.With, With{
			{"dockerUsername", "_json_key"},
			{"dockerPassword", "((halfpipe-gcr.private_key_base64))"},
		}...)
	}
	if task.DockerTag == "gitref" {
		push.With = append(push.With, With{{"dockerTag", "${{ env.GIT_REVISION }}"}}...)
	} else if task.DockerTag == "version" {
		push.With = append(push.With, With{{"dockerTag", "${{ env.BUILD_VERSION }}"}}...)
	}
	deploySteps = append(deploySteps, push)

	deploySteps = append(deploySteps, Step{
		Name: "Check",
		Uses: uses,
		With: addCommonParams(With{
			{"command", "halfpipe-check"},
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
			deploySteps = append(deploySteps, a.dockerComposeSteps(ppTask)...)
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
			{"command", "halfpipe-promote"},
		}),
	})

	deploySteps = append(deploySteps, Step{
		Name: "Cleanup",
		If:   "always()",
		Uses: uses,
		With: addCommonParams(With{
			{"command", "halfpipe-cleanup"},
		}),
	})

	steps = append(steps, deploySteps...)
	return steps
}
