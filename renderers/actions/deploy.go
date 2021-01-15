package actions

import (
	"path"

	"github.com/springernature/halfpipe/manifest"
)

func (a Actions) deployCfJob(task manifest.DeployCF, man manifest.Manifest) Job {
	dockerLogin := Step{
		Name: "Login to registry",
		Uses: "docker/login-action@v1",
		With: With{
			{Key: "registry", Value: "eu.gcr.io"},
			{Key: "username", Value: "_json_key"},
			{Key: "password", Value: "${{ secrets.EE_GCR_PRIVATE_KEY }}"},
		},
	}

	basePath := man.Triggers.GetGitTrigger().BasePath
	manifestPath := path.Join(basePath, task.Manifest)
	appPath := basePath
	if len(task.DeployArtifact) > 0 {
		appPath = path.Join(basePath, task.DeployArtifact)
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

	steps := []Step{checkoutCode, dockerLogin}
	if task.ReadsFromArtifacts() {
		steps = append(steps, restoreArtifacts)
	}
	steps = append(steps, deploy, cleanup)

	return Job{
		Name:   task.GetName(),
		RunsOn: defaultRunner,
		Steps:  steps,
		Env:    Env(task.Vars),
	}
}
