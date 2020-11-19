package defaults

import (
	"github.com/springernature/halfpipe/manifest"
)

func defaultPipelineTrigger(original manifest.PipelineTrigger, defaults Defaults, man manifest.Manifest) (updated manifest.PipelineTrigger) {
	updated = original

	if updated.Team == "" {
		updated.Team = man.Team
	}

	if updated.ConcourseURL == "" {
		updated.ConcourseURL = defaults.Concourse.URL
	}

	if updated.Username == "" {
		updated.Username = defaults.Concourse.Username
	}

	if updated.Password == "" {
		updated.Password = defaults.Concourse.Password
	}

	if updated.Status == "" {
		updated.Status = "succeeded"
	}

	return updated
}
