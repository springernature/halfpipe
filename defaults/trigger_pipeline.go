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
		updated.ConcourseURL = defaults.ConcourseURL
	}

	if updated.Username == "" {
		updated.Username = defaults.ConcourseUsername
	}

	if updated.Password == "" {
		updated.Password = defaults.ConcoursePassword
	}

	if updated.Status == "" {
		updated.Status = "succeeded"
	}

	return updated
}
