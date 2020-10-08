package defaults

import (
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/project"
	"strings"
)

func defaultGitTrigger(original manifest.GitTrigger, defaults Defaults, branchResolver project.GitBranchResolver) (updated manifest.GitTrigger) {
	updated = original
	updated.BasePath = defaults.Project.BasePath

	if updated.Branch == "" {
		branch, err := branchResolver()
		if err == nil {
			if branch == "master" || branch == "main" {
				updated.Branch = branch
			} else if config.CheckBranch == "false" {
				// This is only here for when we develop on a non master branch.
				// Otherwise the e2e tests will all fail with the only diff being
				//<     branch: ""
				//---
				//>     branch: master
				updated.Branch = "master"
			}
		}
	}

	if updated.URI == "" {
		updated.URI = defaults.Project.GitURI

		for from, to := range config.RewriteGitHTTPToSSH {
			if strings.Contains(updated.URI, from) {
				updated.URI = strings.Replace(updated.URI, from, to, 1)
			}
		}
	}

	if updated.URI != "" && !updated.IsPublic() && updated.PrivateKey == "" {
		updated.PrivateKey = defaults.RepoPrivateKey
	}

	return updated
}
