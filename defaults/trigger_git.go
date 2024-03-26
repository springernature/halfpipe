package defaults

import (
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/project"
	"strings"
)

func defaultGitTrigger(original manifest.GitTrigger, defaults Defaults, branchResolver project.GitBranchResolver, platform manifest.Platform) (updated manifest.GitTrigger) {
	updated = original
	updated.BasePath = defaults.Project.BasePath

	if updated.Branch == "" {
		branch, err := branchResolver()
		if err == nil {
			if branch == "master" || branch == "main" {
				updated.Branch = branch
			}
		}
	}

	if updated.URI == "" {
		updated.URI = defaults.Project.GitURI

		for from, to := range config.RewriteGitHTTPToSSH {
			updated.URI = strings.Replace(updated.URI, from, to, 1)
		}
	}

	if updated.URI != "" && !updated.IsPublic() && updated.PrivateKey == "" {
		updated.PrivateKey = defaults.RepoPrivateKey
	}

	// set the default for shallow clone if not defined in manifest
	if !original.ShallowDefined {
		updated.Shallow = defaults.ShallowClone
	}

	return updated
}
