package actions

import (
	"fmt"
	"path"

	"github.com/springernature/halfpipe/manifest"
)

func (a *Actions) deployCFJob(task manifest.DeployCF) Job {
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
			{"manifestPath", manifestPath},
			{"cli_version", task.CliVersion},
		}, params...)
	}

	//uses := "docker://eu.gcr.io/halfpipe-io/cf-resource-v2:stable"
	uses := "docker://simonjohansson/action-test:latest"

	envVars := map[string]string{}
	for k, v := range task.Vars {
		envVars[fmt.Sprintf("CF_ENV_VAR_%s", k)] = v
	}

	envVars["CF_ENV_VAR_GITHUB_WORKFLOW_URL"] = "https://github.com/${{github.repository}}/actions/runs/${{github.run_id}}"

	deploy := Step{
		Name: "Deploy",
		Uses: uses,
		With: addCommonParams(With{
			{"command", "halfpipe-all"},
			{"testDomain", task.TestDomain},
			{"appPath", appPath},
		}),
		Env: envVars,
	}

	cleanup := Step{
		Name: "Cleanup",
		If:   "always()",
		Uses: uses,
		With: addCommonParams(With{
			{"command", "halfpipe-cleanup"},
		}),
	}

	steps := []Step{checkoutCode}
	if task.ReadsFromArtifacts() {
		steps = append(steps, a.restoreArtifacts()...)
	}
	steps = append(steps, loginHalfpipeGCR, deploy, cleanup)

	return Job{
		Name:   task.GetName(),
		RunsOn: defaultRunner,
		Steps:  steps,
	}
}
