package actions

import (
	"fmt"
	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/renderers/concourse"
	"path"
)

func (a *Actions) deployCFSteps(task manifest.DeployCF) Steps {
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

	deploySteps = append(deploySteps, Step{
		Name: "Push",
		Uses: uses,
		With: addCommonParams(With{
			{"command", "halfpipe-push"},
		}),
		Env: envVars,
	})
	deploySteps = append(deploySteps, Step{
		Name: "Check",
		Uses: uses,
		With: addCommonParams(With{
			{"command", "halfpipe-check"},
		}),
	})

	for _, prePromote := range task.PrePromote {
		prePromoteSteps := Steps{}
		switch prePromote := prePromote.(type) {
		case manifest.Run:
			prePromoteSteps = a.runSteps(prePromote)
		case manifest.DockerCompose:
			prePromoteSteps = a.dockerComposeSteps(prePromote)
		}
		prePromoteSteps[0].Env["TEST_ROUTE"] = concourse.BuildTestRoute(task.CfApplication.Name, task.Space, task.TestDomain)
		deploySteps = append(deploySteps, prePromoteSteps...)
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

	steps := Steps{}
	steps = append(steps, deploySteps...)
	return steps
}
