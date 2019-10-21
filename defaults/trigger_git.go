package defaults

import (
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
	"strings"
)

func defaultGitTrigger(original manifest.GitTrigger, defaults Defaults) (updated manifest.GitTrigger) {
	updated = original

	updated.BasePath = defaults.Project.BasePath

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
