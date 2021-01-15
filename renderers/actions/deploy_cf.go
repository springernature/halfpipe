package actions

import (
	"path"

	"github.com/springernature/halfpipe/manifest"
)

func (a Actions) deployCFJob(task manifest.DeployCF, man manifest.Manifest) Job {
	basePath := man.Triggers.GetGitTrigger().BasePath
	manifestPath := path.Join(basePath, task.Manifest)
	appPath := basePath
	if len(task.DeployArtifact) > 0 {
		appPath = path.Join(appPath, task.DeployArtifact)
	}

	deploy := Step{
		Name: "Deploy",
		//Uses: "docker://eu.gcr.io/halfpipe-io/cf-resource-v2:stable",
		Uses: "docker://simonjohansson/action-test:latest",
		With: With{
			{"api", task.API},
			{"org", task.Org},
			{"space", task.Space},
			{"username", task.Username},
			{"password", task.Password},
			{"command", "halfpipe-all"},
			{"appPath", appPath},
			{"manifestPath", manifestPath},
			{"testDomain", task.TestDomain},
			{"cli_version", task.CliVersion},
		},
	}

	cleanup := Step{
		Name: "Cleanup",
		If:   "always()",
		Uses: "docker://simonjohansson/action-test:latest",
		With: With{
			{"api", task.API},
			{"org", task.Org},
			{"space", task.Space},
			{"username", task.Username},
			{"password", task.Password},
			{"command", "halfpipe-cleanup"},
			{"manifestPath", manifestPath},
			{"cli_version", task.CliVersion},
		},
	}

	steps := []Step{checkoutCode}
	if task.ReadsFromArtifacts() {
		steps = append(steps, restoreArtifacts)
	}
	steps = append(steps, gcrLogin, deploy, cleanup)

	return Job{
		Name:   task.GetName(),
		RunsOn: defaultRunner,
		Steps:  steps,
		Env:    Env(task.Vars),
	}
}
