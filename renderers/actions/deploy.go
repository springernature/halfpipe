package actions

import (
	"github.com/springernature/halfpipe/manifest"
	"path"
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
			{"appPath", basePath},
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

	return Job{
		Name:   task.GetName(),
		RunsOn: defaultRunner,
		Steps:  []Step{checkoutCode, dockerLogin, deploy, cleanup},
		Env:    Env(task.Vars),
	}
}
